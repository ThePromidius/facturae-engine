// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

// Package chain implements a Verifactu-compliant invoice chain that links
// consecutive records via SHA-256 fingerprints.
package chain

import "sync"

// MemoryStore is an in-memory implementation of the Store interface, safe for
// concurrent use via a read-write mutex.
type MemoryStore struct {
	mu      sync.RWMutex
	records []Record
}

// NewMemoryStore returns an initialised empty in-memory chain store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

// Save appends a record to the in-memory slice, safe for concurrent access.
func (s *MemoryStore) Save(r Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records = append(s.records, r)
	return nil
}

// Last returns the most recent Record for the given emisorCIF, scanning from
// the end of the in-memory slice.
func (s *MemoryStore) Last(emisorCIF string) (Record, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := len(s.records) - 1; i >= 0; i-- {
		if s.records[i].EmisorCIF == emisorCIF {
			return s.records[i], true, nil
		}
	}
	return Record{}, false, nil
}

// All returns a copy of all records in insertion order.
func (s *MemoryStore) All() ([]Record, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Record, len(s.records))
	copy(out, s.records)
	return out, nil
}

// Close is a no-op for the in-memory store; it always returns nil.
func (s *MemoryStore) Close() error { return nil }
