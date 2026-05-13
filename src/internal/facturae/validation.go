// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package facturae

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
)

// StructuralError describes a single validation failure within a FacturaE struct,
// identifying the field path and a human-readable message.
type StructuralError struct {
	Field   string
	Message string
}

func (e *StructuralError) Error() string {
	return fmt.Sprintf("structural error at %s: %s", e.Field, e.Message)
}

// StructuralErrors is a collection of StructuralError values that itself implements the error interface.
type StructuralErrors []*StructuralError

func (se StructuralErrors) Error() string {
	msgs := make([]string, len(se))
	for i, e := range se {
		msgs[i] = e.Error()
	}
	return strings.Join(msgs, "; ")
}

// ValidateStruct performs structural validation on a FacturaE struct, checking that all required
// fields are present, invoice totals are consistent, and business rules are satisfied.
func ValidateStruct(f *FacturaE) error {
	var errs StructuralErrors

	check := func(condition bool, field, msg string) {
		if !condition {
			errs = append(errs, &StructuralError{Field: field, Message: msg})
		}
	}

	check(f.FileHeader.SchemaVersion != "", "FileHeader.SchemaVersion", "required")
	check(f.FileHeader.Modality != "", "FileHeader.Modality", "required")
	check(f.FileHeader.Batch.InvoicesCount > 0, "FileHeader.Batch.InvoicesCount", "must be > 0")
	check(f.FileHeader.Batch.TotalInvoicesAmount.TotalAmount > 0, "FileHeader.Batch.TotalInvoicesAmount", "must be > 0")

	check(f.Parties.SellerParty.TaxIdentification.TaxIdentificationNumber != "",
		"Parties.SellerParty.TaxIdentification.TaxIdentificationNumber", "required")
	check(f.Parties.BuyerParty.TaxIdentification.TaxIdentificationNumber != "",
		"Parties.BuyerParty.TaxIdentification.TaxIdentificationNumber", "required")

	check(len(f.Invoices.Invoice) > 0, "Invoices", "must contain at least one Invoice")

	for i, inv := range f.Invoices.Invoice {
		prefix := fmt.Sprintf("Invoices.Invoice[%d]", i)

		check(inv.InvoiceHeader.InvoiceNumber != "", prefix+".InvoiceHeader.InvoiceNumber", "required")
		check(inv.InvoiceIssueData.IssueDate != "", prefix+".InvoiceIssueData.IssueDate", "required")
		check(len(inv.Items.InvoiceLine) > 0, prefix+".Items", "must have at least one line")

		check(inv.InvoiceTotals.InvoiceTotal >= inv.InvoiceTotals.TotalGrossAmount,
			prefix+".InvoiceTotals.InvoiceTotal",
			"must be >= TotalGrossAmount (taxes must add, not subtract)")

		check(inv.InvoiceTotals.InvoiceTotal > 0, prefix+".InvoiceTotals.InvoiceTotal", "must be > 0")

		for j, line := range inv.Items.InvoiceLine {
			lprefix := fmt.Sprintf("%s.Items.InvoiceLine[%d]", prefix, j)
			check(line.ItemDescription != "", lprefix+".ItemDescription", "required")
			check(line.Quantity > 0, lprefix+".Quantity", "must be > 0")
			check(line.GrossAmount >= 0, lprefix+".GrossAmount", "cannot be negative")
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// ValidateXMLBytes performs basic validation on serialized Facturae XML data, checking minimum length,
// well-formedness, root element name, and the presence of the required xmlns:fe namespace.
func ValidateXMLBytes(xmlData []byte) error {
	if len(xmlData) < 500 {
		return errors.New("XML suspiciously short (< 500 bytes)")
	}

	dec := xml.NewDecoder(strings.NewReader(string(xmlData)))
	var root xml.StartElement
	foundRoot := false

	for {
		tok, err := dec.Token()
		if err != nil {
			return fmt.Errorf("XML parse error: %w", err)
		}
		if se, ok := tok.(xml.StartElement); ok {
			root = se
			foundRoot = true
			break
		}
	}

	if !foundRoot {
		return errors.New("no root element found in XML")
	}

	localName := root.Name.Local
	if localName != "Facturae" {
		return fmt.Errorf("unexpected root element %q, expected Facturae", localName)
	}

	feNS := ""
	for _, attr := range root.Attr {
		if attr.Name.Local == "fe" && attr.Name.Space == "xmlns" {
			feNS = attr.Value
		}
	}
	if feNS == "" {
		return errors.New("missing xmlns:fe namespace declaration on root element")
	}

	return nil
}
