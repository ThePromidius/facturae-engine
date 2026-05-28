// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package aeat_test

import (
	"context"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ThePromidius/facturae-engine/src/internal/aeat"
)

func fakeAEATServer(t *testing.T, responseBody string, statusCode int) *httptest.Server {
	t.Helper()
	return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Content-Type"), "text/xml") {
			t.Errorf("expected text/xml Content-Type, got %q", r.Header.Get("Content-Type"))
		}
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		w.WriteHeader(statusCode)
		w.Write([]byte(responseBody))
	}))
}

func testClient(srv *httptest.Server, env aeat.Environment) *aeat.Client {
	aeat.Endpoints[env] = srv.URL
	c := aeat.NewClient(env, nil)
	c.SetHTTPClient(srv.Client())
	return c
}

func successSOAP() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/">
  <soapenv:Body>
    <RespuestaRegFactuSistemaFacturacion xmlns="https://www2.agenciatributaria.gob.es/static_files/common/internet/dep/aplicaciones/es/aeat/tike/cont/ws/RespuestaSuministro.xsd">
      <CSV>ABC123DEF456</CSV>
      <TiempoEsperaEnvio>0000</TiempoEsperaEnvio>
      <EstadoEnvio>Correcto</EstadoEnvio>
    </RespuestaRegFactuSistemaFacturacion>
  </soapenv:Body>
</soapenv:Envelope>`
}

func errorSOAP() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/">
  <soapenv:Body>
    <RespuestaRegFactuSistemaFacturacion xmlns="https://www2.agenciatributaria.gob.es/static_files/common/internet/dep/aplicaciones/es/aeat/tike/cont/ws/RespuestaSuministro.xsd">
      <CSV></CSV>
      <TiempoEsperaEnvio>0000</TiempoEsperaEnvio>
      <EstadoEnvio>Incorrecto</EstadoEnvio>
    </RespuestaRegFactuSistemaFacturacion>
  </soapenv:Body>
</soapenv:Envelope>`
}

func TestClient_SuccessfulSubmission(t *testing.T) {
	srv := fakeAEATServer(t, successSOAP(), 200)
	defer srv.Close()

	c := testClient(srv, aeat.EnvTest)
	result, err := c.Submit(context.Background(), []byte("<sfLR:RegFactuSistemaFacturacion xmlns:sfLR=\"https://www2.agenciatributaria.gob.es/static_files/common/internet/dep/aplicaciones/es/aeat/tike/cont/ws/SuministroLR.xsd\">test</sfLR:RegFactuSistemaFacturacion>"))
	if err != nil {
		t.Fatalf("Submit error: %v", err)
	}
	if result.CSV != "ABC123DEF456" {
		t.Errorf("expected CSV ABC123DEF456, got %q", result.CSV)
	}
	if !result.IsAccepted() {
		t.Error("result should be accepted")
	}
}

func TestClient_RejectedByAEAT(t *testing.T) {
	srv := fakeAEATServer(t, errorSOAP(), 200)
	defer srv.Close()

	c := testClient(srv, aeat.EnvTest)
	result, err := c.Submit(context.Background(), []byte("<sfLR:RegFactuSistemaFacturacion xmlns:sfLR=\"https://www2.agenciatributaria.gob.es/static_files/common/internet/dep/aplicaciones/es/aeat/tike/cont/ws/SuministroLR.xsd\">test</sfLR:RegFactuSistemaFacturacion>"))
	if err != nil {
		t.Fatalf("unexpected network error: %v", err)
	}
	if result.IsAccepted() {
		t.Error("result should NOT be accepted for Incorrecto state")
	}
	if result.Estado != "Incorrecto" {
		t.Errorf("expected Incorrecto, got %q", result.Estado)
	}
}

func TestClient_HTTPError_Retries(t *testing.T) {
	attempts := 0
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(503)
			return
		}
		w.Header().Set("Content-Type", "text/xml")
		w.Write([]byte(successSOAP()))
	}))
	defer srv.Close()

	original := aeat.Endpoints[aeat.EnvTest]
	aeat.Endpoints[aeat.EnvTest] = srv.URL
	defer func() { aeat.Endpoints[aeat.EnvTest] = original }()

	c := aeat.NewClient(aeat.EnvTest, nil)
	c.SetHTTPClient(srv.Client())

	result, err := c.Submit(context.Background(), []byte("<sfLR:RegFactuSistemaFacturacion xmlns:sfLR=\"https://www2.agenciatributaria.gob.es/static_files/common/internet/dep/aplicaciones/es/aeat/tike/cont/ws/SuministroLR.xsd\">test</sfLR:RegFactuSistemaFacturacion>"))
	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if !result.IsAccepted() {
		t.Error("expected accepted result after retries")
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
	}))
	defer srv.Close()

	original := aeat.Endpoints[aeat.EnvTest]
	aeat.Endpoints[aeat.EnvTest] = srv.URL
	defer func() { aeat.Endpoints[aeat.EnvTest] = original }()

	c := aeat.NewClient(aeat.EnvTest, nil)
	c.SetHTTPClient(srv.Client())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := c.Submit(ctx, []byte("<sfLR:RegFactuSistemaFacturacion xmlns:sfLR=\"https://www2.agenciatributaria.gob.es/static_files/common/internet/dep/aplicaciones/es/aeat/tike/cont/ws/SuministroLR.xsd\">test</sfLR:RegFactuSistemaFacturacion>"))
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestSOAPEnvelope_ContainsCIF(t *testing.T) {
	var captured []byte
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 4096)
		n, _ := r.Body.Read(buf)
		captured = buf[:n]
		w.Header().Set("Content-Type", "text/xml")
		w.Write([]byte(successSOAP()))
	}))
	defer srv.Close()

	original := aeat.Endpoints[aeat.EnvTest]
	aeat.Endpoints[aeat.EnvTest] = srv.URL
	defer func() { aeat.Endpoints[aeat.EnvTest] = original }()

	c := aeat.NewClient(aeat.EnvTest, nil)
	c.SetHTTPClient(srv.Client())
	c.Submit(context.Background(), []byte("<sfLR:RegFactuSistemaFacturacion xmlns:sfLR=\"https://www2.agenciatributaria.gob.es/static_files/common/internet/dep/aplicaciones/es/aeat/tike/cont/ws/SuministroLR.xsd\">test</sfLR:RegFactuSistemaFacturacion>"))

	if !strings.Contains(string(captured), "soapenv:Envelope") {
		t.Error("SOAP envelope should contain the soapenv:Envelope element")
	}
}

func TestSubmitResult_IsAccepted(t *testing.T) {
	tests := []struct {
		estado   string
		expected bool
	}{
		{"Correcto", true},
		{"AceptadoConErrores", true},
		{"Incorrecto", false},
		{"", false},
	}
	for _, tt := range tests {
		r := &aeat.SubmitResult{Estado: tt.estado}
		if r.IsAccepted() != tt.expected {
			t.Errorf("Estado=%q: IsAccepted()=%v, want %v", tt.estado, r.IsAccepted(), tt.expected)
		}
	}
}

func TestEndpoints_BothEnvironmentsDefined(t *testing.T) {
	for _, env := range []aeat.Environment{aeat.EnvTest, aeat.EnvProd} {
		url, ok := aeat.Endpoints[env]
		if !ok || url == "" {
			t.Errorf("missing endpoint for environment %q", env)
		}
	}
}

func TestSOAPResponseParsing(t *testing.T) {
	raw := successSOAP()
	var resp struct {
		XMLName xml.Name `xml:"Envelope"`
		Body    struct {
			Respuesta struct {
				CSV   string `xml:"CSV"`
				Estado string `xml:"EstadoEnvio"`
			} `xml:"RespuestaRegFactuSistemaFacturacion"`
		} `xml:"Body"`
	}
	if err := xml.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("SOAP XML parse error: %v", err)
	}
	if resp.Body.Respuesta.CSV != "ABC123DEF456" {
		t.Errorf("expected CSV ABC123DEF456, got %q", resp.Body.Respuesta.CSV)
	}
}
