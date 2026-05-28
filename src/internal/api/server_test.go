// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ThePromidius/facturae-engine/src/internal/api"
	"github.com/ThePromidius/facturae-engine/src/internal/invoice"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	dir, err := os.MkdirTemp("", "schema-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	srv, err := api.New(api.Config{
		SocketPath: "",
		SchemaDir:  dir,
		Signer:     nil,
	})
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}
	return httptest.NewServer(srv)
}

func validBody() invoice.Request {
	return invoice.Request{
		Meta: invoice.Meta{Version: "3.2.2", Moneda: "EUR"},
		Factura: invoice.Factura{
			Numero: "2024-TEST-001",
			Serie:  "T",
			Fecha:  time.Date(2024, 5, 12, 0, 0, 0, 0, time.UTC),
		},
		Emisor: invoice.Party{
			CIF: "B12345678", Nombre: "Test Empresa S.L.",
			Direccion: "Calle Test 1", CP: "28001",
			Ciudad: "Madrid", Provincia: "Madrid", Pais: "ESP",
		},
		Receptor: invoice.Party{
			CIF: "A87654321", Nombre: "Cliente Test S.A.",
			Direccion: "Av. Cliente 456", CP: "08001",
			Ciudad: "Barcelona", Provincia: "Barcelona", Pais: "ESP",
		},
		Lineas: []invoice.Linea{
			{Descripcion: "Servicio A", Cantidad: 5, PrecioUnitario: 100, IVATipo: 21},
			{Descripcion: "Servicio B", Cantidad: 2, PrecioUnitario: 250, IVATipo: 10},
		},
	}
}

func postInvoice(t *testing.T, srv *httptest.Server, req invoice.Request) *http.Response {
	t.Helper()
	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	resp, err := http.Post(srv.URL+"/invoice", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /invoice: %v", err)
	}
	return resp
}

func TestInvoice_HappyPath_Returns200(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp := postInvoice(t, srv, validBody())
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, body)
	}
}

func TestInvoice_ResponseIsXML(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp := postInvoice(t, srv, validBody())
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/xml") {
		t.Errorf("expected Content-Type application/xml, got %q", ct)
	}
}

func TestInvoice_ResponseContainsFacturaeRoot(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp := postInvoice(t, srv, validBody())
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Facturae") {
		t.Error("response XML should contain root element Facturae")
	}
}

func TestInvoice_ResponseContainsMockSignature(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp := postInvoice(t, srv, validBody())
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "MOCK SIGNATURE") {
		t.Error("response should contain mock signature comment")
	}
}

func TestInvoice_ResponseHasVerifactuHeader(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp := postInvoice(t, srv, validBody())
	defer resp.Body.Close()

	fp := resp.Header.Get("X-Verifactu-Fingerprint")
	if len(fp) != 64 {
		t.Errorf("X-Verifactu-Fingerprint should be 64 hex chars, got %q (len %d)", fp, len(fp))
	}
}

func TestInvoice_ChainLengthIncrements(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	for i := 1; i <= 3; i++ {
		req := validBody()
		req.Factura.Numero = strings.Repeat("0", 3-len(string(rune('0'+i)))) + string(rune('0' + i))
		resp := postInvoice(t, srv, req)
		resp.Body.Close()

		chainLen := resp.Header.Get("X-Chain-Length")
		expected := string(rune('0' + i))
		if chainLen != expected {
			t.Errorf("invoice %d: expected X-Chain-Length=%s, got %s", i, expected, chainLen)
		}
	}
}

func TestInvoice_InvalidJSON_Returns400(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/invoice", "application/json",
		strings.NewReader("{not valid json"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestInvoice_MissingEmisor_Returns400(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	req := validBody()
	req.Emisor.CIF = ""

	resp := postInvoice(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing emisor CIF, got %d", resp.StatusCode)
	}
}

func TestInvoice_EmptyLineas_Returns400(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	req := validBody()
	req.Lineas = nil

	resp := postInvoice(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for empty lineas, got %d", resp.StatusCode)
	}
}

func TestInvoice_WrongMethod_Returns405(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/invoice")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", resp.StatusCode)
	}
}

func TestInvoice_MultipleVATBrackets_XML(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	req := validBody()
	req.Lineas = []invoice.Linea{
		{Descripcion: "IVA 21%", Cantidad: 1, PrecioUnitario: 100, IVATipo: 21},
		{Descripcion: "IVA 10%", Cantidad: 1, PrecioUnitario: 200, IVATipo: 10},
		{Descripcion: "IVA 0%", Cantidad: 1, PrecioUnitario: 300, IVATipo: 0},
	}

	resp := postInvoice(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, body)
	}

	body, _ := io.ReadAll(resp.Body)
	for _, rate := range []string{"21", "10", "0"} {
		if !strings.Contains(string(body), "<TaxRate>"+rate+"</TaxRate>") {
			t.Errorf("XML should contain TaxRate %s", rate)
		}
	}
}

func TestHealth_Returns200(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHealth_ResponseJSON(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("health response is not valid JSON: %v", err)
	}
	if data["status"] != "ok" {
		t.Errorf("expected status ok, got %v", data["status"])
	}
}

func TestChain_EmptyOnStart(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/chain")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)
	count, ok := data["count"].(float64)
	if !ok || count != 0 {
		t.Errorf("expected empty chain (count=0) on startup, got %v", data)
	}
}

func TestChain_AfterInvoice_HasRecord(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	postInvoice(t, srv, validBody())

	resp, err := http.Get(srv.URL + "/chain")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)

	records, ok := data["records"].([]interface{})
	if !ok || len(records) == 0 {
		t.Fatalf("expected records in chain, got %v", data)
	}

	r, ok := records[0].(map[string]interface{})
	if !ok || r["Fingerprint"] == "" {
		t.Error("chain record should contain a Fingerprint after processing an invoice")
	}
}
