package aeat

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/ThePromidius/facturae-engine/src/internal/chain"
	"github.com/ThePromidius/facturae-engine/src/internal/facturae"
)

// ----- helpers -----

func fakeFacturaE() *facturae.FacturaE {
	return &facturae.FacturaE{
		XmlnsFe: "http://www.facturae.es/Facturae/2009/v3.2/Facturae",
		XmlnsDs: "http://www.w3.org/2000/09/xmldsig#",
		FileHeader: facturae.FileHeader{
			SchemaVersion: "3.2.2",
			Modality:      "I",
			Batch: facturae.Batch{
				BatchIdentifier:         "BATCH001",
				InvoicesCount:           1,
				InvoiceCurrencyCode:     "EUR",
				TotalInvoicesAmount:     facturae.Amount{TotalAmount: 121.00},
				TotalOutstandingAmount:  facturae.Amount{TotalAmount: 121.00},
				TotalExecutableAmount:   facturae.Amount{TotalAmount: 121.00},
			},
		},
	}
}

func fakeInvoice(taxRate, taxableBase, taxAmount, totalTax, invoiceTotal float64, docType string, taxes []facturae.TaxOutput) facturae.Invoice {
	if taxes == nil {
		taxes = []facturae.TaxOutput{
			{
				TaxTypeCode: "01",
				TaxRate:     taxRate,
				TaxableBase: facturae.Amount{TotalAmount: taxableBase},
				TaxAmount:   facturae.Amount{TotalAmount: taxAmount},
			},
		}
	}
	return facturae.Invoice{
		InvoiceHeader: facturae.InvoiceHeader{
			InvoiceNumber:       "001",
			InvoiceSeriesCode:   "S",
			InvoiceDocumentType: docType,
			InvoiceClass:        "OO",
		},
		InvoiceIssueData: facturae.InvoiceIssueData{
			IssueDate:           "2026-05-28",
			InvoiceCurrencyCode: "EUR",
			TaxCurrencyCode:     "EUR",
			LanguageName:        "es",
		},
		TaxesOutputs: facturae.TaxesOutputs{
			Tax: taxes,
		},
		InvoiceTotals: facturae.InvoiceTotals{
			TotalTaxOutputs: totalTax,
			InvoiceTotal:    invoiceTotal,
		},
	}
}

func fakeRecord(prevFP string) chain.Record {
	return chain.Record{
		InvoiceNumber:       "001",
		InvoiceSeries:       "S",
		EmisorCIF:           "A12345678",
		PreviousFingerprint: prevFP,
		Fingerprint:         "CURRENT-HASH-64CHARSxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		Timestamp:           time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC),
	}
}

const testCIF = "A12345678"
const testNombre = "Test Emisor S.L."

const testSignature = `<ds:Signature xmlns:ds="http://www.w3.org/2000/09/xmldsig#"><ds:SignedInfo><ds:CanonicalizationMethod Algorithm="http://www.w3.org/TR/2001/REC-xml-c14n-20010315"/></ds:SignedInfo><ds:SignatureValue>abc123</ds:SignatureValue><ds:KeyInfo><ds:X509Data><ds:X509Certificate>certdata</ds:X509Certificate></ds:X509Data></ds:KeyInfo></ds:Signature>`

// ----- 1. BuildSuministroLR with PrimerRegistro (first invoice, no previous fingerprint) -----
func TestBuildSuministroLR_PrimerRegistro(t *testing.T) {
	inv := fakeInvoice(21.00, 100.00, 21.00, 21.00, 121.00, "FC", nil)
	rec := fakeRecord("")
	reg := BuildSuministroLR(fakeFacturaE(), inv, rec, testCIF, testNombre, "", DefaultConfig())

	if reg.Registros[0].RegistroAlta == nil {
		t.Fatal("RegistroAlta is nil")
	}
	enc := reg.Registros[0].RegistroAlta.Encadenamiento
	if enc.PrimerRegistro != "S" {
		t.Errorf("PrimerRegistro = %q, want %q", enc.PrimerRegistro, "S")
	}
	if enc.RegistroAnterior != nil {
		t.Error("RegistroAnterior should be nil for first record")
	}
}

// ----- 2. BuildSuministroLR with RegistroAnterior (second invoice, has PreviousFingerprint) -----
func TestBuildSuministroLR_RegistroAnterior(t *testing.T) {
	inv := fakeInvoice(21.00, 100.00, 21.00, 21.00, 121.00, "FC", nil)
	prevFP := "PREVIOUS-HASH-64CHARSxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	rec := fakeRecord(prevFP)
	reg := BuildSuministroLR(fakeFacturaE(), inv, rec, testCIF, testNombre, "", DefaultConfig())

	if reg.Registros[0].RegistroAlta == nil {
		t.Fatal("RegistroAlta is nil")
	}
	enc := reg.Registros[0].RegistroAlta.Encadenamiento
	if enc.PrimerRegistro != "" {
		t.Errorf("PrimerRegistro should be empty for non-first record, got %q", enc.PrimerRegistro)
	}
	if enc.RegistroAnterior == nil {
		t.Fatal("RegistroAnterior should not be nil for second record")
	}
	if enc.RegistroAnterior.IDEmisorFactura != testCIF {
		t.Errorf("IDEmisorFactura = %q, want %q", enc.RegistroAnterior.IDEmisorFactura, testCIF)
	}
	if enc.RegistroAnterior.NumSerieFactura != "S001" {
		t.Errorf("NumSerieFactura = %q, want %q", enc.RegistroAnterior.NumSerieFactura, "S001")
	}
	if enc.RegistroAnterior.FechaExpedicionFactura != "28-05-2026" {
		t.Errorf("FechaExpedicionFactura = %q, want %q", enc.RegistroAnterior.FechaExpedicionFactura, "28-05-2026")
	}
	if enc.RegistroAnterior.Huella != prevFP {
		t.Errorf("Huella = %q, want %q", enc.RegistroAnterior.Huella, prevFP)
	}
}

// ----- 3. BuildSuministroLR with zero amounts -----
func TestBuildSuministroLR_ZeroAmounts(t *testing.T) {
	inv := fakeInvoice(0, 0, 0, 0, 0, "FC", nil)
	rec := fakeRecord("")
	reg := BuildSuministroLR(fakeFacturaE(), inv, rec, testCIF, testNombre, "", DefaultConfig())

	alta := reg.Registros[0].RegistroAlta
	if alta.CuotaTotal != "0.00" {
		t.Errorf("CuotaTotal = %q, want %q", alta.CuotaTotal, "0.00")
	}
	if alta.ImporteTotal != "0.00" {
		t.Errorf("ImporteTotal = %q, want %q", alta.ImporteTotal, "0.00")
	}
	if len(alta.Desglose.DetalleDesglose) > 0 {
		dd := alta.Desglose.DetalleDesglose[0]
		if dd.TipoImpositivo != "0.00" {
			t.Errorf("TipoImpositivo = %q, want %q", dd.TipoImpositivo, "0.00")
		}
		if dd.BaseImponibleOimporteNoSujeto != "0.00" {
			t.Errorf("BaseImponible = %q, want %q", dd.BaseImponibleOimporteNoSujeto, "0.00")
		}
		if dd.CuotaRepercutida != "0.00" {
			t.Errorf("CuotaRepercutida = %q, want %q", dd.CuotaRepercutida, "0.00")
		}
	}
}

// ----- 4. BuildSuministroLR with very large amounts -----
func TestBuildSuministroLR_LargeAmounts(t *testing.T) {
	large := 9999999999.99
	inv := fakeInvoice(21.00, large, large*0.21, large*0.21, large*1.21, "FC", nil)
	rec := fakeRecord("")
	reg := BuildSuministroLR(fakeFacturaE(), inv, rec, testCIF, testNombre, "", DefaultConfig())

	alta := reg.Registros[0].RegistroAlta
	if !strings.Contains(alta.CuotaTotal, ".") {
		t.Error("CuotaTotal should contain decimal separator")
	}
	parts := strings.Split(alta.CuotaTotal, ".")
	if len(parts) != 2 || len(parts[1]) != 2 {
		t.Errorf("CuotaTotal should have 2 decimal places, got %q", alta.CuotaTotal)
	}
	parts2 := strings.Split(alta.ImporteTotal, ".")
	if len(parts2) != 2 || len(parts2[1]) != 2 {
		t.Errorf("ImporteTotal should have 2 decimal places, got %q", alta.ImporteTotal)
	}
}

// ----- 5. BuildSuministroLR with multiple tax brackets (2+ DetalleDesglose entries) -----
func TestBuildSuministroLR_MultipleTaxBrackets(t *testing.T) {
	taxes := []facturae.TaxOutput{
		{
			TaxTypeCode: "01",
			TaxRate:     21.00,
			TaxableBase: facturae.Amount{TotalAmount: 100.00},
			TaxAmount:   facturae.Amount{TotalAmount: 21.00},
		},
		{
			TaxTypeCode: "01",
			TaxRate:     10.00,
			TaxableBase: facturae.Amount{TotalAmount: 50.00},
			TaxAmount:   facturae.Amount{TotalAmount: 5.00},
		},
	}
	inv := fakeInvoice(0, 0, 0, 26.00, 176.00, "FC", taxes)
	rec := fakeRecord("")
	reg := BuildSuministroLR(fakeFacturaE(), inv, rec, testCIF, testNombre, "", DefaultConfig())

	desglose := reg.Registros[0].RegistroAlta.Desglose.DetalleDesglose
	if len(desglose) != 2 {
		t.Fatalf("expected 2 DetalleDesglose entries, got %d", len(desglose))
	}
	if desglose[0].TipoImpositivo != "21.00" {
		t.Errorf("first bracket TipoImpositivo = %q, want %q", desglose[0].TipoImpositivo, "21.00")
	}
	if desglose[0].BaseImponibleOimporteNoSujeto != "100.00" {
		t.Errorf("first bracket BaseImponible = %q, want %q", desglose[0].BaseImponibleOimporteNoSujeto, "100.00")
	}
	if desglose[1].TipoImpositivo != "10.00" {
		t.Errorf("second bracket TipoImpositivo = %q, want %q", desglose[1].TipoImpositivo, "10.00")
	}
	if desglose[1].BaseImponibleOimporteNoSujeto != "50.00" {
		t.Errorf("second bracket BaseImponible = %q, want %q", desglose[1].BaseImponibleOimporteNoSujeto, "50.00")
	}
}

// ----- 6. BuildSuministroLR with signature XML (non-empty signature) -----
func TestBuildSuministroLR_WithSignature(t *testing.T) {
	inv := fakeInvoice(21.00, 100.00, 21.00, 21.00, 121.00, "FC", nil)
	rec := fakeRecord("")
	reg := BuildSuministroLR(fakeFacturaE(), inv, rec, testCIF, testNombre, testSignature, DefaultConfig())

	raw := reg.Registros[0].RegistroAlta.SignatureRaw
	if raw == "" {
		t.Fatal("SignatureRaw should not be empty")
	}
	if raw != testSignature {
		t.Errorf("SignatureRaw mismatch:\ngot:  %q\nwant: %q", raw, testSignature)
	}
}

// ----- 7. BuildSuministroLR without signature XML (empty signature string) -----
func TestBuildSuministroLR_WithoutSignature(t *testing.T) {
	inv := fakeInvoice(21.00, 100.00, 21.00, 21.00, 121.00, "FC", nil)
	rec := fakeRecord("")
	reg := BuildSuministroLR(fakeFacturaE(), inv, rec, testCIF, testNombre, "", DefaultConfig())

	raw := reg.Registros[0].RegistroAlta.SignatureRaw
	if raw != "" {
		t.Errorf("SignatureRaw should be empty, got %q", raw)
	}
}

// ----- 8. BuildSuministroLR with custom SuministroLRConfig -----
func TestBuildSuministroLR_CustomConfig(t *testing.T) {
	cfg := SuministroLRConfig{
		NombreSistemaInformatico: "Custom-SIF",
		IdSistemaInformatico:     "99",
		Version:                  "2.0.0",
		NumeroInstalacion:        "INST-999",
		TipoUsoVerifactu:         "N",
		TipoUsoMultiOT:           "S",
		IndicadorMultiOT:         "S",
	}
	inv := fakeInvoice(21.00, 100.00, 21.00, 21.00, 121.00, "FC", nil)
	rec := fakeRecord("")
	reg := BuildSuministroLR(fakeFacturaE(), inv, rec, testCIF, testNombre, "", cfg)

	si := reg.Registros[0].RegistroAlta.SistemaInformatico
	if si.NombreSistemaInformatico != "Custom-SIF" {
		t.Errorf("NombreSistemaInformatico = %q, want %q", si.NombreSistemaInformatico, "Custom-SIF")
	}
	if si.IdSistemaInformatico != "99" {
		t.Errorf("IdSistemaInformatico = %q, want %q", si.IdSistemaInformatico, "99")
	}
	if si.Version != "2.0.0" {
		t.Errorf("Version = %q, want %q", si.Version, "2.0.0")
	}
	if si.NumeroInstalacion != "INST-999" {
		t.Errorf("NumeroInstalacion = %q, want %q", si.NumeroInstalacion, "INST-999")
	}
	if si.TipoUsoPosibleSoloVerifactu != "N" {
		t.Errorf("TipoUsoPosibleSoloVerifactu = %q, want %q", si.TipoUsoPosibleSoloVerifactu, "N")
	}
	if si.TipoUsoPosibleMultiOT != "S" {
		t.Errorf("TipoUsoPosibleMultiOT = %q, want %q", si.TipoUsoPosibleMultiOT, "S")
	}
	if si.IndicadorMultiplesOT != "S" {
		t.Errorf("IndicadorMultiplesOT = %q, want %q", si.IndicadorMultiplesOT, "S")
	}
}

// ----- 9. formatAmount: positive, negative, zero, large, small -----
func TestFormatAmount(t *testing.T) {
	cases := []struct {
		name   string
		input  float64
		signed bool
		want   string
	}{
		{"positive", 123.45, false, "123.45"},
		{"negative", -50.00, false, "-50.00"},
		{"zero", 0, false, "0.00"},
		{"large", 1e10, false, "10000000000.00"},
		{"small", 0.01, false, "0.01"},
		{"positive_signed", 123.45, true, "123.45"},
		{"negative_signed", -50.00, true, "-50.00"},
		{"zero_signed", 0, true, "0.00"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := formatAmount(tc.input, tc.signed)
			if got != tc.want {
				t.Errorf("formatAmount(%v, %v) = %q, want %q", tc.input, tc.signed, got, tc.want)
			}
		})
	}
}

// ----- 10. formatFechaDDMMYYYY: valid, invalid (already DD-MM-YYYY), empty, partial -----
func TestFormatFechaDDMMYYYY(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"valid YYYY-MM-DD", "2026-05-28", "28-05-2026"},
		{"already DD-MM-YYYY", "28-05-2026", "2026-05-28"},
		{"empty string", "", ""},
		{"partial YYYY", "2026", "2026"},
		{"partial YYYY-MM", "2026-05", "2026-05"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := formatFechaDDMMYYYY(tc.input)
			if got != tc.want {
				t.Errorf("formatFechaDDMMYYYY(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// ----- 11. mapInvoiceType: FC->F1, FA->F2, AF->R1, unknown->F1 -----
func TestMapInvoiceType(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"FC", "F1"},
		{"FA", "F2"},
		{"AF", "R1"},
		{"??", "F1"},
		{"", "F1"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := mapInvoiceType(tc.input)
			if got != tc.want {
				t.Errorf("mapInvoiceType(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// ----- 12. SubmitResult.IsAccepted: all states -----
func TestSubmitResult_IsAccepted(t *testing.T) {
	cases := []struct {
		estado string
		want   bool
	}{
		{"Correcto", true},
		{"ParcialmenteCorrecto", true},
		{"AceptadoConErrores", true},
		{"Incorrecto", false},
		{"", false},
	}
	for _, tc := range cases {
		t.Run(tc.estado, func(t *testing.T) {
			r := SubmitResult{Estado: tc.estado}
			got := r.IsAccepted()
			if got != tc.want {
				t.Errorf("IsAccepted() for Estado=%q = %v, want %v", tc.estado, got, tc.want)
			}
		})
	}
}

// ----- 13. ExtractSignature: real XML with ds:Signature element -----
func TestExtractSignature_RealXML(t *testing.T) {
	xmlInput := []byte(`<?xml version="1.0"?>
<Root>
  <ds:Signature xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
    <ds:SignedInfo>
      <ds:CanonicalizationMethod Algorithm="http://www.w3.org/TR/2001/REC-xml-c14n-20010315"/>
      <ds:SignatureMethod Algorithm="http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"/>
    </ds:SignedInfo>
    <ds:SignatureValue>BASE64SIGVALUE</ds:SignatureValue>
    <ds:KeyInfo>
      <ds:X509Data>
        <ds:X509Certificate>BASE64CERT</ds:X509Certificate>
      </ds:X509Data>
    </ds:KeyInfo>
  </ds:Signature>
</Root>`)

	sig, err := ExtractSignature(xmlInput)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(sig, "<ds:Signature") {
		t.Errorf("expected prefix <ds:Signature, got %q", sig[:20])
	}
	if !strings.Contains(sig, "xmlns:ds=\"http://www.w3.org/2000/09/xmldsig#\"") {
		t.Error("output should contain xmlns:ds declaration")
	}
	if !strings.Contains(sig, "http://www.w3.org/2000/09/xmldsig#:SignedInfo") {
		t.Error("output should contain SignedInfo")
	}
	if !strings.Contains(sig, "http://www.w3.org/2000/09/xmldsig#:SignatureValue>BASE64SIGVALUE<") {
		t.Error("output should contain SignatureValue")
	}
	if !strings.Contains(sig, "http://www.w3.org/2000/09/xmldsig#:KeyInfo") {
		t.Error("output should contain KeyInfo")
	}
	if !strings.HasSuffix(sig, "</ds:Signature>") {
		t.Error("output should end with </ds:Signature>")
	}
}

// ----- 14. ExtractSignature: XML with NO signature element -----
func TestExtractSignature_NoSignature(t *testing.T) {
	xmlInput := []byte("<root><foo>bar</foo></root>")
	_, err := ExtractSignature(xmlInput)
	if err == nil {
		t.Error("expected error for XML without signature")
	}
}

// ----- 15. ExtractSignature: XML with ds:Signature but missing xmlns:ds (namespace on parent) -----
func TestExtractSignature_MissingNamespace(t *testing.T) {
	xmlInput := []byte(`<root xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
  <ds:Signature>
    <ds:SignedInfo>test</ds:SignedInfo>
    <ds:SignatureValue>val</ds:SignatureValue>
  </ds:Signature>
</root>`)

	sig, err := ExtractSignature(xmlInput)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sig, "xmlns:ds") {
		t.Error("output should explicitly contain xmlns:ds even if not on the Signature element itself")
	}
	if !strings.Contains(sig, "http://www.w3.org/2000/09/xmldsig#:SignedInfo") {
		t.Error("output should contain SignedInfo content")
	}
	if !strings.Contains(sig, "test") {
		t.Error("output should contain SignedInfo text content")
	}
	if !strings.Contains(sig, "</ds:Signature>") {
		t.Error("output should end with </ds:Signature>")
	}
}

// ----- 16. ExtractSignature: empty input -----
func TestExtractSignature_EmptyInput(t *testing.T) {
	_, err := ExtractSignature([]byte{})
	if err == nil {
		t.Error("expected error for empty input")
	}
}

// ----- 17. DefaultConfig returns expected defaults -----
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	expected := SuministroLRConfig{
		NombreSistemaInformatico: "Vivex-SIF",
		IdSistemaInformatico:     "01",
		Version:                  "1.0.0",
		NumeroInstalacion:        "INST-001",
		TipoUsoVerifactu:         "S",
		TipoUsoMultiOT:           "N",
		IndicadorMultiOT:         "N",
	}
	if cfg != expected {
		t.Errorf("DefaultConfig() = %+v, want %+v", cfg, expected)
	}
}

// ----- 18. XML round-trip: marshal a RegFactuSistemaFacturacion, unmarshal it, verify fields -----
func TestXMLRoundTrip(t *testing.T) {
	orig := RegFactuSistemaFacturacion{
		NsLR:   NSLR,
		NsInfo: NSInfo,
		NsSig:  NSSig,
		Cabecera: CabeceraType{
			ObligadoEmision: PersonaFisicaJuridicaESType{
				NombreRazon: "Round-Trip S.L.",
				NIF:         "B87654321",
			},
		},
		Registros: []RegistroFacturaType{
			{
				RegistroAlta: &RegistroFacturacionAltaType{
					IDVersion: "1.0",
					IDFactura: IDFacturaExpedidaType{
						IDEmisorFactura:       "B87654321",
						NumSerieFactura:       "RT001",
						FechaExpedicionFactura: "28-05-2026",
					},
					NombreRazonEmisor:    "Round-Trip S.L.",
					TipoFactura:          "F1",
					DescripcionOperacion: "Round trip test",
					Desglose: DesgloseType{
						DetalleDesglose: []DetalleType{
							{
								Impuesto:                     "01",
								ClaveRegimen:                 "01",
								CalificacionOperacion:        "S1",
								TipoImpositivo:               "21.00",
								BaseImponibleOimporteNoSujeto: "100.00",
								CuotaRepercutida:             "21.00",
							},
						},
					},
					CuotaTotal:   "21.00",
					ImporteTotal: "121.00",
					Encadenamiento: EncadenamientoType{
						PrimerRegistro: "S",
					},
					SistemaInformatico: SistemaInformaticoType{
						NombreRazon:                  "Round-Trip S.L.",
						NIF:                          "B87654321",
						NombreSistemaInformatico:     "RoundTrip-SIF",
						IdSistemaInformatico:         "01",
						Version:                      "1.0.0",
						NumeroInstalacion:            "INST-001",
						TipoUsoPosibleSoloVerifactu:  "S",
						TipoUsoPosibleMultiOT:        "N",
						IndicadorMultiplesOT:         "N",
					},
					FechaHoraHusoGenRegistro: "2026-05-28T12:00:00Z",
					TipoHuella:               "01",
					Huella:                   "RTHASH",
				},
			},
		},
	}

	data, err := xml.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	xmlStr := string(data)

	checks := []string{
		"<sfLR:RegFactuSistemaFacturacion",
		"xmlns:sfLR=\"" + NSLR + "\"",
		"xmlns:sf=\"" + NSInfo + "\"",
		"xmlns:ds=\"" + NSSig + "\"",
		"<sfLR:Cabecera>",
		"<sf:ObligadoEmision>",
		"<sf:NombreRazon>Round-Trip S.L.</sf:NombreRazon>",
		"<sf:NIF>B87654321</sf:NIF>",
		"<sf:IDVersion>1.0</sf:IDVersion>",
		"<sf:IDFactura>",
		"<sf:NumSerieFactura>RT001</sf:NumSerieFactura>",
		"<sf:FechaExpedicionFactura>28-05-2026</sf:FechaExpedicionFactura>",
		"<sf:NombreRazonEmisor>Round-Trip S.L.</sf:NombreRazonEmisor>",
		"<sf:TipoFactura>F1</sf:TipoFactura>",
		"<sf:DescripcionOperacion>Round trip test</sf:DescripcionOperacion>",
		"<sf:Desglose>",
		"<sf:DetalleDesglose>",
		"<sf:Impuesto>01</sf:Impuesto>",
		"<sf:ClaveRegimen>01</sf:ClaveRegimen>",
		"<sf:CalificacionOperacion>S1</sf:CalificacionOperacion>",
		"<sf:TipoImpositivo>21.00</sf:TipoImpositivo>",
		"<sf:BaseImponibleOimporteNoSujeto>100.00</sf:BaseImponibleOimporteNoSujeto>",
		"<sf:CuotaRepercutida>21.00</sf:CuotaRepercutida>",
		"<sf:CuotaTotal>21.00</sf:CuotaTotal>",
		"<sf:ImporteTotal>121.00</sf:ImporteTotal>",
		"<sf:Encadenamiento>",
		"<sf:PrimerRegistro>S</sf:PrimerRegistro>",
		"<sf:SistemaInformatico>",
		"<sf:NombreSistemaInformatico>RoundTrip-SIF</sf:NombreSistemaInformatico>",
		"<sf:IdSistemaInformatico>01</sf:IdSistemaInformatico>",
		"<sf:FechaHoraHusoGenRegistro>2026-05-28T12:00:00Z</sf:FechaHoraHusoGenRegistro>",
		"<sf:TipoHuella>01</sf:TipoHuella>",
		"<sf:Huella>RTHASH</sf:Huella>",
	}
	for _, c := range checks {
		if !strings.Contains(xmlStr, c) {
			t.Errorf("marshalled XML missing expected content: %s", c)
		}
	}
}

// ----- 19. XML namespaces: verify all three namespaces appear in marshalled output -----
func TestXMLNamespaces(t *testing.T) {
	reg := RegFactuSistemaFacturacion{
		NsLR:   NSLR,
		NsInfo: NSInfo,
		NsSig:  NSSig,
		Cabecera: CabeceraType{
			ObligadoEmision: PersonaFisicaJuridicaESType{
				NombreRazon: "NS Test S.L.",
				NIF:         "A12345678",
			},
		},
		Registros: []RegistroFacturaType{
			{
				RegistroAlta: &RegistroFacturacionAltaType{
					IDVersion: "1.0",
					IDFactura: IDFacturaExpedidaType{
						IDEmisorFactura:       "A12345678",
						NumSerieFactura:       "NS001",
						FechaExpedicionFactura: "28-05-2026",
					},
					NombreRazonEmisor:    "NS Test S.L.",
					TipoFactura:          "F1",
					DescripcionOperacion: "namespace test",
					Desglose: DesgloseType{
						DetalleDesglose: []DetalleType{
							{
								Impuesto:                     "01",
								ClaveRegimen:                 "01",
								CalificacionOperacion:        "S1",
								TipoImpositivo:               "21.00",
								BaseImponibleOimporteNoSujeto: "100.00",
								CuotaRepercutida:             "21.00",
							},
						},
					},
					CuotaTotal:   "21.00",
					ImporteTotal: "121.00",
					Encadenamiento: EncadenamientoType{
						PrimerRegistro: "S",
					},
					SistemaInformatico: SistemaInformaticoType{
						NombreRazon:                  "NS Test S.L.",
						NIF:                          "A12345678",
						NombreSistemaInformatico:     "NS-SIF",
						IdSistemaInformatico:         "01",
						Version:                      "1.0.0",
						NumeroInstalacion:            "INST-001",
						TipoUsoPosibleSoloVerifactu:  "S",
						TipoUsoPosibleMultiOT:        "N",
						IndicadorMultiplesOT:         "N",
					},
					FechaHoraHusoGenRegistro: "2026-05-28T12:00:00Z",
					TipoHuella:               "01",
					Huella:                   "NS-HASH",
				},
			},
		},
	}

	data, err := xml.Marshal(reg)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	xmlStr := string(data)

	checks := []struct {
		name string
		attr string
	}{
		{"sfLR", "xmlns:sfLR=\"" + NSLR + "\""},
		{"sf", "xmlns:sf=\"" + NSInfo + "\""},
		{"ds", "xmlns:ds=\"" + NSSig + "\""},
	}
	for _, c := range checks {
		if !strings.Contains(xmlStr, c.attr) {
			t.Errorf("marshalled XML missing namespace %s", c.name)
		}
	}
}

// ----- 20. Encadenamiento: PrimerRegistro="S" appears for first record -----
func TestEncadenamiento_PrimerRegistro(t *testing.T) {
	inv := fakeInvoice(21.00, 100.00, 21.00, 21.00, 121.00, "FC", nil)
	rec := fakeRecord("")
	reg := BuildSuministroLR(fakeFacturaE(), inv, rec, testCIF, testNombre, "", DefaultConfig())

	enc := reg.Registros[0].RegistroAlta.Encadenamiento
	if enc.PrimerRegistro != "S" {
		t.Errorf("first record should have PrimerRegistro=\"S\", got %q", enc.PrimerRegistro)
	}
	if enc.RegistroAnterior != nil {
		t.Error("first record should not have RegistroAnterior")
	}
}

// ----- 21. Encadenamiento: RegistroAnterior appears with correct fields for second record -----
func TestEncadenamiento_RegistroAnterior(t *testing.T) {
	inv := fakeInvoice(21.00, 100.00, 21.00, 21.00, 121.00, "FC", nil)
	rec := fakeRecord("PREVIOUS-FP-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
	reg := BuildSuministroLR(fakeFacturaE(), inv, rec, testCIF, testNombre, "", DefaultConfig())

	enc := reg.Registros[0].RegistroAlta.Encadenamiento
	if enc.PrimerRegistro != "" {
		t.Errorf("second record should not have PrimerRegistro, got %q", enc.PrimerRegistro)
	}
	if enc.RegistroAnterior == nil {
		t.Fatal("second record should have RegistroAnterior")
	}
	if enc.RegistroAnterior.IDEmisorFactura != testCIF {
		t.Errorf("RegistroAnterior.IDEmisorFactura = %q, want %q", enc.RegistroAnterior.IDEmisorFactura, testCIF)
	}
	if enc.RegistroAnterior.NumSerieFactura != "S001" {
		t.Errorf("RegistroAnterior.NumSerieFactura = %q, want %q", enc.RegistroAnterior.NumSerieFactura, "S001")
	}
	if enc.RegistroAnterior.FechaExpedicionFactura != "28-05-2026" {
		t.Errorf("RegistroAnterior.FechaExpedicionFactura = %q, want %q", enc.RegistroAnterior.FechaExpedicionFactura, "28-05-2026")
	}
	if enc.RegistroAnterior.Huella != "PREVIOUS-FP-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" {
		t.Errorf("RegistroAnterior.Huella = %q", enc.RegistroAnterior.Huella)
	}
}

// ----- 22. Desglose: verify DetalleDesglose list is correctly populated -----
func TestDesglose_DetalleDesglose(t *testing.T) {
	taxes := []facturae.TaxOutput{
		{
			TaxTypeCode: "01",
			TaxRate:     21.00,
			TaxableBase: facturae.Amount{TotalAmount: 100.00},
			TaxAmount:   facturae.Amount{TotalAmount: 21.00},
		},
	}
	inv := fakeInvoice(21.00, 100.00, 21.00, 21.00, 121.00, "FC", taxes)
	rec := fakeRecord("")
	reg := BuildSuministroLR(fakeFacturaE(), inv, rec, testCIF, testNombre, "", DefaultConfig())

	dd := reg.Registros[0].RegistroAlta.Desglose.DetalleDesglose
	if len(dd) != 1 {
		t.Fatalf("expected 1 DetalleDesglose entry, got %d", len(dd))
	}
	det := dd[0]
	if det.Impuesto != "01" {
		t.Errorf("Impuesto = %q, want %q", det.Impuesto, "01")
	}
	if det.ClaveRegimen != "01" {
		t.Errorf("ClaveRegimen = %q, want %q", det.ClaveRegimen, "01")
	}
	if det.CalificacionOperacion != "S1" {
		t.Errorf("CalificacionOperacion = %q, want %q", det.CalificacionOperacion, "S1")
	}
	if det.TipoImpositivo != "21.00" {
		t.Errorf("TipoImpositivo = %q, want %q", det.TipoImpositivo, "21.00")
	}
	if det.BaseImponibleOimporteNoSujeto != "100.00" {
		t.Errorf("BaseImponible = %q, want %q", det.BaseImponibleOimporteNoSujeto, "100.00")
	}
	if det.CuotaRepercutida != "21.00" {
		t.Errorf("CuotaRepercutida = %q, want %q", det.CuotaRepercutida, "21.00")
	}
}

// ----- 23. SistemaInformatico: all fields populated correctly -----
func TestSistemaInformatico_Fields(t *testing.T) {
	inv := fakeInvoice(21.00, 100.00, 21.00, 21.00, 121.00, "FC", nil)
	rec := fakeRecord("")
	reg := BuildSuministroLR(fakeFacturaE(), inv, rec, testCIF, testNombre, "", DefaultConfig())

	si := reg.Registros[0].RegistroAlta.SistemaInformatico
	if si.NombreRazon != testNombre {
		t.Errorf("NombreRazon = %q, want %q", si.NombreRazon, testNombre)
	}
	if si.NIF != testCIF {
		t.Errorf("NIF = %q, want %q", si.NIF, testCIF)
	}
	if si.NombreSistemaInformatico != "Vivex-SIF" {
		t.Errorf("NombreSistemaInformatico = %q, want %q", si.NombreSistemaInformatico, "Vivex-SIF")
	}
	if si.IdSistemaInformatico != "01" {
		t.Errorf("IdSistemaInformatico = %q, want %q", si.IdSistemaInformatico, "01")
	}
	if si.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", si.Version, "1.0.0")
	}
	if si.NumeroInstalacion != "INST-001" {
		t.Errorf("NumeroInstalacion = %q, want %q", si.NumeroInstalacion, "INST-001")
	}
	if si.TipoUsoPosibleSoloVerifactu != "S" {
		t.Errorf("TipoUsoPosibleSoloVerifactu = %q, want %q", si.TipoUsoPosibleSoloVerifactu, "S")
	}
	if si.TipoUsoPosibleMultiOT != "N" {
		t.Errorf("TipoUsoPosibleMultiOT = %q, want %q", si.TipoUsoPosibleMultiOT, "N")
	}
	if si.IndicadorMultiplesOT != "N" {
		t.Errorf("IndicadorMultiplesOT = %q, want %q", si.IndicadorMultiplesOT, "N")
	}
}

// ----- 24. Cabecera: ObligadoEmision has both NombreRazon and NIF -----
func TestCabecera_ObligadoEmision(t *testing.T) {
	inv := fakeInvoice(21.00, 100.00, 21.00, 21.00, 121.00, "FC", nil)
	rec := fakeRecord("")
	reg := BuildSuministroLR(fakeFacturaE(), inv, rec, testCIF, testNombre, "", DefaultConfig())

	cab := reg.Cabecera
	if cab.ObligadoEmision.NombreRazon != testNombre {
		t.Errorf("NombreRazon = %q, want %q", cab.ObligadoEmision.NombreRazon, testNombre)
	}
	if cab.ObligadoEmision.NIF != testCIF {
		t.Errorf("NIF = %q, want %q", cab.ObligadoEmision.NIF, testCIF)
	}
}

// ----- 25. FechaExpedicionFactura: verify format is DD-MM-YYYY not YYYY-MM-DD -----
func TestFechaExpedicionFactura_Format(t *testing.T) {
	inv := fakeInvoice(21.00, 100.00, 21.00, 21.00, 121.00, "FC", nil)
	rec := fakeRecord("")
	reg := BuildSuministroLR(fakeFacturaE(), inv, rec, testCIF, testNombre, "", DefaultConfig())

	fecha := reg.Registros[0].RegistroAlta.IDFactura.FechaExpedicionFactura
	if fecha != "28-05-2026" {
		t.Errorf("FechaExpedicionFactura = %q, want %q (DD-MM-YYYY)", fecha, "28-05-2026")
	}
	if strings.Contains(fecha, "2026-05") {
		t.Error("FechaExpedicionFactura should use DD-MM-YYYY format, not YYYY-MM-DD")
	}
}
