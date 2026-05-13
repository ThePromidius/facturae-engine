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
	Total               float64
	PreviousFingerprint string
	Fingerprint         string
	Timestamp           time.Time
}

// canonicalize serialises a Record into a pipe-delimited string that serves as
// the input for fingerprint computation.
func canonicalize(r Record) string {
	return fmt.Sprintf("%s|%s|%s|%s|%.2f|%s",
		r.EmisorCIF,
		r.InvoiceSeries,
		r.InvoiceNumber,
		r.IssueDate.Format("2006-01-02"),
		r.Total,
		r.PreviousFingerprint,
	)
}

// fingerprint returns the hex-encoded SHA-256 hash of the canonical string.
func fingerprint(canonical string) string {
	h := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(h[:])
}
