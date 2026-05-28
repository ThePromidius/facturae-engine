#!/usr/bin/env bash
set -euo pipefail

# Test: Cadena Verifactu (Hash Chain)

INVOICE1="${1:-src/testdata/invoice_simple.json}"
INVOICE2="${2:-src/testdata/invoice_multi_iva.json}"
SOCKET="${3:-127.0.0.1:9091}"
ENGINE_BIN="${4:-./facturae-engine}"

pass=0; fail=0

assert() { if eval "$1"; then pass=$((pass+1)); echo "  [PASS] $2"; else fail=$((fail+1)); echo "  [FAIL] $2"; fi; }
assert_status() { if [ "$1" = "$2" ]; then pass=$((pass+1)); echo "  [PASS] $3"; else fail=$((fail+1)); echo "  [FAIL] $3 (esperado $2, recibido $1)"; fi; }

cleanup() { [ -n "${ENGINE_PID:-}" ] && kill "$ENGINE_PID" 2>/dev/null || true; }
trap cleanup EXIT

echo "===== Test: Cadena Verifactu (Hash Chain) ====="
echo ""

echo "[START] Arrancando servidor en $SOCKET..."
"$ENGINE_BIN" -socket "$SOCKET" &
ENGINE_PID=$!
sleep 2

base="http://$SOCKET"

# --- Factura 1 ---
echo ""
echo "--- Factura 1 ---"
r1=$(curl -s -i -X POST "$base/invoice" -H "Content-Type: application/json" -d @"$INVOICE1" 2>/dev/null) || true
s1=$(echo "$r1" | head -1 | awk '{print $2}')
assert_status "$s1" "200" "Status 200"

fp1=$(echo "$r1" | grep -i "^X-Verifactu-Fingerprint:" | sed 's/.*: //' | tr -d '\r')
len1=$(echo "$r1" | grep -i "^X-Chain-Length:" | sed 's/.*: //' | tr -d '\r')
assert "[ $len1 -eq 1 ]" "X-Chain-Length = 1"
assert "[ ${#fp1} -eq 64 ]" "Fingerprint 1 = 64 hex"

# --- Factura 2 ---
echo ""
echo "--- Factura 2 ---"
r2=$(curl -s -i -X POST "$base/invoice" -H "Content-Type: application/json" -d @"$INVOICE2" 2>/dev/null) || true
s2=$(echo "$r2" | head -1 | awk '{print $2}')
assert_status "$s2" "200" "Status 200"

fp2=$(echo "$r2" | grep -i "^X-Verifactu-Fingerprint:" | sed 's/.*: //' | tr -d '\r')
len2=$(echo "$r2" | grep -i "^X-Chain-Length:" | sed 's/.*: //' | tr -d '\r')
assert "[ $len2 -eq 2 ]" "X-Chain-Length = 2"
assert "[ ${#fp2} -eq 64 ]" "Fingerprint 2 = 64 hex"
assert "[ '$fp1' != '$fp2' ]" "Fingerprint 1 != 2 (distintos)"

# --- GET /chain ---
echo ""
echo "--- GET /chain ---"
chain=$(curl -sf "$base/chain" 2>/dev/null || echo "")
assert "echo '$chain' | grep -q '\"count\":2'" "Chain.count = 2"

first_prev=$(echo "$chain" | grep -o '"PreviousFingerprint":"[^"]*"' | head -1)
assert "echo '$first_prev' | grep -q '\"PreviousFingerprint\":\"\"'" "Primer registro sin PreviousFingerprint"

echo ""
echo "===== Resultados: $pass pasados, $fail fallidos ====="
if [ "$fail" -gt 0 ]; then exit 1; fi
