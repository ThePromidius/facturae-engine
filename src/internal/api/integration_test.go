// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package api_test

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ThePromidius/facturae-engine/src/internal/api"
	"github.com/ThePromidius/facturae-engine/src/internal/invoice"
)

func loadTestdata(t *testing.T, filename string) invoice.Request {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading testdata %q: %v", filename, err)
	}
	var req invoice.Request
	if err := json.Unmarshal(data, &req); err != nil {
		t.Fatalf("parsing testdata %q: %v", filename, err)
	}
	return req
}

func newIntegrationServer(t *testing.T) *httptest.Server {
	t.Helper()
	dir, _ := os.MkdirTemp("", "schema-int-*")
	t.Cleanup(func() { os.RemoveAll(dir) })
	srv, err := api.New(api.Config{SchemaDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	return httptest.NewServer(srv)
}

func post(t *testing.T, srv *httptest.Server, req invoice.Request) (int, []byte, http.Header) {
	t.Helper()
	body, _ := json.Marshal(req)
	resp, err := http.Post(srv.URL+"/invoice", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, b, resp.Header
}

func TestIntegration_SimpleInvoice(t *testing.T) {
	srv := newIntegrationServer(t)
	defer srv.Close()

	req := loadTestdata(t, "invoice_simple.json")
	code, body, _ := post(t, srv, req)

	if code != 200 {
		t.Fatalf("expected 200, got %d: %s", code, body)
	}

	xmlStr := string(body)
	checks := []struct{ name, substr string }{
		{"root element", "Facturae"},
		{"namespace", "facturae.gob.es"},
		{"seller CIF", "B12345678"},
		{"buyer CIF", "A87654321"},
		{"invoice number", "2024-0001"},
		{"issue date", "2024-05-12"},
		{"mock signature", "MOCK SIGNATURE"},
	}
	for _, c := range checks {
		if !strings.Contains(xmlStr, c.substr) {
			t.Errorf("[%s] XML should contain %q", c.name, c.substr)
		}
	}
}

func TestIntegration_SimpleInvoice_TotalsCorrect(t *testing.T) {
	srv := newIntegrationServer(t)
	defer srv.Close()

	req := loadTestdata(t, "invoice_simple.json")
	_, body, _ := post(t, srv, req)

	xmlStr := string(body)
	for _, total := range []string{"1500", "315", "1815"} {
		if !strings.Contains(xmlStr, total) {
			t.Errorf("XML should contain total value %s", total)
		}
	}
}

func TestIntegration_SimpleInvoice_IsValidXML(t *testing.T) {
	srv := newIntegrationServer(t)
	defer srv.Close()

	req := loadTestdata(t, "invoice_simple.json")
	_, body, _ := post(t, srv, req)

	xmlPart := strings.Split(string(body), "<!-- XAdES")[0]
	if err := xml.Unmarshal([]byte(xmlPart), new(interface{})); err != nil {
		t.Errorf("output XML is not valid: %v", err)
	}
}

func TestIntegration_MultiIVA(t *testing.T) {
	srv := newIntegrationServer(t)
	defer srv.Close()

	req := loadTestdata(t, "invoice_multi_iva.json")
	code, body, _ := post(t, srv, req)

	if code != 200 {
		t.Fatalf("expected 200, got %d: %s", code, body)
	}

	xmlStr := string(body)
	for _, rate := range []string{"<TaxRate>4</TaxRate>", "<TaxRate>10</TaxRate>",
		"<TaxRate>21</TaxRate>", "<TaxRate>0</TaxRate>"} {
		if !strings.Contains(xmlStr, rate) {
			t.Errorf("XML should contain %s", rate)
		}
	}
}

func TestIntegration_MultiIVA_TotalsCorrect(t *testing.T) {
	srv := newIntegrationServer(t)
	defer srv.Close()

	req := loadTestdata(t, "invoice_multi_iva.json")
	_, body, _ := post(t, srv, req)

	xmlStr := string(body)
	for _, v := range []string{"3910", "266.6", "4176.6"} {
		if !strings.Contains(xmlStr, v) {
			t.Errorf("XML should contain value %s", v)
		}
	}
}

func TestIntegration_ChainGrowsAcrossRequests(t *testing.T) {
	srv := newIntegrationServer(t)
	defer srv.Close()

	req1 := loadTestdata(t, "invoice_simple.json")
	req2 := loadTestdata(t, "invoice_multi_iva.json")

	_, _, h1 := post(t, srv, req1)
	_, _, h2 := post(t, srv, req2)

	fp1 := h1.Get("X-Verifactu-Fingerprint")
	fp2 := h2.Get("X-Verifactu-Fingerprint")

	if fp1 == "" || fp2 == "" {
		t.Fatal("fingerprints should not be empty")
	}
	if fp1 == fp2 {
		t.Error("consecutive invoices should have different fingerprints")
	}

	if h2.Get("X-Chain-Length") != "2" {
		t.Errorf("expected chain length 2 after 2 invoices, got %q", h2.Get("X-Chain-Length"))
	}
}

func TestIntegration_ChainEndpoint_AfterTwoInvoices(t *testing.T) {
	srv := newIntegrationServer(t)
	defer srv.Close()

	post(t, srv, loadTestdata(t, "invoice_simple.json"))
	post(t, srv, loadTestdata(t, "invoice_multi_iva.json"))

	resp, err := http.Get(srv.URL + "/chain")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)

	records, ok := data["records"].([]interface{})
	if !ok {
		t.Fatalf("expected records array in chain response")
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records in chain, got %v", data["count"])
	}

	last, ok := records[1].(map[string]interface{})
	if !ok {
		t.Fatalf("expected record to be a map")
	}
	if last["InvoiceNumber"] != "2024-0002" {
		t.Errorf("expected last chain record to be 2024-0002, got %v", last["InvoiceNumber"])
	}
	if last["PreviousFingerprint"] == "" {
		t.Error("second record should have a non-empty PreviousFingerprint")
	}
}

func TestIntegration_ChainVerify_AfterTwoInvoices(t *testing.T) {
	srv := newIntegrationServer(t)
	defer srv.Close()

	post(t, srv, loadTestdata(t, "invoice_simple.json"))
	post(t, srv, loadTestdata(t, "invoice_multi_iva.json"))

	resp, err := http.Get(srv.URL + "/chain/verify")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)

	if data["status"] != "ok" {
		t.Fatalf("expected chain status 'ok', got %q: %v", data["status"], data["error"])
	}
	if data["chain_length"] != 2.0 {
		t.Errorf("expected chain_length 2, got %v", data["chain_length"])
	}
}

func TestIntegration_ChainVerify_EmptyChain(t *testing.T) {
	srv := newIntegrationServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/chain/verify")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)

	if data["status"] != "ok" {
		t.Errorf("expected chain status 'ok' for empty chain, got %q", data["status"])
	}
}
