package facturae_test

import (
	"testing"
	"time"

	"github.com/ThePromidius/facturae-engine/src/internal/invoice"
	"github.com/ThePromidius/facturae-engine/src/internal/ubl"
	"github.com/ThePromidius/facturae-engine/src/internal/facturae"
)

func TestInteroperability_UBL_and_FacturaE(t *testing.T) {
	req := invoice.Request{
		Meta: invoice.Meta{Moneda: "EUR"},
		Factura: invoice.Factura{
			Numero: "001",
			Serie:  "SERIE",
			Fecha:  time.Now(),
		},
		Emisor: invoice.Party{CIF: "B12345678", Nombre: "Test Emisor"},
	}

	// 1. Build FacturaE
	_, err := facturae.Build(req)
	if err != nil {
		t.Fatalf("FacturaE build failed: %v", err)
	}

	// 2. Build UBL (Interoperability proof)
	uBuilder := &ubl.Builder{}
	_, err = uBuilder.Build(req)
	if err != nil {
		t.Fatalf("UBL build failed: %v", err)
	}

	t.Log("Successfully generated both FacturaE and UBL from same internal model.")
}
