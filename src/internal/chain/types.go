// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package chain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// Record represents a single entry in the Verifactu invoice chain.
type Record struct {
	InvoiceNumber       string
	InvoiceSeries       string
	EmisorCIF           string
	IssueDate           time.Time
	InvoiceType         string // F1, F2, R1, R2, R3, R4, R5
	TaxAmount           float64
	Total               float64
	PreviousFingerprint string
	Fingerprint         string
	Timestamp           time.Time
}

// canonicalize serialises a Record into a pipe-delimited string as per Art. 13 Orden HAC/1177/2024.
func canonicalize(r Record) string {
	// 1. NIF Emisor
	// 2. NumSerieFactura (SERIE-NUMBER)
	// 3. FechaExpedicionFactura (DD-MM-YYYY)
	// 4. TipoFactura
	// 5. CuotaTotal (2 decimals)
	// 6. ImporteTotal (2 decimals)
	// 7. Huella anterior (Full 64 chars)
	// 8. FechaHoraHito (ISO 8601 UTC)

	prevFP := r.PreviousFingerprint
	if len(prevFP) > 64 {
		prevFP = prevFP[:64]
	}

	return fmt.Sprintf("%s|%s-%s|%s|%s|%.2f|%.2f|%s|%s",
		r.EmisorCIF,
		r.InvoiceSeries, r.InvoiceNumber,
		r.IssueDate.Format("02-01-2006"), // DD-MM-YYYY
		r.InvoiceType,
		r.TaxAmount,
		r.Total,
		prevFP,
		r.Timestamp.In(time.UTC).Format("2006-01-02T15:04:05Z"), // Strict ISO 8601 UTC
	)
}

// fingerprint returns the hex-encoded SHA-256 hash of the canonical string.
func fingerprint(canonical string) string {
	h := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(h[:])
}
