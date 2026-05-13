// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

// Package facturae implements the Facturae XML invoice format for Spanish electronic invoicing.
// It provides types and builders to construct, validate, and serialize Facturae 3.2.2 documents.
package facturae

import "encoding/xml"

// FacturaE represents the root element of a Facturae XML document, containing the file header,
// seller/buyer parties, and one or more invoices.
type FacturaE struct {
	XMLName xml.Name `xml:"fe:Facturae"`
	XmlnsFe string   `xml:"xmlns:fe,attr"`
	XmlnsDs string   `xml:"xmlns:ds,attr"`

	FileHeader FileHeader `xml:"FileHeader"`
	Parties    Parties    `xml:"Parties"`
	Invoices   Invoices   `xml:"Invoices"`
}

// FileHeader contains metadata about the invoice file including schema version, modality, and batch info.
type FileHeader struct {
	SchemaVersion     string `xml:"SchemaVersion"`
	Modality          string `xml:"Modality"`
	InvoiceIssuerType string `xml:"InvoiceIssuerType"`
	Batch             Batch  `xml:"Batch"`
}

// Batch groups summary information for a batch of invoices, including counts and totals.
type Batch struct {
	BatchIdentifier         string `xml:"BatchIdentifier"`
	InvoicesCount           int    `xml:"InvoicesCount"`
	TotalInvoicesAmount     Amount `xml:"TotalInvoicesAmount"`
	TotalOutstandingAmount  Amount `xml:"TotalOutstandingAmount"`
	TotalExecutableAmount   Amount `xml:"TotalExecutableAmount"`
	InvoiceCurrencyCode     string `xml:"InvoiceCurrencyCode"`
}

// Amount represents a monetary amount with an optional euro equivalent.
// TotalAmount is the primary value; EquivalentInEuros is used when converting from another currency.
type Amount struct {
	TotalAmount      float64 `xml:"TotalAmount"`
	EquivalentInEuros float64 `xml:"EquivalentInEuros,omitempty"`
}

// Parties holds the seller and buyer party information for the invoice.
type Parties struct {
	SellerParty Party `xml:"SellerParty"`
	BuyerParty  Party `xml:"BuyerParty"`
}

// Party represents a participant in the invoice (seller or buyer), with tax identification and legal entity details.
type Party struct {
	TaxIdentification TaxIdentification `xml:"TaxIdentification"`
	LegalEntity       *LegalEntity      `xml:"LegalEntity,omitempty"`
}

// TaxIdentification holds the tax-related identifiers for a party, including person type, residence type, and tax number.
type TaxIdentification struct {
	PersonTypeCode          string `xml:"PersonTypeCode"`
	ResidenceTypeCode       string `xml:"ResidenceTypeCode"`
	TaxIdentificationNumber string `xml:"TaxIdentificationNumber"`
}

// LegalEntity represents the legal entity information for a party, including the corporate name and optional registered address.
type LegalEntity struct {
	CorporateName    string   `xml:"CorporateName"`
	RegistrationData *Address `xml:"RegistrationData>Address,omitempty"`
}

// Address holds a complete postal address including street, post code, town, province, and country code.
type Address struct {
	Address     string `xml:"Address"`
	PostCode    string `xml:"PostCode"`
	Town        string `xml:"Town"`
	Province    string `xml:"Province,omitempty"`
	CountryCode string `xml:"CountryCode"`
}

// Invoices wraps a list of Invoice elements for XML serialization.
type Invoices struct {
	Invoice []Invoice `xml:"Invoice"`
}

// Invoice represents a single invoice within a Facturae document, containing header, issue data,
// tax outputs, totals, and line items.
type Invoice struct {
	InvoiceHeader    InvoiceHeader    `xml:"InvoiceHeader"`
	InvoiceIssueData InvoiceIssueData `xml:"InvoiceIssueData"`
	TaxesOutputs     TaxesOutputs     `xml:"TaxesOutputs"`
	InvoiceTotals    InvoiceTotals    `xml:"InvoiceTotals"`
	Items            Items            `xml:"Items"`
}

// InvoiceHeader contains the identifying information for an invoice: number, series code, document type, and class.
type InvoiceHeader struct {
	InvoiceNumber       string `xml:"InvoiceNumber"`
	InvoiceSeriesCode   string `xml:"InvoiceSeriesCode,omitempty"`
	InvoiceDocumentType string `xml:"InvoiceDocumentType"`
	InvoiceClass        string `xml:"InvoiceClass"`
}

// InvoiceIssueData holds the issue details for an invoice, including the issue date, currencies, and language.
type InvoiceIssueData struct {
	IssueDate           string `xml:"IssueDate"`
	InvoiceCurrencyCode string `xml:"InvoiceCurrencyCode"`
	TaxCurrencyCode     string `xml:"TaxCurrencyCode"`
	LanguageName        string `xml:"LanguageName"`
}

// TaxesOutputs wraps a list of tax output entries summarizing the taxes applied at the invoice level.
type TaxesOutputs struct {
	Tax []TaxOutput `xml:"Tax"`
}

// TaxOutput represents a summary tax entry for a given tax type and rate, with the aggregated taxable base and amount.
type TaxOutput struct {
	TaxTypeCode string `xml:"TaxTypeCode"`
	TaxRate     float64 `xml:"TaxRate"`
	TaxableBase Amount  `xml:"TaxableBase"`
	TaxAmount   Amount  `xml:"TaxAmount"`
}

// InvoiceTotals holds all monetary totals for an invoice, from gross amount through to the executable amount.
type InvoiceTotals struct {
	TotalGrossAmount            float64 `xml:"TotalGrossAmount"`
	TotalGeneralDiscounts       float64 `xml:"TotalGeneralDiscounts,omitempty"`
	TotalGeneralSurcharges      float64 `xml:"TotalGeneralSurcharges,omitempty"`
	TotalGrossAmountBeforeTaxes float64 `xml:"TotalGrossAmountBeforeTaxes"`
	TotalTaxOutputs             float64 `xml:"TotalTaxOutputs"`
	TotalTaxesWithheld          float64 `xml:"TotalTaxesWithheld,omitempty"`
	InvoiceTotal                float64 `xml:"InvoiceTotal"`
	TotalOutstandingAmount      float64 `xml:"TotalOutstandingAmount"`
	TotalExecutableAmount       float64 `xml:"TotalExecutableAmount"`
}

// Items wraps a list of invoice line items for XML serialization.
type Items struct {
	InvoiceLine []InvoiceLine `xml:"InvoiceLine"`
}

// InvoiceLine represents a single line item on an invoice with description, quantity, unit price, and tax breakdown.
type InvoiceLine struct {
	ItemDescription     string    `xml:"ItemDescription"`
	Quantity            float64   `xml:"Quantity"`
	UnitOfMeasure       string    `xml:"UnitOfMeasure,omitempty"`
	UnitPriceWithoutTax float64   `xml:"UnitPriceWithoutTax"`
	TotalCost           float64   `xml:"TotalCost"`
	GrossAmount         float64   `xml:"GrossAmount"`
	TaxesOutputs        LineTaxes `xml:"TaxesOutputs"`
}

// LineTaxes wraps the tax breakdown for an invoice line.
type LineTaxes struct {
	Tax []LineTax `xml:"Tax"`
}

// LineTax represents a single tax applied to an invoice line, specifying the type, rate, taxable base, and amount.
type LineTax struct {
	TaxTypeCode string `xml:"TaxTypeCode"`
	TaxRate     float64 `xml:"TaxRate"`
	TaxableBase Amount  `xml:"TaxableBase"`
	TaxAmount   Amount  `xml:"TaxAmount"`
}
