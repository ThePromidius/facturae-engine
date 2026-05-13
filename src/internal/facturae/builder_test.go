// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package facturae_test

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/ThePromidius/facturae-engine/src/internal/facturae"
	"github.com/ThePromidius/facturae-engine/src/internal/invoice"
)

func testRequest() invoice.Request {
	return invoice.Request{
		Meta: invoice.Meta{Version: "3.2.2", Moneda: "EUR"},
		Factura: invoice.Factura{
			Numero: "2024-0001",
			Serie:  "F",
			Fecha:  time.Date(2024, 5, 12, 0, 0, 0, 0, time.UTC),
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
			{Descripcion: "Hosting", Cantidad: 1, PrecioUnitario: 50, IVATipo: 21},
		},
	}
}

func mustBuild(t *testing.T, req invoice.Request) *facturae.FacturaE {
	t.Helper()
	f, err := facturae.Build(req)
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}
	return f
}

func marshalXML(t *testing.T, f *facturae.FacturaE) []byte {
	t.Helper()
	out, err := xml.MarshalIndent(f, "", "  ")
	if err != nil {
		t.Fatalf("xml.MarshalIndent failed: %v", err)
	}
	return append([]byte(xml.Header), out...)
}

func TestBuild_StructureIsCorrect(t *testing.T) {
	f := mustBuild(t, testRequest())

	if f.FileHeader.SchemaVersion != "3.2.2" {
		t.Errorf("SchemaVersion: got %q, want 3.2.2", f.FileHeader.SchemaVersion)
	}
	if f.Parties.SellerParty.TaxIdentification.TaxIdentificationNumber != "B12345678" {
		t.Error("Seller CIF mismatch")
	}
	if f.Parties.BuyerParty.TaxIdentification.TaxIdentificationNumber != "A87654321" {
		t.Error("Buyer CIF mismatch")
	}
	if len(f.Invoices.Invoice) != 1 {
		t.Fatalf("expected 1 invoice, got %d", len(f.Invoices.Invoice))
	}
}

func TestBuild_TotalsAreCorrect(t *testing.T) {
	f := mustBuild(t, testRequest())
	inv := f.Invoices.Invoice[0]

	wantGross := 1050.0
	wantTax := 220.50
	wantTotal := 1270.50

	if inv.InvoiceTotals.TotalGrossAmount != wantGross {
		t.Errorf("TotalGrossAmount: got %.2f, want %.2f", inv.InvoiceTotals.TotalGrossAmount, wantGross)
	}
	if inv.InvoiceTotals.TotalTaxOutputs != wantTax {
		t.Errorf("TotalTaxOutputs: got %.2f, want %.2f", inv.InvoiceTotals.TotalTaxOutputs, wantTax)
	}
	if inv.InvoiceTotals.InvoiceTotal != wantTotal {
		t.Errorf("InvoiceTotal: got %.2f, want %.2f", inv.InvoiceTotals.InvoiceTotal, wantTotal)
	}
}

func TestBuild_MultipleVATBrackets(t *testing.T) {
	req := testRequest()
	req.Lineas = []invoice.Linea{
		{Descripcion: "Normal", Cantidad: 1, PrecioUnitario: 100, IVATipo: 21},
		{Descripcion: "Reducido", Cantidad: 1, PrecioUnitario: 100, IVATipo: 10},
		{Descripcion: "Superreducido", Cantidad: 1, PrecioUnitario: 100, IVATipo: 4},
	}
	f := mustBuild(t, req)
	inv := f.Invoices.Invoice[0]

	if len(inv.TaxesOutputs.Tax) != 3 {
		t.Errorf("expected 3 tax brackets, got %d", len(inv.TaxesOutputs.Tax))
	}
	if inv.InvoiceTotals.InvoiceTotal != 335.0 {
		t.Errorf("InvoiceTotal: got %.2f, want 335.00", inv.InvoiceTotals.InvoiceTotal)
	}
}

func TestBuild_ZeroVAT(t *testing.T) {
	req := testRequest()
	req.Lineas = []invoice.Linea{
		{Descripcion: "Exento", Cantidad: 5, PrecioUnitario: 200, IVATipo: 0},
	}
	f := mustBuild(t, req)
	inv := f.Invoices.Invoice[0]

	if inv.InvoiceTotals.TotalTaxOutputs != 0 {
		t.Errorf("expected 0 tax for 0%% IVA, got %.2f", inv.InvoiceTotals.TotalTaxOutputs)
	}
	if inv.InvoiceTotals.InvoiceTotal != 1000.0 {
		t.Errorf("InvoiceTotal: got %.2f, want 1000.00", inv.InvoiceTotals.InvoiceTotal)
	}
}

func TestBuild_IssueDate(t *testing.T) {
	f := mustBuild(t, testRequest())
	inv := f.Invoices.Invoice[0]
	if inv.InvoiceIssueData.IssueDate != "2024-05-12" {
		t.Errorf("IssueDate: got %q, want 2024-05-12", inv.InvoiceIssueData.IssueDate)
	}
}

func TestBuild_SellerAddress(t *testing.T) {
	f := mustBuild(t, testRequest())
	addr := f.Parties.SellerParty.LegalEntity.RegistrationData
	if addr == nil {
		t.Fatal("SellerParty should have address")
	}
	if addr.PostCode != "28001" {
		t.Errorf("PostCode: got %q, want 28001", addr.PostCode)
	}
}

func TestBuild_BatchCounterMatchesInvoices(t *testing.T) {
	f := mustBuild(t, testRequest())
	if f.FileHeader.Batch.InvoicesCount != len(f.Invoices.Invoice) {
		t.Error("Batch.InvoicesCount does not match number of invoices")
	}
}

func TestValidateStruct_HappyPath(t *testing.T) {
	f := mustBuild(t, testRequest())
	if err := facturae.ValidateStruct(f); err != nil {
		t.Fatalf("ValidateStruct() unexpected error: %v", err)
	}
}

func TestValidateStruct_MissingSchemaVersion(t *testing.T) {
	f := mustBuild(t, testRequest())
	f.FileHeader.SchemaVersion = ""
	err := facturae.ValidateStruct(f)
	if err == nil {
		t.Fatal("expected error for missing SchemaVersion")
	}
}

func TestValidateStruct_NegativeInvoiceTotal(t *testing.T) {
	f := mustBuild(t, testRequest())
	f.Invoices.Invoice[0].InvoiceTotals.InvoiceTotal = -1
	err := facturae.ValidateStruct(f)
	if err == nil {
		t.Fatal("expected error for negative InvoiceTotal")
	}
}

func TestValidateXMLBytes_HappyPath(t *testing.T) {
	f := mustBuild(t, testRequest())
	data := marshalXML(t, f)
	if err := facturae.ValidateXMLBytes(data); err != nil {
		t.Fatalf("ValidateXMLBytes() unexpected error: %v", err)
	}
}

func TestValidateXMLBytes_TooShort(t *testing.T) {
	err := facturae.ValidateXMLBytes([]byte("<xml/>"))
	if err == nil {
		t.Fatal("expected error for short XML")
	}
}

func TestValidateXMLBytes_InvalidXML(t *testing.T) {
	garbage := []byte(strings.Repeat("garbage xml content not parseable ", 20))
	err := facturae.ValidateXMLBytes(garbage)
	if err == nil {
		t.Fatal("expected error for invalid XML")
	}
}

func TestValidateXMLBytes_WrongRootElement(t *testing.T) {
	wrongXML := []byte(`<?xml version="1.0" encoding="UTF-8"?>` + strings.Repeat(" ", 500) + `<Invoice xmlns:fe="http://x"/>`)
	err := facturae.ValidateXMLBytes(wrongXML)
	if err == nil {
		t.Fatal("expected error for wrong root element")
	}
}

func TestXMLOutput_ContainsNamespace(t *testing.T) {
	f := mustBuild(t, testRequest())
	data := marshalXML(t, f)
	if !strings.Contains(string(data), "facturae.gob.es") {
		t.Error("XML should contain the official Hacienda namespace")
	}
}

func TestXMLOutput_ContainsCIF(t *testing.T) {
	f := mustBuild(t, testRequest())
	data := marshalXML(t, f)
	xmlStr := string(data)
	if !strings.Contains(xmlStr, "B12345678") {
		t.Error("XML should contain the seller CIF")
	}
	if !strings.Contains(xmlStr, "A87654321") {
		t.Error("XML should contain the buyer CIF")
	}
}

func TestXMLOutput_ContainsTotals(t *testing.T) {
	f := mustBuild(t, testRequest())
	data := marshalXML(t, f)
	if !strings.Contains(string(data), "1270.5") {
		t.Error("XML should contain the invoice total 1270.50")
	}
}
