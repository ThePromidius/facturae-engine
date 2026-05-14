// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

// Package chain implements a Verifactu-compliant invoice chain that links
// consecutive records via SHA-256 fingerprints.
package chain

import (
	"fmt"
	"time"
)

// Store is the persistence interface for the invoice chain. Implementations
// may use in-memory storage, PostgreSQL, or SQLite.
type Store interface {
	Save(r Record) error
	Last(emisorCIF string) (Record, bool, error)
	All() ([]Record, error)
	Close() error
}

// Chain is a Verifactu-compliant hash chain that appends invoice records and
// computes chained SHA-256 fingerprints for tamper detection.
type Chain struct {
	store Store
}

// NewChain creates a new Chain backed by the given Store.
func NewChain(s Store) *Chain {
	return &Chain{store: s}
}

// Store returns the underlying storage backend.
func (c *Chain) Store() Store {
	return c.store
}

// Append creates a new chain Record linked to the previous record for the same
// emisorCIF, computes its fingerprint, persists it via the Store, and returns
// it.
func (c *Chain) Append(
	invoiceNumber, invoiceSeries, emisorCIF string,
	issueDate time.Time,
	total float64,
) (Record, error) {
	prev, ok, err := c.store.Last(emisorCIF)
	if err != nil {
		return Record{}, fmt.Errorf("chain: failed to fetch last record: %w", err)
	}

	prevFP := ""
	if ok {
		prevFP = prev.Fingerprint
	}

	r := Record{
		InvoiceNumber:       invoiceNumber,
		InvoiceSeries:       invoiceSeries,
		EmisorCIF:           emisorCIF,
		IssueDate:           issueDate,
		Total:               total,
		PreviousFingerprint: prevFP,
		Timestamp:           time.Now().UTC(),
	}
	r.Fingerprint = fingerprint(canonicalize(r))

	if err := c.store.Save(r); err != nil {
		return Record{}, fmt.Errorf("chain: failed to save record: %w", err)
	}
	return r, nil
}

// Last returns the most recent Record across all emisores, or false if the
// chain is empty.
func (c *Chain) Last() (Record, bool) {
	all, err := c.store.All()
	if err != nil || len(all) == 0 {
		return Record{}, false
	}
	return all[len(all)-1], true
}

// Len returns the total number of records in the chain.
func (c *Chain) Len() int {
	all, err := c.store.All()
	if err != nil {
		return 0
	}
	return len(all)
}

// All returns a slice of all chain records in insertion order.
func (c *Chain) All() []Record {
	all, err := c.store.All()
	if err != nil {
		return nil
	}
	return all
}

// Verify walks the entire chain and checks that every record's
// PreviousFingerprint matches the preceding record's Fingerprint and that each
// record's own Fingerprint is consistent with its canonicalised fields. It
// returns nil if the chain is intact.
func (c *Chain) Verify() error {
	all, err := c.store.All()
	if err != nil {
		return err
	}

	prevFP := ""
	for i, r := range all {
		if r.PreviousFingerprint != prevFP {
			return fmt.Errorf("record %d (%s): previous fingerprint mismatch (chain broken)",
				i, r.InvoiceNumber)
		}
		expected := fingerprint(canonicalize(r))
		if r.Fingerprint != expected {
			return fmt.Errorf("record %d (%s): fingerprint tampered (expected %s, got %s)",
				i, r.InvoiceNumber, expected, r.Fingerprint)
		}
		prevFP = r.Fingerprint
	}
	return nil
}
