package api_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ThePromidius/facturae-engine/src/internal/api"
	"github.com/ThePromidius/facturae-engine/src/internal/chain"
)

func TestServer_HandleInvoice_NativeXML(t *testing.T) {
	// 1. Setup Server
	store := chain.NewMemoryStore()
	srv, _ := api.New(api.Config{
		ChainStore: store,
		SchemaDir:  t.TempDir(),
	})

	// 2. Prepare Native FacturaE XML
	nativeXML := `<?xml version="1.0" encoding="UTF-8"?>
<fe:Facturae xmlns:fe="http://www.facturae.gob.es/formato/Versiones/Facturaev3_2_2.xml">
	<TaxIdentificationNumber>B12345678</TaxIdentificationNumber>
	<InvoiceNumber>001</InvoiceNumber>
	<IssueDate>2026-05-14</IssueDate>
	<InvoiceTotal>121.00</InvoiceTotal>
</fe:Facturae>`

	req := httptest.NewRequest(http.MethodPost, "/invoice", bytes.NewReader([]byte(nativeXML)))
	req.Header.Set("Content-Type", "application/xml")
	w := httptest.NewRecorder()

	// 3. Process
	srv.ServeHTTP(w, req)

	// 4. Verify
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Ensure Verifactu tags were appended
	if !bytes.Contains(w.Body.Bytes(), []byte("<Huella>")) {
		t.Error("Response should contain Verifactu <Huella> tag")
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("XAdES-BES MOCK SIGNATURE")) {
		t.Error("Response should contain mock signature comment")
	}
}
