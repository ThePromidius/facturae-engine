// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package invoice_test

import (
	"testing"
	"time"

	"github.com/ThePromidius/facturae-engine/src/internal/invoice"
)

func validRequest() invoice.Request {
	return invoice.Request{
		Meta: invoice.Meta{Version: "3.2.2", Moneda: "EUR"},
		Factura: invoice.Factura{
			Numero: "2024-0001",
			Serie:  "F",
			Fecha:  time.Now(),
		},
		Emisor: invoice.Party{
			CIF:       "B12345678",
			Nombre:    "Mi Empresa S.L.",
			Direccion: "Calle Falsa 123",
			CP:        "28001",
			Ciudad:    "Madrid",
			Provincia: "Madrid",
			Pais:      "ESP",
		},
		Receptor: invoice.Party{
			CIF:    "A87654321",
			Nombre: "Cliente S.A.",
		},
		Lineas: []invoice.Linea{
			{Descripcion: "Consultoria", Cantidad: 10, PrecioUnitario: 100, IVATipo: 21},
		},
	}
}

func TestValidate_HappyPath(t *testing.T) {
	if err := invoice.Validate(validRequest()); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestValidate_MissingNumero(t *testing.T) {
	req := validRequest()
	req.Factura.Numero = ""
	err := invoice.Validate(req)
	assertValidationError(t, err, "factura.numero")
}

func TestValidate_MissingEmisorCIF(t *testing.T) {
	req := validRequest()
	req.Emisor.CIF = ""
	err := invoice.Validate(req)
	assertValidationError(t, err, "emisor.cif")
}

func TestValidate_MissingEmisorDireccion(t *testing.T) {
	req := validRequest()
	req.Emisor.Direccion = ""
	err := invoice.Validate(req)
	assertValidationError(t, err, "emisor.direccion")
}

func TestValidate_EmptyLineas(t *testing.T) {
	req := validRequest()
	req.Lineas = nil
	err := invoice.Validate(req)
	assertValidationError(t, err, "lineas")
}

func TestValidate_NegativeCantidad(t *testing.T) {
	req := validRequest()
	req.Lineas[0].Cantidad = -1
	err := invoice.Validate(req)
	assertValidationError(t, err, "cantidad")
}

func TestValidate_InvalidIVA(t *testing.T) {
	req := validRequest()
	req.Lineas[0].IVATipo = 150
	err := invoice.Validate(req)
	assertValidationError(t, err, "iva_tipo")
}

func TestValidate_MultipleErrors(t *testing.T) {
	req := validRequest()
	req.Factura.Numero = ""
	req.Emisor.CIF = ""
	req.Lineas = nil

	ve, ok := invoice.Validate(req).(*invoice.ValidationError)
	if !ok {
		t.Fatal("expected *ValidationError")
	}
	if len(ve.Errors) < 3 {
		t.Errorf("expected at least 3 errors, got %d: %v", len(ve.Errors), ve.Errors)
	}
}

func TestDefaultMeta(t *testing.T) {
	m := invoice.DefaultMeta(invoice.Meta{})
	if m.Version != "3.2.2" {
		t.Errorf("expected default version 3.2.2, got %s", m.Version)
	}
	if m.Moneda != "EUR" {
		t.Errorf("expected default currency EUR, got %s", m.Moneda)
	}
}

func assertValidationError(t *testing.T, err error, contains string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected validation error containing %q, got nil", contains)
	}
	if ve, ok := err.(*invoice.ValidationError); ok {
		for _, e := range ve.Errors {
			if len(e) > 0 {
				if stringContains(e, contains) {
					return
				}
			}
		}
		t.Errorf("validation error %q not found in: %v", contains, ve.Errors)
	} else {
		t.Errorf("expected *ValidationError, got %T: %v", err, err)
	}
}

func stringContains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
