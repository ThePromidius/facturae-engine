// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package facturae

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// NativeMetadata contains the minimal fields extracted from a native XML file
// needed to perform Verifactu chaining.
type NativeMetadata struct {
	IssuerNIF string
	Number    string
	Series    string
	Date      time.Time
	Total     float64
	Type      string // F1, F2, etc.
}

// ExtractMetadata attempts to parse a FacturaE or UBL XML to extract fields for chaining.
func ExtractMetadata(data []byte) (*NativeMetadata, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	meta := &NativeMetadata{Type: "F1"} // Default

	for {
		t, err := decoder.Token()
		if err != nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			// FacturaE Detection
			if se.Name.Local == "TaxIdentificationNumber" {
				var nif string
				decoder.DecodeElement(&nif, &se)
				if meta.IssuerNIF == "" { meta.IssuerNIF = nif }
			}
			if se.Name.Local == "InvoiceNumber" {
				decoder.DecodeElement(&meta.Number, &se)
			}
			if se.Name.Local == "InvoiceSeriesCode" {
				decoder.DecodeElement(&meta.Series, &se)
			}
			if se.Name.Local == "IssueDate" {
				var dateStr string
				decoder.DecodeElement(&dateStr, &se)
				meta.Date, _ = time.Parse("2006-01-02", dateStr)
			}
			if se.Name.Local == "InvoiceTotal" {
				var totalStr string
				decoder.DecodeElement(&totalStr, &se)
				meta.Total, _ = strconv.ParseFloat(totalStr, 64)
			}

			// UBL Detection (Partial)
			if se.Name.Local == "ID" && meta.Number == "" {
				decoder.DecodeElement(&meta.Number, &se)
			}
		}
	}

	if meta.IssuerNIF == "" || meta.Number == "" {
		return nil, fmt.Errorf("could not extract mandatory Verifactu fields from XML")
	}

	return meta, nil
}

// PatchVerifactu inserts Verifactu specific tags into an existing FacturaE XML.
func PatchVerifactu(xmlData []byte, fingerprint, prevFingerprint string) []byte {
	// For FacturaE, we usually want to insert before the signature or end of Invoice
	// This is a simplified string-based patcher for the POC
	huellaBlock := fmt.Sprintf("\n<Huella>%s</Huella>\n<HuellaAnterior>%s</HuellaAnterior>", fingerprint, prevFingerprint)
	
	insertionPoint := "</Invoice>"
	if strings.Contains(string(xmlData), insertionPoint) {
		return []byte(strings.Replace(string(xmlData), insertionPoint, huellaBlock+insertionPoint, 1))
	}
	
	return xmlData
}
