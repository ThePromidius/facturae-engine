#!/usr/bin/env bash
set -euo pipefail

# Run all integration test scripts sequentially

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ENGINE="${ENGINE_BIN:-$ROOT/facturae-engine}"
INVOICE1="${INVOICE1:-$ROOT/src/testdata/invoice_simple.json}"
INVOICE2="${INVOICE2:-$ROOT/src/testdata/invoice_multi_iva.json}"
LABEL="${LABEL:-local}"

total_pass=0
total_fail=0

run_test() {
    local name="$1" script="$2" port="$3" extra="${4:-}"
    echo ""
    echo "========================================="
    echo "  [$LABEL] $name (port $port)"
    echo "========================================="

    set +e
    if [ -n "$extra" ]; then
        bash "$script" "$INVOICE1" "$INVOICE2" "127.0.0.1:$port" "$ENGINE" "$extra"
    else
        bash "$script" "$INVOICE1" "$INVOICE2" "127.0.0.1:$port" "$ENGINE"
    fi
    local ec=$?
    set -e

    if [ $ec -eq 0 ]; then
        total_pass=$((total_pass+1))
        echo "  [SUITE] $name PASSED"
    else
        total_fail=$((total_fail+1))
        echo "  [SUITE] $name FAILED (exit $ec)"
    fi
}

# test-invoice expects: INVOICE SOCKET ENGINE (3 args)
echo ""
echo "========================================="
echo "  [$LABEL] test-invoice (port 9090)"
echo "========================================="
set +e
bash "$ROOT/scripts/test-invoice.sh" "$INVOICE1" "127.0.0.1:9090" "$ENGINE"
ec=$?; set -e
if [ $ec -eq 0 ]; then total_pass=$((total_pass+1)); echo "  [SUITE] test-invoice PASSED"
else total_fail=$((total_fail+1)); echo "  [SUITE] test-invoice FAILED (exit $ec)"; fi

run_test "test-chain"               "$ROOT/scripts/test-chain.sh"         9091
run_test "test-chain-verify"        "$ROOT/scripts/test-chain-verify.sh"  9093
run_test "test-graceful-shutdown"   "$ROOT/scripts/test-graceful-shutdown.sh" 9094 "sqlite"

echo ""
echo "========================================="
echo "  Suite completo: $total_pass pasados, $total_fail fallidos"
echo "========================================="
[ "$total_fail" -eq 0 ]
