package aeat_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ThePromidius/facturae-engine/src/internal/aeat"
	"github.com/ThePromidius/facturae-engine/src/internal/chain"
	"github.com/ThePromidius/facturae-engine/src/internal/facturae"
	"github.com/ThePromidius/facturae-engine/src/internal/invoice"
	"github.com/ThePromidius/facturae-engine/src/internal/signing"
)

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func computeFingerprint(r chain.Record) string {
	prevFP := r.PreviousFingerprint
	if len(prevFP) > 64 {
		prevFP = prevFP[:64]
	}
	canonical := fmt.Sprintf("%s|%s-%s|%s|%s|%.2f|%.2f|%s|%s",
		r.EmisorCIF,
		r.InvoiceSeries, r.InvoiceNumber,
		r.IssueDate.Format("02-01-2006"),
		r.InvoiceType,
		r.TaxAmount,
		r.Total,
		prevFP,
		r.Timestamp.In(time.UTC).Format("2006-01-02T15:04:05Z"),
	)
	h := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(h[:])
}

func loadRequest(t *testing.T, path string) invoice.Request {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	var req invoice.Request
	if err := json.Unmarshal(data, &req); err != nil {
		t.Fatalf("failed to unmarshal %s: %v", path, err)
	}
	return req
}

func mapInvoiceType(docType string) string {
	switch docType {
	case "FC":
		return "F1"
	case "FA":
		return "F2"
	case "AF":
		return "R1"
	default:
		return "F1"
	}
}

func TestFullPipeline(t *testing.T) {
	t.Run("simple_invoice", func(t *testing.T) {
		runPipeline(t, "../../testdata/invoice_simple.json", false)
	})
	t.Run("multi_iva_invoice", func(t *testing.T) {
		runPipeline(t, "../../testdata/invoice_multi_iva.json", false)
	})
	t.Run("rectificativa", func(t *testing.T) {
		runPipeline(t, "../../testdata/invoice_simple.json", true)
	})
}

func runPipeline(t *testing.T, jsonPath string, rectificativa bool) {
	t.Helper()

	// ---- 1. Load JSON ----
	req := loadRequest(t, jsonPath)

	// ---- 2. Build FacturaE ----
	f, err := facturae.Build(req)
	if err != nil {
		t.Fatalf("facturae.Build failed: %v", err)
	}
	if rectificativa {
		f.Invoices.Invoice[0].InvoiceHeader.InvoiceDocumentType = "AF"
	}

	// ---- 3. Marshal to XML ----
	facturaeXML, err := xml.MarshalIndent(f, "", "  ")
	if err != nil {
		t.Fatalf("xml.MarshalIndent failed: %v", err)
	}
	facturaeXML = append([]byte(xml.Header), facturaeXML...)

	// ---- 4. Sign using MockSigner ----
	signer := signing.MockSigner{}
	signedXML, err := signer.Sign(facturaeXML)
	if err != nil {
		t.Fatalf("MockSigner.Sign failed: %v", err)
	}

	// ---- 5. Extract signature ----
	sigXML, sigErr := aeat.ExtractSignature(signedXML)
	if sigErr != nil {
		sigXML = ""
	} else {
		if !strings.Contains(sigXML, "<ds:Signature") {
			t.Error("extracted signature missing ds:Signature element")
		}
	}

	// ---- 6. Build chain Record manually ----
	inv := f.Invoices.Invoice[0]
	docType := inv.InvoiceHeader.InvoiceDocumentType
	mappedType := mapInvoiceType(docType)

	issueDate, err := time.Parse("2006-01-02", inv.InvoiceIssueData.IssueDate)
	if err != nil {
		t.Fatalf("failed to parse issue date %q: %v", inv.InvoiceIssueData.IssueDate, err)
	}

	taxAmt := round2(inv.InvoiceTotals.TotalTaxOutputs)
	totalAmt := round2(inv.InvoiceTotals.InvoiceTotal)

	now := time.Now().UTC()
	rec := chain.Record{
		InvoiceNumber:       inv.InvoiceHeader.InvoiceNumber,
		InvoiceSeries:       inv.InvoiceHeader.InvoiceSeriesCode,
		EmisorCIF:           req.Emisor.CIF,
		IssueDate:           issueDate,
		InvoiceType:         mappedType,
		TaxAmount:           taxAmt,
		Total:               totalAmt,
		PreviousFingerprint: "",
		Fingerprint:         "",
		Timestamp:           now,
	}
	rec.Fingerprint = computeFingerprint(rec)

	// ---- 7. Build SuministroLR ----
	config := aeat.DefaultConfig()
	suministro := aeat.BuildSuministroLR(
		f, inv, rec,
		req.Emisor.CIF, req.Emisor.Nombre,
		sigXML, config,
	)

	// ---- 8. Marshal SuministroLR to XML ----
	sumXML, err := xml.MarshalIndent(suministro, "", "  ")
	if err != nil {
		t.Fatalf("xml.MarshalIndent SuministroLR failed: %v", err)
	}
	sumXML = append([]byte(xml.Header), sumXML...)
	xmlStr := string(sumXML)

	// ---- 9. Verify ALL structural elements ----
	t.Run("verify_xml_structure", func(t *testing.T) {
		if !strings.Contains(xmlStr, "<sfLR:RegFactuSistemaFacturacion") {
			t.Error("missing root element sfLR:RegFactuSistemaFacturacion")
		}
		if !strings.Contains(xmlStr, `xmlns:sfLR="`+aeat.NSLR+`"`) {
			t.Error("missing xmlns:sfLR namespace")
		}
		if !strings.Contains(xmlStr, `xmlns:sf="`+aeat.NSInfo+`"`) {
			t.Error("missing xmlns:sf namespace")
		}
		if !strings.Contains(xmlStr, "<sf:ObligadoEmision>") {
			t.Error("missing sf:ObligadoEmision")
		}
		if !strings.Contains(xmlStr, "<sf:NombreRazon>"+req.Emisor.Nombre+"</sf:NombreRazon>") {
			t.Error("missing or wrong sf:NombreRazon in ObligadoEmision")
		}
		if !strings.Contains(xmlStr, "<sf:NIF>"+req.Emisor.CIF+"</sf:NIF>") {
			t.Error("missing or wrong sf:NIF in ObligadoEmision")
		}
		if !strings.Contains(xmlStr, "<sf:IDVersion>1.0</sf:IDVersion>") {
			t.Error("missing sf:IDVersion 1.0")
		}
		if !strings.Contains(xmlStr, "<sf:TipoFactura>"+mappedType+"</sf:TipoFactura>") {
			t.Errorf("missing or wrong sf:TipoFactura (expected %s)", mappedType)
		}
		if !strings.Contains(xmlStr, "<sf:CuotaTotal>") {
			t.Error("missing sf:CuotaTotal")
		}
		if !strings.Contains(xmlStr, "<sf:ImporteTotal>") {
			t.Error("missing sf:ImporteTotal")
		}
		if !strings.Contains(xmlStr, fmt.Sprintf("<sf:CuotaTotal>%.2f</sf:CuotaTotal>", taxAmt)) {
			t.Errorf("sf:CuotaTotal value mismatch, expected %.2f", taxAmt)
		}
		if !strings.Contains(xmlStr, fmt.Sprintf("<sf:ImporteTotal>%.2f</sf:ImporteTotal>", totalAmt)) {
			t.Errorf("sf:ImporteTotal value mismatch, expected %.2f", totalAmt)
		}
		expectedDate := issueDate.Format("02-01-2006")
		if !strings.Contains(xmlStr, expectedDate) {
			t.Errorf("date not in DD-MM-YYYY format, expected %s", expectedDate)
		}
		if !strings.Contains(xmlStr, "<sf:Encadenamiento>") {
			t.Error("missing sf:Encadenamiento")
		}
		if !strings.Contains(xmlStr, "<sf:PrimerRegistro>S</sf:PrimerRegistro>") {
			t.Error("missing or wrong sf:PrimerRegistro=S")
		}
		if !strings.Contains(xmlStr, "<sf:SistemaInformatico>") {
			t.Error("missing sf:SistemaInformatico")
		}
		siChecks := map[string]string{
			"<sf:NombreSistemaInformatico>":    config.NombreSistemaInformatico,
			"<sf:IdSistemaInformatico>":        config.IdSistemaInformatico,
			"<sf:Version>":                     config.Version,
			"<sf:NumeroInstalacion>":           config.NumeroInstalacion,
			"<sf:TipoUsoPosibleSoloVerifactu>": config.TipoUsoVerifactu,
			"<sf:TipoUsoPosibleMultiOT>":       config.TipoUsoMultiOT,
			"<sf:IndicadorMultiplesOT>":        config.IndicadorMultiOT,
		}
		for elem, val := range siChecks {
			if !strings.Contains(xmlStr, elem+val+"</"+elem[1:]) {
				t.Errorf("SistemaInformatico missing or wrong: %s%s", elem, val)
			}
		}
		if !strings.Contains(xmlStr, "<sf:TipoHuella>01</sf:TipoHuella>") {
			t.Error("missing sf:TipoHuella 01")
		}
		if !strings.Contains(xmlStr, "<sf:Huella>"+rec.Fingerprint+"</sf:Huella>") {
			t.Error("missing or wrong sf:Huella")
		}
		if !strings.Contains(xmlStr, "<sf:Desglose>") {
			t.Error("missing sf:Desglose")
		}
		if !strings.Contains(xmlStr, "<sf:DetalleDesglose") {
			t.Error("missing sf:DetalleDesglose")
		}
		if sigXML != "" {
			if !strings.Contains(xmlStr, "<ds:Signature") {
				t.Error("missing ds:Signature element")
			}
			if !strings.Contains(xmlStr, `xmlns:ds="`+aeat.NSSig+`"`) {
				t.Error("missing xmlns:ds namespace on Signature")
			}
		}
	})

	// ---- 12. Unmarshal round-trip ----
	t.Run("unmarshal_roundtrip", func(t *testing.T) {
		decoder := xml.NewDecoder(strings.NewReader(string(sumXML)))
		var found struct {
			cabeceraNIF  string
			cabeceraRazon string
			idVersion    string
			tipoFactura  string
			cuotaTotal   string
			importeTotal string
			huella       string
			tipoHuella   string
			primerReg    string
			siNombre     string
			detalleCount int
		}
		inCabecera := 0
		inObligado := 0
		inRegistroAlta := 0
		for {
			tok, err := decoder.Token()
			if err != nil {
				break
			}
			switch t := tok.(type) {
			case xml.StartElement:
				local := t.Name.Local
				switch {
				case local == "Cabecera":
					inCabecera++
				case local == "ObligadoEmision" && inCabecera > 0:
					inObligado++
				case local == "NIF" && inObligado > 0:
					if ct, _ := decoder.Token(); ct != nil {
						if cd, ok := ct.(xml.CharData); ok {
							found.cabeceraNIF = string(cd)
						}
					}
				case local == "NombreRazon" && inObligado > 0:
					if ct, _ := decoder.Token(); ct != nil {
						if cd, ok := ct.(xml.CharData); ok {
							found.cabeceraRazon = string(cd)
						}
					}
				case local == "RegistroAlta":
					inRegistroAlta++
				case local == "IDVersion" && inRegistroAlta > 0:
					if ct, _ := decoder.Token(); ct != nil {
						if cd, ok := ct.(xml.CharData); ok {
							found.idVersion = string(cd)
						}
					}
				case local == "TipoFactura" && inRegistroAlta > 0:
					if ct, _ := decoder.Token(); ct != nil {
						if cd, ok := ct.(xml.CharData); ok {
							found.tipoFactura = string(cd)
						}
					}
				case local == "CuotaTotal" && inRegistroAlta > 0:
					if ct, _ := decoder.Token(); ct != nil {
						if cd, ok := ct.(xml.CharData); ok {
							found.cuotaTotal = string(cd)
						}
					}
				case local == "ImporteTotal" && inRegistroAlta > 0:
					if ct, _ := decoder.Token(); ct != nil {
						if cd, ok := ct.(xml.CharData); ok {
							found.importeTotal = string(cd)
						}
					}
				case local == "Huella" && inRegistroAlta > 0:
					if ct, _ := decoder.Token(); ct != nil {
						if cd, ok := ct.(xml.CharData); ok {
							found.huella = string(cd)
						}
					}
				case local == "TipoHuella" && inRegistroAlta > 0:
					if ct, _ := decoder.Token(); ct != nil {
						if cd, ok := ct.(xml.CharData); ok {
							found.tipoHuella = string(cd)
						}
					}
				case local == "PrimerRegistro" && inRegistroAlta > 0:
					if ct, _ := decoder.Token(); ct != nil {
						if cd, ok := ct.(xml.CharData); ok {
							found.primerReg = string(cd)
						}
					}
				case local == "NombreSistemaInformatico" && inRegistroAlta > 0:
					if ct, _ := decoder.Token(); ct != nil {
						if cd, ok := ct.(xml.CharData); ok {
							found.siNombre = string(cd)
						}
					}
				case local == "DetalleDesglose" && inRegistroAlta > 0:
					found.detalleCount++
				}
			case xml.EndElement:
				local := t.Name.Local
				switch {
				case local == "Cabecera":
					inCabecera--
				case local == "ObligadoEmision":
					inObligado--
				case local == "RegistroAlta":
					inRegistroAlta--
				}
			}
		}
		if found.cabeceraNIF != req.Emisor.CIF {
			t.Errorf("round-trip NIF: got %q, want %q", found.cabeceraNIF, req.Emisor.CIF)
		}
		if found.cabeceraRazon != req.Emisor.Nombre {
			t.Errorf("round-trip NombreRazon: got %q, want %q", found.cabeceraRazon, req.Emisor.Nombre)
		}
		if found.idVersion != "1.0" {
			t.Errorf("IDVersion round-trip: got %q, want 1.0", found.idVersion)
		}
		if found.tipoFactura != mappedType {
			t.Errorf("TipoFactura round-trip: got %q, want %q", found.tipoFactura, mappedType)
		}
		if found.huella != rec.Fingerprint {
			t.Errorf("Huella round-trip: got %q, want %q", found.huella, rec.Fingerprint)
		}
		if found.tipoHuella != "01" {
			t.Errorf("TipoHuella round-trip: got %q, want 01", found.tipoHuella)
		}
		if found.detalleCount == 0 {
			t.Error("DetalleDesglose count is 0 after round-trip")
		}
		if found.cuotaTotal != fmt.Sprintf("%.2f", taxAmt) {
			t.Errorf("CuotaTotal round-trip: got %q, want %.2f", found.cuotaTotal, taxAmt)
		}
		if found.importeTotal != fmt.Sprintf("%.2f", totalAmt) {
			t.Errorf("ImporteTotal round-trip: got %q, want %.2f", found.importeTotal, totalAmt)
		}
		if found.primerReg != "S" {
			t.Errorf("PrimerRegistro round-trip: got %q, want S", found.primerReg)
		}
		if found.siNombre != config.NombreSistemaInformatico {
			t.Errorf("NombreSistemaInformatico round-trip: got %q, want %q", found.siNombre, config.NombreSistemaInformatico)
		}
	})

	// ---- 13. SOAP envelope ----
	t.Run("soap_envelope", func(t *testing.T) {
		var capturedBody []byte
		srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			buf := make([]byte, 65536)
			n, _ := r.Body.Read(buf)
			capturedBody = buf[:n]
			w.Header().Set("Content-Type", "text/xml")
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/">
  <soapenv:Body>
    <RespuestaRegFactuSistemaFacturacion xmlns="https://www2.agenciatributaria.gob.es/static_files/common/internet/dep/aplicaciones/es/aeat/tike/cont/ws/RespuestaSuministro.xsd">
      <CSV>INTEG-CSV</CSV>
      <TiempoEsperaEnvio>0000</TiempoEsperaEnvio>
      <EstadoEnvio>Correcto</EstadoEnvio>
    </RespuestaRegFactuSistemaFacturacion>
  </soapenv:Body>
</soapenv:Envelope>`))
		}))
		defer srv.Close()

		origEndpoint := aeat.Endpoints[aeat.EnvTest]
		aeat.Endpoints[aeat.EnvTest] = srv.URL
		defer func() { aeat.Endpoints[aeat.EnvTest] = origEndpoint }()

		client := aeat.NewClient(aeat.EnvTest, nil)
		client.SetHTTPClient(srv.Client())

		result, err := client.Submit(context.Background(), sumXML)
		if err != nil {
			t.Fatalf("Client.Submit failed: %v", err)
		}
		if result.CSV != "INTEG-CSV" {
			t.Errorf("CSV: got %q, want INTEG-CSV", result.CSV)
		}
		if !result.IsAccepted() {
			t.Error("expected accepted result")
		}

		soapStr := string(capturedBody)
		if !strings.Contains(soapStr, `<?xml version="1.0" encoding="UTF-8"?>`) {
			t.Error("SOAP envelope missing XML declaration")
		}
		if !strings.Contains(soapStr, "<soapenv:Envelope") {
			t.Error("SOAP envelope missing soapenv:Envelope")
		}
		if !strings.Contains(soapStr, `xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"`) {
			t.Error("SOAP envelope missing soapenv namespace")
		}
		if !strings.Contains(soapStr, "<soapenv:Header/>") {
			t.Error("SOAP envelope missing soapenv:Header")
		}
		if !strings.Contains(soapStr, "<soapenv:Body>") {
			t.Error("SOAP envelope missing soapenv:Body")
		}
		if !strings.Contains(soapStr, "<sfLR:RegFactuSistemaFacturacion") {
			t.Error("SOAP body missing SuministroLR content")
		}
	})
}
