// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package api_test

import (
	"bytes"
	"encoding/json"
	"image/png"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ThePromidius/facturae-engine/src/internal/api"
	"github.com/ThePromidius/facturae-engine/src/internal/invoice"
)

func newQRTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	dir, _ := os.MkdirTemp("", "qr-schema-*")
	t.Cleanup(func() { os.RemoveAll(dir) })
	srv, err := api.New(api.Config{SchemaDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	return httptest.NewServer(srv)
}

func validInvoiceBody(numero string) invoice.Request {
	return invoice.Request{
		Meta: invoice.Meta{Version: "3.2.2", Moneda: "EUR"},
		Factura: invoice.Factura{
			Numero: numero, Serie: "T",
			Fecha: time.Date(2024, 5, 12, 0, 0, 0, 0, time.UTC),
		},
		Emisor: invoice.Party{
			CIF: "B12345678", Nombre: "Test Empresa S.L.",
			Direccion: "Calle Test 1", CP: "28001",
			Ciudad: "Madrid", Provincia: "Madrid", Pais: "ESP",
		},
		Receptor: invoice.Party{CIF: "A87654321", Nombre: "Cliente Test S.A.", Direccion: "Av. Cliente 456", CP: "08001", Ciudad: "Barcelona", Provincia: "Barcelona", Pais: "ESP"},
		Lineas: []invoice.Linea{
			{Descripcion: "Servicio", Cantidad: 1, PrecioUnitario: 100, IVATipo: 21},
		},
	}
}

func TestQREndpoint_ValidText(t *testing.T) {
	srv := newQRTestServer(t)
	defer srv.Close()

	text := url.QueryEscape("https://agenciatributaria.gob.es/test")
	resp, err := http.Get(srv.URL + "/qr?text=" + text)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(resp.Header.Get("Content-Type"), "image/png") {
		t.Errorf("expected image/png Content-Type, got %q", resp.Header.Get("Content-Type"))
	}
}

func TestQREndpoint_ReturnsPNG(t *testing.T) {
	srv := newQRTestServer(t)
	defer srv.Close()

	text := url.QueryEscape("https://example.com/qr-test")
	resp, err := http.Get(srv.URL + "/qr?text=" + text)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	_, err = png.Decode(resp.Body)
	if err != nil {
		t.Errorf("response is not a valid PNG: %v", err)
	}
}

func TestQREndpoint_MissingText_Returns400(t *testing.T) {
	srv := newQRTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/qr")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 400 {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestInvoice_ResponseHasQRURL(t *testing.T) {
	srv := newQRTestServer(t)
	defer srv.Close()

	body, _ := json.Marshal(validInvoiceBody("QR-001"))
	resp, err := http.Post(srv.URL+"/invoice", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	qrURL := resp.Header.Get("X-Verifactu-QR-URL")
	if qrURL == "" {
		t.Error("expected X-Verifactu-QR-URL header in invoice response")
	}
	if !strings.Contains(qrURL, "agenciatributaria") {
		t.Errorf("QR URL should point to AEAT, got: %s", qrURL)
	}
	if !strings.Contains(qrURL, "B12345678") {
		t.Errorf("QR URL should contain emisor CIF, got: %s", qrURL)
	}
	if !strings.Contains(qrURL, "numserie=TQR-001") {
		t.Errorf("QR URL should contain numserie, got: %s", qrURL)
	}
	if !strings.Contains(qrURL, "importe=121.00") {
		t.Errorf("QR URL should contain importe, got: %s", qrURL)
	}
}

func TestInvoice_QRDataURIHeader(t *testing.T) {
	srv := newQRTestServer(t)
	defer srv.Close()

	body, _ := json.Marshal(validInvoiceBody("QR-003"))
	resp, err := http.Post(srv.URL+"/invoice", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	dataURI := resp.Header.Get("X-Verifactu-QR-DataURI")
	if !strings.HasPrefix(dataURI, "data:image/png;base64,") {
		t.Errorf("expected data URI header, got: %s", dataURI)
	}
}
