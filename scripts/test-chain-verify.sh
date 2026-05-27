#!/usr/bin/env bash
set -euo pipefail

# Test: Verificacion de Cadena Verifactu (Bunker)

INVOICE1="${1:-src/testdata/invoice_simple.json}"
INVOICE2="${2:-src/testdata/invoice_multi_iva.json}"
SOCKET="${3:-127.0.0.1:9093}"
ENGINE_BIN="${4:-./facturae-engine}"
USE_SQL="${5:-}"

pass=0; fail=0

assert() { if eval "$1"; then pass=$((pass+1)); echo "  [PASS] $2"; else fail=$((fail+1)); echo "  [FAIL] $2"; fi; }
assert_status() { if [ "$1" = "$2" ]; then pass=$((pass+1)); echo "  [PASS] $3"; else fail=$((fail+1)); echo "  [FAIL] $3 (esperado $2, recibido $1)"; fi; }

cleanup() {
    [ -n "${ENGINE_PID:-}" ] && kill "$ENGINE_PID" 2>/dev/null || true
    [ -n "${ENGINE2_PID:-}" ] && kill "$ENGINE2_PID" 2>/dev/null || true
}
trap cleanup EXIT

echo "===== Test: Verificacion de Cadena Verifactu (Bunker) ====="
echo ""

sql_flag=""
sql_label=""
db_path="/tmp/chain-verify.db"
rm -f "$db_path"

if [ -n "$USE_SQL" ]; then
    sql_flag="-db sqlite -dsn $db_path"
    sql_label="(SQLite)"
    echo "[SQL] Usando SQLite persistente"
fi

echo "[START] Arrancando servidor en $SOCKET $sql_label..."
"$ENGINE_BIN" -socket "$SOCKET" $sql_flag &
ENGINE_PID=$!
sleep 3

base="http://$SOCKET"

# --- Prueba 1: Cadena vacia ---
echo ""
echo "--- Prueba 1: Cadena vacia ---"
v1=$(curl -sf "$base/chain/verify" 2>/dev/null || echo "")
assert "echo '$v1' | grep -q '\"status\":\"ok\"'" "status = ok"

# --- Prueba 2: Enviar 2 facturas ---
echo ""
echo "--- Prueba 2: Enviar 2 facturas ---"
s1=$(curl -s -o /dev/null -w '%{http_code}' -X POST "$base/invoice" -H "Content-Type: application/json" -d @"$INVOICE1" 2>/dev/null)
assert_status "$s1" "200" "Factura 1: Status 200"

s2=$(curl -s -o /dev/null -w '%{http_code}' -X POST "$base/invoice" -H "Content-Type: application/json" -d @"$INVOICE2" 2>/dev/null)
assert_status "$s2" "200" "Factura 2: Status 200"

# --- Prueba 3: Verify ---
echo ""
echo "--- Prueba 3: Verificar (2 facturas) ---"
v2=$(curl -sf "$base/chain/verify" 2>/dev/null || echo "")
assert "echo '$v2' | grep -q '\"status\":\"ok\"'" "status = ok"
assert "echo '$v2' | grep -q '\"chain_length\":2'" "chain_length = 2"

# --- Prueba 4: Fingerprints ---
echo ""
echo "--- Prueba 4: Integridad fingerprints ---"
chain=$(curl -sf "$base/chain" 2>/dev/null || echo "")
assert "echo '$chain' | grep -q '\"count\":2'" "Chain tiene 2 registros"

# --- Prueba 5: Persistencia SQL ---
if [ -n "$USE_SQL" ]; then
    echo ""
    echo "--- Prueba 5: Persistencia SQL (re-arranque) ---"
    echo "  [INFO] Re-arrancando con misma DB..."
    kill "$ENGINE_PID" 2>/dev/null; sleep 1
    rm -f /tmp/chain-verify-err.log

    "$ENGINE_BIN" -socket "$SOCKET" $sql_flag &
    ENGINE2_PID=$!
    sleep 3

    v3=$(curl -sf "$base/chain/verify" 2>/dev/null || echo "")
    assert "echo '$v3' | grep -q '\"status\":\"ok\"'" "Re-arranque: cadena intacta"
    assert "echo '$v3' | grep -q '\"chain_length\":2'" "Re-arranque: chain_length = 2"

    kill "$ENGINE2_PID" 2>/dev/null || true
    rm -f "$db_path"
fi

echo ""
echo "===== Resultados: $pass pasados, $fail fallidos ====="
if [ "$fail" -gt 0 ]; then exit 1; fi
