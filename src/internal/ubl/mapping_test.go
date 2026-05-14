package ubl_test

import (
	"encoding/xml"
	"testing"
	"time"

	"github.com/ThePromidius/facturae-engine/src/internal/invoice"
	"github.com/ThePromidius/facturae-engine/src/internal/ubl"
)

func TestBuilder_Build_Mapping(t *testing.T) {
	req := invoice.Request{
		Meta: invoice.Meta{Moneda: "EUR"},
		Factura: invoice.Factura{
			Numero: "123",
			Serie:  "SER",
			Fecha:  time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC),
		},
		Emisor: invoice.Party{CIF: "B12345678", Nombre: "Issuer S.L."},
		Receptor: invoice.Party{CIF: "A87654321", Nombre: "Client S.A."},
		Lineas: []invoice.Linea{
			{Descripcion: "Test Item", Cantidad: 10, PrecioUnitario: 100.0, IVATipo: 21.0},
		},
	}

	builder := &ubl.Builder{}
	output, err := builder.Build(req)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Basic check for UBL elements
	sOutput := string(output)
	expectedElements := []string{
		"<cbc:ID>SER-123</cbc:ID>",
		"<cbc:IssueDate>2026-05-14</cbc:IssueDate>",
		"<cbc:CompanyID>B12345678</cbc:CompanyID>",
		"<cbc:CompanyID>A87654321</cbc:CompanyID>",
		"<cbc:LineExtensionAmount currencyID=\"EUR\">1000.00</cbc:LineExtensionAmount>",
		"<cbc:TaxAmount currencyID=\"EUR\">210.00</cbc:TaxAmount>",
		"<cbc:PayableAmount currencyID=\"EUR\">1210.00</cbc:PayableAmount>",
	}

	for _, el := range expectedElements {
		if !contains(sOutput, el) {
			t.Errorf("Expected UBL output to contain %q", el)
		}
	}

	// Verify it's valid XML
	var dummy ubl.Invoice
	if err := xml.Unmarshal(output, &dummy); err != nil {
		t.Errorf("Generated UBL is not valid XML: %v", err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && func() bool {
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	}()
}
