package ubl

import (
	"encoding/xml"
	"fmt"

	"github.com/ThePromidius/facturae-engine/src/internal/invoice"
)

// Invoice represents a simplified UBL 2.1 Invoice structure for Crea y Crece compliance.
type Invoice struct {
	XMLName           xml.Name `xml:"ubl:Invoice"`
	XmlnsUbl          string   `xml:"xmlns:ubl,attr"`
	XmlnsCac          string   `xml:"xmlns:cac,attr"`
	XmlnsCbc          string   `xml:"xmlns:cbc,attr"`
	ID                string   `xml:"cbc:ID"`
	IssueDate         string   `xml:"cbc:IssueDate"`
	InvoiceTypeCode   string   `xml:"cbc:InvoiceTypeCode"`
	DocumentCurrency  string   `xml:"cbc:DocumentCurrencyCode"`
	
	// Parties, Lines, Totals omitted for brevity in this architectural stub
}

// Builder constructs UBL XML from internal invoice requests.
type Builder struct{}

func (b *Builder) Build(req invoice.Request) ([]byte, error) {
	// 1. Map internal request to UBL structure
	inv := Invoice{
		XmlnsUbl: "urn:oasis:names:specification:ubl:schema:xsd:Invoice-2",
		XmlnsCac: "urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2",
		XmlnsCbc: "urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2",
		ID:       req.Factura.Serie + "-" + req.Factura.Numero,
		IssueDate: req.Factura.Fecha.Format("2006-01-02"),
		InvoiceTypeCode: "380", // Standard Commercial Invoice
		DocumentCurrency: req.Meta.Moneda,
	}

	// 2. Marshal to XML
	output, err := xml.MarshalIndent(inv, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("ubl: failed to marshal xml: %w", err)
	}

	return append([]byte(xml.Header), output...), nil
}
