#!/usr/bin/env bash
set -euo pipefail

# Test: Engine gracefully handles SIGTERM and persists chain data across restarts.
# Usage: test-graceful-shutdown.sh INVOICE INVOICE2 SOCKET ENGINE
INVOICE="${1:?missing INVOICE}"
INVOICE2="${2:?missing INVOICE2}"
SOCKET="${3:?missing SOCKET}"
ENGINE="${4:?missing ENGINE}"
USE_SQL="${5:-}"

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DATA_DIR="${ROOT}/.tmp/graceful-$$"
cleanup() { rm -rf "$DATA_DIR"; }
trap cleanup EXIT

mkdir -p "$DATA_DIR"

pass=0
fail=0

check() {
    local name="$1" got="$2" want="$3"
    if [ "$got" = "$want" ]; then
        echo "  PASS: $name"
        pass=$((pass+1))
    else
        echo "  FAIL: $name (got: $got, want: $want)"
        fail=$((fail+1))
    fi
}

RESP_FILE="$DATA_DIR/resp.json"

run_engine() {
    local socket="$1" engine_flags=""
    [ -n "${USE_SQL}" ] && engine_flags="-db sqlite -dsn $DATA_DIR/chain.db"
    $ENGINE -socket "$socket" $engine_flags &
    local pid=$!
    echo "$pid" > "$DATA_DIR/engine.pid"
    # Wait for engine to be ready
    for i in $(seq 1 20); do
        if curl -sf "http://$socket/health" > /dev/null 2>&1; then
            return 0
        fi
        sleep 0.3
    done
    return 1
}

stop_engine() {
    local pid
    pid=$(cat "$DATA_DIR/engine.pid" 2>/dev/null || true)
    [ -z "$pid" ] && return 0
    echo "  -> Stopping engine (PID $pid)..."
    kill -TERM "$pid" 2>/dev/null || true
    # Wait up to 5s for graceful shutdown
    for i in $(seq 1 10); do
        if ! kill -0 "$pid" 2>/dev/null; then
            echo "  -> Engine stopped gracefully"
            return 0
        fi
        sleep 0.5
    done
    echo "  -> Force killing engine..."
    kill -KILL "$pid" 2>/dev/null || true
}

send_invoice() {
    local invoice="$1" socket="$2"
    curl -sf -X POST "http://$socket/invoice" \
        -H "Content-Type: application/json" \
        -d @"$invoice" -o "$DATA_DIR/last_response.xml" 2>/dev/null
    local ec=$?
    return $ec
}

get_chain_length() {
    local socket="$1"
    curl -sf "http://$socket/chain" 2>/dev/null | jq -r '.count // 0' 2>/dev/null || echo "0"
}

echo "========================================="
echo "  TEST: Graceful Shutdown & Persistence"
echo "========================================="

echo ""
echo "--- Phase 1: Start engine, send invoice ---"
if ! run_engine "$SOCKET"; then
    echo "  FAIL: Engine did not start"
    exit 1
fi
echo "  Engine started on $SOCKET"

if ! send_invoice "$INVOICE" "$SOCKET"; then
    echo "  FAIL: First invoice rejected"
    stop_engine
    exit 1
fi
chain1=$(get_chain_length "$SOCKET")
check "Chain length after invoice 1" "$chain1" "1"

echo ""
echo "--- Phase 2: Graceful shutdown ---"
stop_engine

echo ""
echo "--- Phase 3: Restart engine, verify data persisted ---"
rm -f "$SOCKET"
if ! run_engine "$SOCKET"; then
    echo "  FAIL: Engine did not restart"
    exit 1
fi
echo "  Engine restarted on $SOCKET"

chain2=$(get_chain_length "$SOCKET")
check "Chain length after restart" "$chain2" "1"

echo ""
echo "--- Phase 4: Send second invoice, verify chain grows ---"
if ! send_invoice "$INVOICE2" "$SOCKET"; then
    echo "  FAIL: Second invoice rejected"
    stop_engine
    exit 1
fi
chain3=$(get_chain_length "$SOCKET")
check "Chain length after invoice 2" "$chain3" "2"

stop_engine

echo ""
echo "========================================="
echo "  Graceful shutdown: $pass pasados, $fail fallidos"
echo "========================================="
[ "$fail" -eq 0 ]
