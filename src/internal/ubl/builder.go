// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package ubl

import (
	"encoding/xml"
	"fmt"
	"github.com/ThePromidius/facturae-engine/src/internal/invoice"
)

// UBL Namespaces
const (
	XmlnsUbl = "urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"
	XmlnsCac = "urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2"
	XmlnsCbc = "urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2"
)

// Invoice represents a UBL 2.1 Invoice structure (EN 16931 compliant).
type Invoice struct {
	XMLName           xml.Name          `xml:"urn:oasis:names:specification:ubl:schema:xsd:Invoice-2 Invoice"`
	XmlnsUbl          string            `xml:"xmlns:ubl,attr"`
	XmlnsCac          string            `xml:"xmlns:cac,attr"`
	XmlnsCbc          string            `xml:"xmlns:cbc,attr"`
	ID                string            `xml:"cbc:ID"`
	IssueDate         string            `xml:"cbc:IssueDate"`
	InvoiceTypeCode   string            `xml:"cbc:InvoiceTypeCode"`
	DocumentCurrency  string            `xml:"cbc:DocumentCurrencyCode"`
	SupplierParty     SupplierParty     `xml:"cac:AccountingSupplierParty"`
	CustomerParty     CustomerParty     `xml:"cac:AccountingCustomerParty"`
	TaxTotal          TaxTotal          `xml:"cac:TaxTotal"`
	LegalMonetaryTotal LegalMonetaryTotal `xml:"cac:LegalMonetaryTotal"`
	InvoiceLines      []InvoiceLine     `xml:"cac:InvoiceLine"`
}

type SupplierParty struct {
	Party Party `xml:"cac:Party"`
}

type CustomerParty struct {
	Party Party `xml:"cac:Party"`
}

type Party struct {
	PartyName      *PartyName      `xml:"cac:PartyName,omitempty"`
	PartyTaxScheme *PartyTaxScheme `xml:"cac:PartyTaxScheme,omitempty"`
}

type PartyName struct {
	Name string `xml:"cbc:Name"`
}

type PartyTaxScheme struct {
	CompanyID   string      `xml:"cbc:CompanyID"`
	TaxScheme   TaxScheme   `xml:"cac:TaxScheme"`
}

type TaxScheme struct {
	ID string `xml:"cbc:ID"` // e.g., VAT
}

type TaxTotal struct {
	TaxAmount Amount `xml:"cbc:TaxAmount"`
}

type LegalMonetaryTotal struct {
	LineExtensionAmount Amount `xml:"cbc:LineExtensionAmount"`
	TaxExclusiveAmount  Amount `xml:"cbc:TaxExclusiveAmount"`
	TaxInclusiveAmount  Amount `xml:"cbc:TaxInclusiveAmount"`
	PayableAmount       Amount `xml:"cbc:PayableAmount"`
}

type InvoiceLine struct {
	ID            string      `xml:"cbc:ID"`
	InvoicedQty   Quantity    `xml:"cbc:InvoicedQuantity"`
	LineExtension Amount      `xml:"cbc:LineExtensionAmount"`
	Item          Item        `xml:"cac:Item"`
	Price         Price       `xml:"cac:Price"`
}

type Quantity struct {
	UnitCode string  `xml:"unitCode,attr"`
	Value    string  `xml:",chardata"`
}

type Item struct {
	Description string `xml:"cbc:Description"`
}

type Price struct {
	PriceAmount Amount `xml:"cbc:PriceAmount"`
}

type Amount struct {
	CurrencyID string `xml:"currencyID,attr"`
	Value      string `xml:",chardata"`
}

// Builder constructs UBL XML from internal invoice requests.
type Builder struct{}

func (b *Builder) Build(req invoice.Request) ([]byte, error) {
	var lineExtensionTotal float64
	var taxTotalAmount float64
	
	lines := make([]InvoiceLine, len(req.Lineas))
	for i, l := range req.Lineas {
		lineTotal := l.PrecioUnitario * l.QuantityFallback()
		lineExtensionTotal += lineTotal
		taxTotalAmount += lineTotal * (l.IVATipo / 100.0)

		lines[i] = InvoiceLine{
			ID:          fmt.Sprintf("%d", i+1),
			InvoicedQty: Quantity{UnitCode: "C62", Value: fmt.Sprintf("%.2f", l.QuantityFallback())},
			LineExtension: Amount{CurrencyID: req.Meta.Moneda, Value: fmt.Sprintf("%.2f", round2(lineTotal))},
			Item: Item{
				Description: l.Descripcion,
			},
			Price: Price{
				PriceAmount: Amount{CurrencyID: req.Meta.Moneda, Value: fmt.Sprintf("%.2f", round2(l.PrecioUnitario))},
			},
		}
	}

	totalInclusive := lineExtensionTotal + taxTotalAmount

	inv := Invoice{
		XmlnsUbl:         XmlnsUbl,
		XmlnsCac:         XmlnsCac,
		XmlnsCbc:         XmlnsCbc,
		ID:               req.Factura.Serie + "-" + req.Factura.Numero,
		IssueDate:        req.Factura.Fecha.Format("2006-01-02"),
		InvoiceTypeCode:  "380",
		DocumentCurrency: req.Meta.Moneda,
		SupplierParty: SupplierParty{
			Party: Party{
				PartyName: &PartyName{Name: req.Emisor.Nombre},
				PartyTaxScheme: &PartyTaxScheme{
					CompanyID: req.Emisor.CIF,
					TaxScheme: TaxScheme{ID: "VAT"},
				},
			},
		},
		CustomerParty: CustomerParty{
			Party: Party{
				PartyName: &PartyName{Name: req.Receptor.Nombre},
				PartyTaxScheme: &PartyTaxScheme{
					CompanyID: req.Receptor.CIF,
					TaxScheme: TaxScheme{ID: "VAT"},
				},
			},
		},
		TaxTotal: TaxTotal{
			TaxAmount: Amount{CurrencyID: req.Meta.Moneda, Value: fmt.Sprintf("%.2f", round2(taxTotalAmount))},
		},
		LegalMonetaryTotal: LegalMonetaryTotal{
			LineExtensionAmount: Amount{CurrencyID: req.Meta.Moneda, Value: fmt.Sprintf("%.2f", round2(lineExtensionTotal))},
			TaxExclusiveAmount:  Amount{CurrencyID: req.Meta.Moneda, Value: fmt.Sprintf("%.2f", round2(lineExtensionTotal))},
			TaxInclusiveAmount:  Amount{CurrencyID: req.Meta.Moneda, Value: fmt.Sprintf("%.2f", round2(totalInclusive))},
			PayableAmount:       Amount{CurrencyID: req.Meta.Moneda, Value: fmt.Sprintf("%.2f", round2(totalInclusive))},
		},
		InvoiceLines: lines,
	}

	output, err := xml.MarshalIndent(inv, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("ubl: failed to marshal xml: %w", err)
	}

	return append([]byte(xml.Header), output...), nil
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
