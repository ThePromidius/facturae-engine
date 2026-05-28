#!/usr/bin/env bash
set -euo pipefail

# Test: Pipeline de Factura (Mock AEAT)

INVOICE="${1:-src/testdata/invoice_simple.json}"
SOCKET="${2:-127.0.0.1:9090}"
ENGINE_BIN="${3:-./facturae-engine}"

pass=0
fail=0

assert() {
    if eval "$1"; then
        pass=$((pass+1))
        echo "  [PASS] $2"
    else
        fail=$((fail+1))
        echo "  [FAIL] $2"
    fi
}

assert_status() {
    local code="$1" expected="$2" msg="$3"
    if [ "$code" = "$expected" ]; then
        pass=$((pass+1))
        echo "  [PASS] $msg"
    else
        fail=$((fail+1))
        echo "  [FAIL] $msg (esperado $expected, recibido $code)"
    fi
}

assert_header() {
    local headers="$1" name="$2" msg="$3"
    if echo "$headers" | grep -qi "^$name:"; then
        pass=$((pass+1))
        echo "  [PASS] $msg"
    else
        fail=$((fail+1))
        echo "  [FAIL] $msg (header '$name' ausente)"
    fi
}

cleanup() {
    if [ -n "${ENGINE_PID:-}" ]; then
        kill "$ENGINE_PID" 2>/dev/null || true
    fi
}
trap cleanup EXIT

echo "===== Test: Pipeline de Factura (Mock AEAT) ====="
echo ""

echo "[START] Arrancando servidor en $SOCKET..."
"$ENGINE_BIN" -socket "$SOCKET" &
ENGINE_PID=$!
sleep 2

base_url="http://$SOCKET"

# --- Health Check ---
echo ""
echo "--- Health Check ---"
health=$(curl -sf "$base_url/health" 2>/dev/null || echo "")
assert "echo '$health' | grep -q '\"status\":\"ok\"'" "Health status = ok"

# --- POST Invoice ---
echo ""
echo "--- POST /invoice ---"

body=$(cat "$INVOICE")
response=$(curl -s -i -X POST "$base_url/invoice" \
    -H "Content-Type: application/json" \
    -d "$body" 2>/dev/null) || true

status=$(echo "$response" | head -1 | awk '{print $2}')
assert_status "$status" "200" "Status 200"

echo "$response" | grep -qi "Content-Type:.*xml" && pass=$((pass+1)) && echo "  [PASS] Content-Type = application/xml" || { fail=$((fail+1)); echo "  [FAIL] Content-Type = application/xml"; }

assert "echo '$response' | grep -q '<fe:Facturae'" "Body contiene <fe:Facturae> root"
assert "echo '$response' | grep -qE '<ds:Signature|MOCK SIGNATURE'" "Body contiene firma"

assert_header "$response" "X-Verifactu-Fingerprint" "Header X-Verifactu-Fingerprint presente"
assert_header "$response" "X-Chain-Length" "Header X-Chain-Length presente"

fp=$(echo "$response" | grep -i "^X-Verifactu-Fingerprint:" | sed 's/.*: //' | tr -d '\r')
assert "[ ${#fp} -eq 64 ]" "Fingerprint = 64 caracteres hex"

chain_len=$(echo "$response" | grep -i "^X-Chain-Length:" | sed 's/.*: //' | tr -d '\r')
assert "[ $chain_len -ge 1 ]" "X-Chain-Length >= 1"

# --- Chain Endpoint ---
echo ""
echo "--- GET /chain ---"
chain=$(curl -sf "$base_url/chain" 2>/dev/null || echo "")
assert "echo '$chain' | grep -q '\"count\":' " "Chain tiene al menos 1 registro"

echo ""
echo "===== Resultados: $pass pasados, $fail fallidos ====="
if [ "$fail" -gt 0 ]; then exit 1; fi
