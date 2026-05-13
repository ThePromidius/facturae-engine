// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package chain_test

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ThePromidius/facturae-engine/src/internal/chain"
)

var baseDate = time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)

func TestChain_FirstRecordHasNoParent(t *testing.T) {
	c := chain.NewChain(chain.NewMemoryStore())
	r, err := c.Append("INV-001", "F", "B12345678", baseDate, 1000.0)
	if err != nil {
		t.Fatalf("Append failed: %v", err)
	}
	if r.PreviousFingerprint != "" {
		t.Errorf("first record should have empty PreviousFingerprint, got %q", r.PreviousFingerprint)
	}
	if r.Fingerprint == "" {
		t.Error("first record should have a non-empty Fingerprint")
	}
}

func TestChain_SubsequentRecordsChain(t *testing.T) {
	c := chain.NewChain(chain.NewMemoryStore())
	r1, _ := c.Append("INV-001", "F", "B12345678", baseDate, 100.0)
	r2, _ := c.Append("INV-002", "F", "B12345678", baseDate.AddDate(0, 0, 1), 200.0)

	if r2.PreviousFingerprint != r1.Fingerprint {
		t.Errorf("r2.PreviousFingerprint should equal r1.Fingerprint\n  r1.FP=%s\n  r2.PrevFP=%s",
			r1.Fingerprint, r2.PreviousFingerprint)
	}
}

func TestChain_FingerprintIsDeterministic(t *testing.T) {
	c1 := chain.NewChain(chain.NewMemoryStore())
	c2 := chain.NewChain(chain.NewMemoryStore())

	r1, _ := c1.Append("INV-001", "F", "B12345678", baseDate, 100.0)
	r2, _ := c2.Append("INV-001", "F", "B12345678", baseDate, 100.0)

	if r1.Fingerprint != r2.Fingerprint {
		t.Errorf("same inputs should produce same fingerprint:\n  c1=%s\n  c2=%s",
			r1.Fingerprint, r2.Fingerprint)
	}
}

func TestChain_FingerprintChangesOnDifferentTotal(t *testing.T) {
	c := chain.NewChain(chain.NewMemoryStore())
	r1, _ := c.Append("INV-001", "F", "B12345678", baseDate, 100.0)

	c2 := chain.NewChain(chain.NewMemoryStore())
	r2, _ := c2.Append("INV-001", "F", "B12345678", baseDate, 200.0)

	if r1.Fingerprint == r2.Fingerprint {
		t.Error("different totals should produce different fingerprints")
	}
}

func TestChain_Verify_CleanChain(t *testing.T) {
	c := chain.NewChain(chain.NewMemoryStore())
	for i := 1; i <= 5; i++ {
		c.Append(
			strings.Repeat("0", 4-len(string(rune('0'+i))))+string(rune('0'+i)),
			"F", "B12345678",
			baseDate.AddDate(0, 0, i-1),
			float64(i*100),
		)
	}
	if err := c.Verify(); err != nil {
		t.Fatalf("Verify() failed on clean chain: %v", err)
	}
}

func TestChain_Len(t *testing.T) {
	c := chain.NewChain(chain.NewMemoryStore())
	if c.Len() != 0 {
		t.Error("empty chain should have Len 0")
	}
	c.Append("INV-001", "F", "B12345678", baseDate, 100.0)
	c.Append("INV-002", "F", "B12345678", baseDate, 200.0)
	if c.Len() != 2 {
		t.Errorf("expected Len 2, got %d", c.Len())
	}
}

func TestChain_Last_EmptyChain(t *testing.T) {
	c := chain.NewChain(chain.NewMemoryStore())
	_, ok := c.Last()
	if ok {
		t.Error("Last() on empty chain should return false")
	}
}

func TestChain_All_ReturnsCopy(t *testing.T) {
	c := chain.NewChain(chain.NewMemoryStore())
	c.Append("INV-001", "F", "B12345678", baseDate, 100.0)
	all := c.All()
	all[0].Fingerprint = "tampered"

	last, _ := c.Last()
	if last.Fingerprint == "tampered" {
		t.Error("All() should return a copy, not references to internal state")
	}
}

func TestChain_ConcurrentAppend(t *testing.T) {
	c := chain.NewChain(chain.NewMemoryStore())
	const workers = 50

	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		i := i
		go func() {
			defer wg.Done()
			c.Append(
				strings.Repeat("I", i%5+1),
				"F",
				"B12345678",
				baseDate.AddDate(0, 0, i),
				float64(i+1)*10,
			)
		}()
	}
	wg.Wait()

	if c.Len() != workers {
		t.Errorf("expected %d records after concurrent append, got %d", workers, c.Len())
	}
}

func TestChain_FingerprintIsHex64Chars(t *testing.T) {
	c := chain.NewChain(chain.NewMemoryStore())
	r, _ := c.Append("INV-001", "F", "B12345678", baseDate, 100.0)
	if len(r.Fingerprint) != 64 {
		t.Errorf("SHA-256 hex fingerprint should be 64 chars, got %d", len(r.Fingerprint))
	}
}
