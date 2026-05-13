// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

// Package chain implements a Verifactu-compliant invoice chain that links
// consecutive records via SHA-256 fingerprints. Each record stores a
// cryptographic hash of the previous record, forming a tamper-evident chain.
package chain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// Record represents a single entry in the Verifactu invoice chain. It links
// to the previous record via PreviousFingerprint and carries a Fingerprint
// computed over its own fields.
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

// canonicalize serialises a Record into a pipe-delimited string that serves as
// the input for fingerprint computation according to Art. 13 Orden HAC/1177/2024.
func canonicalize(r Record) string {
	// 1. NIF Emisor
	// 2. NumSerieFactura (concatenado)
	// 3. FechaExpedicionFactura (YYYY-MM-DD)
	// 4. TipoFactura
	// 5. CuotaTotal
	// 6. ImporteTotal
	// 7. Huella anterior
	// 8. FechaHoraHito (ISO 8601 con huso horario)
	return fmt.Sprintf("%s|%s%s|%s|%s|%.2f|%.2f|%s|%s",
		r.EmisorCIF,
		r.InvoiceSeries, r.InvoiceNumber,
		r.IssueDate.Format("2006-01-02"),
		r.InvoiceType,
		r.TaxAmount,
		r.Total,
		r.PreviousFingerprint,
		r.Timestamp.Format(time.RFC3339),
	)
}

// fingerprint returns the hex-encoded SHA-256 hash of the canonical string.
func fingerprint(canonical string) string {
	h := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(h[:])
}
