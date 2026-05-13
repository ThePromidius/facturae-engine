package aeat_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ThePromidius/facturae-engine/src/internal/aeat"
)

func TestClient_Submit_Mock(t *testing.T) {
	// 1. Setup Mock AEAT Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Headers
		if r.Header.Get("SOAPAction") == "" {
			t.Error("missing SOAPAction header")
		}

		// Mock Successful Response
		resp := `<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/">
  <soapenv:Body>
    <RespuestaRegFactuSistemaFacturacion xmlns="http://www.agenciatributaria.gob.es/AEAT/VERIFACTU/SistemaFacturacion.xsd">
      <CSV>ABC-123-XYZ</CSV>
      <EstadoEnvio>Correcto</EstadoEnvio>
      <DescripcionEstadoEnvio>Prueba OK</DescripcionEstadoEnvio>
    </RespuestaRegFactuSistemaFacturacion>
  </soapenv:Body>
</soapenv:Envelope>`
		w.Header().Set("Content-Type", "text/xml")
		fmt.Fprint(w, resp)
	}))
	defer server.Close()

	// 2. Setup Client to point to mock
	client := aeat.NewClient(aeat.EnvTest, nil)
	client.SetHTTPClient(server.Client())
	
	// Backup and Restore endpoint
	original := aeat.Endpoints[aeat.EnvTest]
	aeat.Endpoints[aeat.EnvTest] = server.URL
	defer func() { aeat.Endpoints[aeat.EnvTest] = original }()

	// 3. Submit
	signedXML := []byte(`<sum:RegFactuSistemaFacturacion xmlns:sum="http://www.agenciatributaria.gob.es/AEAT/VERIFACTU/SistemaFacturacion.xsd">test</sum:RegFactuSistemaFacturacion>`)
	res, err := client.Submit(context.Background(), signedXML, "B12345678")
	
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	if res.Estado != "Correcto" {
		t.Errorf("Expected status Correcto, got %s", res.Estado)
	}

	if res.CSV != "ABC-123-XYZ" {
		t.Errorf("Expected CSV ABC-123-XYZ, got %s", res.CSV)
	}
}
