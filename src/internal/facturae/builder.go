// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package facturae

import (
	"github.com/ThePromidius/facturae-engine/src/internal/invoice"
	"fmt"
)

const (
	schemaFe322        = "http://www.facturae.gob.es/formato/Versiones/Facturaev3_2_2.xml"
	schemaDs           = "http://www.w3.org/2000/09/xmldsig#"
	taxTypeIVA         = "01"
	personTypeLegal    = "J"
	residenceTypeE     = "R"
	invoiceDocumentTypeFC = "FC"
	invoiceClassOR     = "OR"
	modalityI          = "I"
	issuerType         = "EM"
)

// Build converts an invoice.Request into a fully-populated FacturaE struct
// ready for XML serialization. It builds invoice lines, computes tax summaries,
// and constructs the seller and buyer parties.
func Build(req invoice.Request) (*FacturaE, error) {
	req.Meta = invoice.DefaultMeta(req.Meta)

	f := &FacturaE{
		XmlnsFe: schemaFe322,
		XmlnsDs: schemaDs,
	}

	lines, taxSummary, totals, err := buildLines(req.Lineas, req.Meta.Moneda)
	if err != nil {
		return nil, fmt.Errorf("building lines: %w", err)
	}

	f.FileHeader = FileHeader{
		SchemaVersion:     req.Meta.Version,
		Modality:          modalityI,
		InvoiceIssuerType: issuerType,
		Batch: Batch{
			BatchIdentifier:    req.Emisor.CIF + req.Factura.Serie + req.Factura.Numero,
			InvoicesCount:      1,
			TotalInvoicesAmount:  Amount{TotalAmount: round2(totals.InvoiceTotal)},
			TotalOutstandingAmount: Amount{TotalAmount: round2(totals.InvoiceTotal)},
			TotalExecutableAmount:  Amount{TotalAmount: round2(totals.InvoiceTotal)},
			InvoiceCurrencyCode: req.Meta.Moneda,
		},
	}

	f.Parties.SellerParty = buildParty(req.Emisor, true)
	f.Parties.BuyerParty = buildParty(req.Receptor, false)

	inv := Invoice{
		InvoiceHeader: InvoiceHeader{
			InvoiceNumber:       req.Factura.Numero,
			InvoiceSeriesCode:   req.Factura.Serie,
			InvoiceDocumentType: invoiceDocumentTypeFC,
			InvoiceClass:        invoiceClassOR,
		},
		InvoiceIssueData: InvoiceIssueData{
			IssueDate:           req.Factura.Fecha.Format("2006-01-02"),
			InvoiceCurrencyCode: req.Meta.Moneda,
			TaxCurrencyCode:     req.Meta.Moneda,
			LanguageName:        "es",
		},
		TaxesOutputs:  TaxesOutputs{Tax: taxSummary},
		InvoiceTotals: *totals,
		Items:         Items{InvoiceLine: lines},
	}
	f.Invoices.Invoice = append(f.Invoices.Invoice, inv)

	return f, nil
}

// buildParty converts an invoice.Party into a facturae.Party, optionally including
// the registration address when withAddress is true and the address data is non-empty.
func buildParty(p invoice.Party, withAddress bool) Party {
	party := Party{
		TaxIdentification: TaxIdentification{
			PersonTypeCode:          personTypeLegal,
			ResidenceTypeCode:       residenceTypeE,
			TaxIdentificationNumber: p.CIF,
		},
		LegalEntity: &LegalEntity{
			CorporateName: p.Nombre,
		},
	}

	if withAddress && p.Direccion != "" {
		party.LegalEntity.RegistrationData = &Address{
			Address:     p.Direccion,
			PostCode:    p.CP,
			Town:        p.Ciudad,
			Province:    p.Provincia,
			CountryCode: p.Pais,
		}
	}

	return party
}
