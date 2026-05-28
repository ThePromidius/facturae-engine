// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package chain_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ThePromidius/facturae-engine/src/internal/chain"
)

// ---- canonicalisation helpers (mirrors chain.canonicalize / chain.fingerprint) ----

func testCanonicalize(r chain.Record) string {
	prevFP := r.PreviousFingerprint
	if len(prevFP) > 64 {
		prevFP = prevFP[:64]
	}
	return fmt.Sprintf("%s|%s-%s|%s|%s|%.2f|%.2f|%s|%s",
		r.EmisorCIF,
		r.InvoiceSeries, r.InvoiceNumber,
		r.IssueDate.Format("02-01-2006"),
		r.InvoiceType,
		r.TaxAmount,
		r.Total,
		prevFP,
		r.Timestamp.In(time.UTC).Format("2006-01-02T15:04:05Z"),
	)
}

func testFingerprint(r chain.Record) string {
	h := sha256.Sum256([]byte(testCanonicalize(r)))
	return hex.EncodeToString(h[:])
}

// ---- custom Store implementations for edge cases ----

// tamperStore wraps a store and modifies the Total on the second record.
type tamperStore struct {
	inner chain.Store
}

func (s *tamperStore) Save(r chain.Record) error                        { return s.inner.Save(r) }
func (s *tamperStore) Last(cif string) (chain.Record, bool, error)      { return s.inner.Last(cif) }
func (s *tamperStore) Close() error                                     { return s.inner.Close() }
func (s *tamperStore) All() ([]chain.Record, error) {
	recs, err := s.inner.All()
	if err != nil {
		return nil, err
	}
	if len(recs) >= 2 {
		recs[1].Total = 99999.99
	}
	return recs, nil
}

// gapStore wraps a store and drops the second record from All() to create a gap.
type gapStore struct {
	inner chain.Store
}

func (s *gapStore) Save(r chain.Record) error                        { return s.inner.Save(r) }
func (s *gapStore) Last(cif string) (chain.Record, bool, error)      { return s.inner.Last(cif) }
func (s *gapStore) Close() error                                     { return s.inner.Close() }
func (s *gapStore) All() ([]chain.Record, error) {
	recs, err := s.inner.All()
	if err != nil {
		return nil, err
	}
	if len(recs) > 1 {
		return append(recs[:1], recs[2:]...), nil
	}
	return recs, nil
}

// ---- helpers ----

func mustAppend(t *testing.T, c *chain.Chain, num, series, cif string, d time.Time, invType string, tax, total float64) chain.Record {
	t.Helper()
	r, err := c.Append(num, series, cif, d, invType, tax, total)
	if err != nil {
		t.Fatalf("Append(%q, %q, %q): %v", num, series, cif, err)
	}
	return r
}

// ---- 1 / 5 / 18. Empty chain and single append ----

func TestEdgeCase_EmptyChain(t *testing.T) {
	store := chain.NewMemoryStore()
	defer store.Close()
	c := chain.NewChain(store)

	if err := c.Verify(); err != nil {
		t.Errorf("empty chain Verify: %v", err)
	}
	if _, ok := c.Last(); ok {
		t.Error("Last on empty: got true, want false")
	}
	if n := c.Len(); n != 0 {
		t.Errorf("Len on empty: got %d, want 0", n)
	}
	if all := c.All(); len(all) != 0 {
		t.Errorf("All on empty: got %d records, want 0", len(all))
	}
}

// ---- 1. Single CIF happy path ----

func TestEdgeCase_SingleAppendAndVerify(t *testing.T) {
	store := chain.NewMemoryStore()
	defer store.Close()
	c := chain.NewChain(store)
	d := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)

	r, err := c.Append("001", "A", "CIF1", d, "F1", 21.00, 121.00)
	if err != nil {
		t.Fatalf("Append: %v", err)
	}
	if err := c.Verify(); err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if r.PreviousFingerprint != "" {
		t.Errorf("PrevFP: got %q, want empty", r.PreviousFingerprint)
	}
	if len(r.Fingerprint) != 64 {
		t.Errorf("fingerprint len: got %d, want 64", len(r.Fingerprint))
	}
	if _, err := hex.DecodeString(r.Fingerprint); err != nil {
		t.Errorf("fingerprint not valid hex: %v", err)
	}
}

// ---- 2. Two invoices same CIF ----

func TestEdgeCase_TwoInvoicesSameCIF(t *testing.T) {
	store := chain.NewMemoryStore()
	defer store.Close()
	c := chain.NewChain(store)
	d := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)

	r1 := mustAppend(t, c, "001", "A", "CIF1", d, "F1", 21.00, 121.00)
	r2 := mustAppend(t, c, "002", "A", "CIF1", d, "F1", 42.00, 242.00)

	if err := c.Verify(); err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if r2.PreviousFingerprint != r1.Fingerprint {
		t.Errorf("PrevFP: got %q, want %q", r2.PreviousFingerprint, r1.Fingerprint)
	}
	if r2.Fingerprint == r1.Fingerprint {
		t.Error("second fingerprint must differ from first")
	}
}

// ---- 3. Two invoices different CIFs (multi-tenant) ----

func TestEdgeCase_TwoInvoicesDifferentCIFs(t *testing.T) {
	store := chain.NewMemoryStore()
	defer store.Close()
	c := chain.NewChain(store)
	d := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)

	r1 := mustAppend(t, c, "001", "A", "CIF1", d, "F1", 10.0, 100.0)
	r2 := mustAppend(t, c, "001", "A", "CIF2", d, "F1", 20.0, 200.0)

	if r1.PreviousFingerprint != "" {
		t.Errorf("CIF1 first PrevFP: got %q, want empty", r1.PreviousFingerprint)
	}
	if r2.PreviousFingerprint != "" {
		t.Errorf("CIF2 first PrevFP: got %q, want empty", r2.PreviousFingerprint)
	}

	r3 := mustAppend(t, c, "002", "A", "CIF1", d, "F1", 15.0, 150.0)
	if r3.PreviousFingerprint != r1.Fingerprint {
		t.Errorf("CIF1 chain broken: got %q, want %q", r3.PreviousFingerprint, r1.Fingerprint)
	}

	r4 := mustAppend(t, c, "002", "A", "CIF2", d, "F1", 25.0, 250.0)
	if r4.PreviousFingerprint != r2.Fingerprint {
		t.Errorf("CIF2 chain broken: got %q, want %q", r4.PreviousFingerprint, r2.Fingerprint)
	}

	if err := c.Verify(); err != nil {
		t.Fatalf("Verify multi-tenant: %v", err)
	}
}

// ---- 4. Tamper detection ----

func TestEdgeCase_TamperDetection(t *testing.T) {
	inner := chain.NewMemoryStore()
	store := &tamperStore{inner: inner}
	c := chain.NewChain(store)
	d := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)

	mustAppend(t, c, "001", "A", "CIF1", d, "F1", 21.00, 121.00)
	mustAppend(t, c, "002", "A", "CIF1", d, "F1", 42.00, 242.00)

	err := c.Verify()
	if err == nil {
		t.Fatal("tampered data: expected error, got nil")
	}
	if !strings.Contains(err.Error(), "tampered") {
		t.Errorf("unexpected error message: %v", err)
	}
	_ = store.Close()
}

// ---- 6. Len ----

func TestEdgeCase_Len(t *testing.T) {
	store := chain.NewMemoryStore()
	defer store.Close()
	c := chain.NewChain(store)
	d := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)

	if n := c.Len(); n != 0 {
		t.Errorf("empty: got %d, want 0", n)
	}
	mustAppend(t, c, "001", "A", "CIF1", d, "F1", 10.0, 100.0)
	if n := c.Len(); n != 1 {
		t.Errorf("after 1: got %d, want 1", n)
	}
	mustAppend(t, c, "002", "A", "CIF1", d, "F1", 20.0, 200.0)
	if n := c.Len(); n != 2 {
		t.Errorf("after 2: got %d, want 2", n)
	}
	mustAppend(t, c, "003", "A", "CIF2", d, "F1", 30.0, 300.0)
	if n := c.Len(); n != 3 {
		t.Errorf("after 3 mixed CIFs: got %d, want 3", n)
	}
}

// ---- 7. Last ----

func TestEdgeCase_Last(t *testing.T) {
	store := chain.NewMemoryStore()
	defer store.Close()
	c := chain.NewChain(store)
	d := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)

	if _, ok := c.Last(); ok {
		t.Error("Last on empty: got true, want false")
	}

	mustAppend(t, c, "001", "A", "CIF1", d, "F1", 10.0, 100.0)
	mustAppend(t, c, "002", "A", "CIF2", d, "F1", 20.0, 200.0)
	mustAppend(t, c, "003", "A", "CIF1", d, "F1", 15.0, 150.0)

	last, ok := c.Last()
	if !ok {
		t.Fatal("Last: got false, want true")
	}
	if last.InvoiceNumber != "003" {
		t.Errorf("InvoiceNumber: got %q, want %q", last.InvoiceNumber, "003")
	}
	if last.EmisorCIF != "CIF1" {
		t.Errorf("EmisorCIF: got %q, want %q", last.EmisorCIF, "CIF1")
	}
}

// ---- 8. All ----

func TestEdgeCase_All(t *testing.T) {
	store := chain.NewMemoryStore()
	defer store.Close()
	c := chain.NewChain(store)
	d := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)

	if all := c.All(); len(all) != 0 {
		t.Errorf("empty: got %d, want 0", len(all))
	}

	mustAppend(t, c, "001", "A", "CIF1", d, "F1", 10.0, 100.0)
	mustAppend(t, c, "002", "A", "CIF2", d, "F1", 20.0, 200.0)
	mustAppend(t, c, "003", "A", "CIF1", d, "F1", 15.0, 150.0)

	all := c.All()
	if len(all) != 3 {
		t.Fatalf("All: got %d, want 3", len(all))
	}
	if all[0].InvoiceNumber != "001" {
		t.Errorf("All[0]: got %q, want %q", all[0].InvoiceNumber, "001")
	}
	if all[1].InvoiceNumber != "002" {
		t.Errorf("All[1]: got %q, want %q", all[1].InvoiceNumber, "002")
	}
	if all[2].InvoiceNumber != "003" {
		t.Errorf("All[2]: got %q, want %q", all[2].InvoiceNumber, "003")
	}

	// Verify insertion order maintained
	for i := 1; i < len(all); i++ {
		if all[i].Timestamp.Before(all[i-1].Timestamp) {
			t.Errorf("All[%d] Timestamp before All[%d]", i, i-1)
		}
	}
}

// ---- 9 & 10. PreviousFingerprint ----

func TestEdgeCase_PreviousFingerprint(t *testing.T) {
	store := chain.NewMemoryStore()
	defer store.Close()
	c := chain.NewChain(store)
	d := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)

	r1 := mustAppend(t, c, "001", "A", "CIF1", d, "F1", 10.0, 100.0)
	if r1.PreviousFingerprint != "" {
		t.Errorf("Test 9: first PrevFP = %q, want empty", r1.PreviousFingerprint)
	}

	r2 := mustAppend(t, c, "002", "A", "CIF1", d, "F1", 20.0, 200.0)
	if r2.PreviousFingerprint != r1.Fingerprint {
		t.Errorf("Test 10: second PrevFP = %q, want %q", r2.PreviousFingerprint, r1.Fingerprint)
	}

	// New CIF starts fresh chain
	r3 := mustAppend(t, c, "001", "A", "CIF2", d, "F1", 30.0, 300.0)
	if r3.PreviousFingerprint != "" {
		t.Errorf("Test 9: new CIF PrevFP = %q, want empty", r3.PreviousFingerprint)
	}
}

// ---- 11. Multiple appends (5+) ----

func TestEdgeCase_MultipleAppends(t *testing.T) {
	store := chain.NewMemoryStore()
	defer store.Close()
	c := chain.NewChain(store)
	d := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)

	var prev chain.Record
	const count = 7

	for i := 1; i <= count; i++ {
		num := fmt.Sprintf("%03d", i)
		total := float64(i * 100)
		tax := float64(i * 21)
		r := mustAppend(t, c, num, "A", "CIF1", d, "F1", tax, total)

		if i == 1 {
			if r.PreviousFingerprint != "" {
				t.Errorf("record %d: PrevFP = %q, want empty", i, r.PreviousFingerprint)
			}
		} else {
			if r.PreviousFingerprint != prev.Fingerprint {
				t.Errorf("record %d: PrevFP mismatch:\n  got:  %q\n  want: %q",
					i, r.PreviousFingerprint, prev.Fingerprint)
			}
		}
		if i > 1 && r.Fingerprint == prev.Fingerprint {
			t.Errorf("record %d: fingerprint unchanged from previous", i)
		}
		prev = r
	}

	if err := c.Verify(); err != nil {
		t.Fatalf("Verify after %d appends: %v", count, err)
	}
	if n := c.Len(); n != count {
		t.Errorf("Len: got %d, want %d", n, count)
	}
}

// ---- 12. Extreme values ----

func TestEdgeCase_ExtremeValues(t *testing.T) {
	store := chain.NewMemoryStore()
	defer store.Close()
	c := chain.NewChain(store)

	type extremeCase struct {
		name        string
		num         string
		series      string
		cif         string
		date        time.Time
		invType     string
		tax, total  float64
	}

	cases := []extremeCase{
		{"empty invoice number",  "",          "A", "CIF1", time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC), "F1", 10.0, 100.0},
		{"future date",           "002",       "A", "CIF1", time.Date(2099, 12, 31, 0, 0, 0, 0, time.UTC), "F1", 10.0, 100.0},
		{"zero total",            "003",       "A", "CIF1", time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC), "F1", 0.0, 0.0},
		{"negative total",        "004",       "A", "CIF1", time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC), "R1", -5.0, -100.0},
		{"special chars",         "INV/001-A_B", "A", "CIF1", time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC), "F1", 10.0, 100.0},
		{"empty series",          "006",       "",  "CIF1", time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC), "F1", 10.0, 100.0},
		{"zero tax",              "007",       "A", "CIF1", time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC), "F1", 0.0, 100.0},
		{"large values",          "008",       "A", "CIF1", time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC), "F1", 1e10, 1e12},
		{"minimal valid values",  "009",       "A", "CIF1", time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC), "",  0.0, 0.01},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := mustAppend(t, c, tc.num, tc.series, tc.cif, tc.date, tc.invType, tc.tax, tc.total)
			if len(r.Fingerprint) != 64 {
				t.Errorf("fingerprint length: got %d, want 64", len(r.Fingerprint))
			}
			if r.IssueDate.Format("02-01-2006") != tc.date.Format("02-01-2006") {
				t.Errorf("IssueDate: got %v, want %v", r.IssueDate, tc.date)
			}
		})
	}

	if err := c.Verify(); err != nil {
		t.Fatalf("Verify after extreme values: %v", err)
	}
	if n := c.Len(); n != len(cases) {
		t.Errorf("Len: got %d, want %d", n, len(cases))
	}
}

// ---- 13. Gap in chain ----

func TestEdgeCase_GapInChain(t *testing.T) {
	inner := chain.NewMemoryStore()
	store := &gapStore{inner: inner}
	c := chain.NewChain(store)
	d := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)

	mustAppend(t, c, "001", "A", "CIF1", d, "F1", 10.0, 100.0)
	mustAppend(t, c, "002", "A", "CIF1", d, "F1", 20.0, 200.0)
	mustAppend(t, c, "003", "A", "CIF1", d, "F1", 30.0, 300.0)

	err := c.Verify()
	if err == nil {
		t.Fatal("gap in chain: expected error, got nil")
	}
	if !strings.Contains(err.Error(), "broken") && !strings.Contains(err.Error(), "mismatch") {
		t.Errorf("unexpected error: %v", err)
	}
	_ = store.Close()
}

// ---- 14. Concurrent appends ----

func TestEdgeCase_ConcurrentAppends(t *testing.T) {
	store := chain.NewMemoryStore()
	defer store.Close()
	c := chain.NewChain(store)
	d := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)

	var wg sync.WaitGroup
	const numCIFs = 5
	const perCIF = 3

	for i := range numCIFs {
		wg.Add(1)
		cif := fmt.Sprintf("CIF-CONC-%d", i)
		go func(cifID string) {
			defer wg.Done()
			for j := range perCIF {
				num := fmt.Sprintf("%s-%03d", cifID, j+1)
				if _, err := c.Append(num, "A", cifID, d, "F1", 10.0, 100.0); err != nil {
					t.Errorf("concurrent Append(%s): %v", num, err)
				}
			}
		}(cif)
	}
	wg.Wait()

	if n := c.Len(); n != numCIFs*perCIF {
		t.Errorf("Len: got %d, want %d", n, numCIFs*perCIF)
	}
	if err := c.Verify(); err != nil {
		t.Fatalf("Verify after concurrent appends: %v", err)
	}

	// Each CIF has exactly perCIF records in order
	all := c.All()
	cifCounts := make(map[string]int)
	for _, r := range all {
		cifCounts[r.EmisorCIF]++
	}
	for cif, count := range cifCounts {
		if count != perCIF {
			t.Errorf("CIF %s: got %d records, want %d", cif, count, perCIF)
		}
	}

	// Verify each per-CIF chain links correctly
	cifChains := make(map[string][]chain.Record)
	for _, r := range all {
		cifChains[r.EmisorCIF] = append(cifChains[r.EmisorCIF], r)
	}
	for cif, recs := range cifChains {
		prevFP := ""
		for i, r := range recs {
			if r.PreviousFingerprint != prevFP {
				t.Errorf("CIF %s record %d: PrevFP mismatch", cif, i)
			}
			prevFP = r.Fingerprint
		}
	}
}

// ---- 15. Record fields populated ----

func TestEdgeCase_RecordFieldsPopulated(t *testing.T) {
	store := chain.NewMemoryStore()
	defer store.Close()
	c := chain.NewChain(store)
	d := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)

	r, err := c.Append("INV-100", "SERIE-A", "B87654321", d, "R1", 15.50, 115.50)
	if err != nil {
		t.Fatalf("Append: %v", err)
	}

	if r.InvoiceNumber != "INV-100" {
		t.Errorf("InvoiceNumber: got %q, want %q", r.InvoiceNumber, "INV-100")
	}
	if r.InvoiceSeries != "SERIE-A" {
		t.Errorf("InvoiceSeries: got %q, want %q", r.InvoiceSeries, "SERIE-A")
	}
	if r.EmisorCIF != "B87654321" {
		t.Errorf("EmisorCIF: got %q, want %q", r.EmisorCIF, "B87654321")
	}
	if !r.IssueDate.Equal(d) {
		t.Errorf("IssueDate: got %v, want %v", r.IssueDate, d)
	}
	if r.InvoiceType != "R1" {
		t.Errorf("InvoiceType: got %q, want %q", r.InvoiceType, "R1")
	}
	if r.TaxAmount != 15.50 {
		t.Errorf("TaxAmount: got %v, want %v", r.TaxAmount, 15.50)
	}
	if r.Total != 115.50 {
		t.Errorf("Total: got %v, want %v", r.Total, 115.50)
	}
	if r.PreviousFingerprint != "" {
		t.Errorf("PreviousFingerprint: got %q, want empty", r.PreviousFingerprint)
	}
	if len(r.Fingerprint) != 64 {
		t.Errorf("Fingerprint length: got %d, want 64", len(r.Fingerprint))
	}
	if r.Timestamp.IsZero() {
		t.Error("Timestamp is zero")
	}
	if time.Since(r.Timestamp) > 10*time.Second {
		t.Errorf("Timestamp too old: %v ago", time.Since(r.Timestamp))
	}
}

// ---- 16. Fingerprint deterministic ----

func TestEdgeCase_FingerprintDeterministic(t *testing.T) {
	store := chain.NewMemoryStore()
	defer store.Close()
	c := chain.NewChain(store)
	d := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)

	r := mustAppend(t, c, "001", "A", "CIF1", d, "F1", 21.00, 121.00)

	// Recompute using our reference implementation
	expected := testFingerprint(r)
	if r.Fingerprint != expected {
		t.Errorf("fingerprint mismatch:\n  got:  %q\n  want: %q", r.Fingerprint, expected)
	}

	// Verify the canonical string is also deterministic
	c1 := testCanonicalize(r)
	c2 := testCanonicalize(r)
	if c1 != c2 {
		t.Error("canonicalize not deterministic for identical record")
	}

	// Verify chain matches with our reference
	if err := c.Verify(); err != nil {
		t.Fatalf("Verify: %v", err)
	}
}

// ---- 17. Fingerprint changes when any field changes ----

func TestEdgeCase_FingerprintChanges(t *testing.T) {
	store := chain.NewMemoryStore()
	defer store.Close()
	c := chain.NewChain(store)
	d := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)

	// Baseline
	r0 := mustAppend(t, c, "001", "A", "CIF1", d, "F1", 21.00, 121.00)

	// For each variant we use a different CIF so chain dependency doesn't interfere
	variants := []struct {
		name string
		run  func() chain.Record
	}{
		{"different InvoiceNumber", func() chain.Record {
			return mustAppend(t, c, "XXX", "A", "CIF-VN", d, "F1", 21.00, 121.00)
		}},
		{"different InvoiceSeries", func() chain.Record {
			return mustAppend(t, c, "001", "ZZZ", "CIF-VS", d, "F1", 21.00, 121.00)
		}},
		{"different EmisorCIF", func() chain.Record {
			return mustAppend(t, c, "001", "A", "OTHER-CIF", d, "F1", 21.00, 121.00)
		}},
		{"different IssueDate", func() chain.Record {
			return mustAppend(t, c, "001", "A", "CIF-VD",
				time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC), "F1", 21.00, 121.00)
		}},
		{"different InvoiceType", func() chain.Record {
			return mustAppend(t, c, "001", "A", "CIF-VT", d, "R2", 21.00, 121.00)
		}},
		{"different TaxAmount", func() chain.Record {
			return mustAppend(t, c, "001", "A", "CIF-VA", d, "F1", 99.99, 121.00)
		}},
		{"different Total", func() chain.Record {
			return mustAppend(t, c, "001", "A", "CIF-VL", d, "F1", 21.00, 999.99)
		}},
	}

	fingerprints := make(map[string]string)
	for _, v := range variants {
		r := v.run()
		fingerprints[v.name] = r.Fingerprint
	}

	if err := c.Verify(); err != nil {
		t.Fatalf("Verify: %v", err)
	}

	// All fingerprints must differ from baseline and from each other
	seen := make(map[string]string)
	for name, fp := range fingerprints {
		if fp == r0.Fingerprint {
			t.Errorf("%s: fingerprint equals baseline", name)
		}
		if prev, dup := seen[fp]; dup {
			t.Errorf("%s and %s: identical fingerprint", name, prev)
		}
		seen[fp] = name
	}
}

// ---- 19. Chain length with mixed CIFs ----

func TestEdgeCase_ChainLengthMixedCIFs(t *testing.T) {
	store := chain.NewMemoryStore()
	defer store.Close()
	c := chain.NewChain(store)
	d := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)

	mustAppend(t, c, "001", "A", "CIF1", d, "F1", 10.0, 100.0)
	mustAppend(t, c, "002", "A", "CIF2", d, "F1", 20.0, 200.0)
	mustAppend(t, c, "003", "A", "CIF1", d, "F1", 15.0, 150.0)
	mustAppend(t, c, "004", "A", "CIF3", d, "F1", 25.0, 250.0)
	mustAppend(t, c, "005", "A", "CIF1", d, "F1", 30.0, 300.0)

	if n := c.Len(); n != 5 {
		t.Errorf("Len: got %d, want 5 (all CIFs)", n)
	}
	if all := c.All(); len(all) != 5 {
		t.Errorf("All: got %d, want 5", len(all))
	}
	if err := c.Verify(); err != nil {
		t.Fatalf("Verify: %v", err)
	}
}

// ---- 20. Canonicalize produces consistent output ----

func TestEdgeCase_CanonicalizeConsistent(t *testing.T) {
	store := chain.NewMemoryStore()
	defer store.Close()
	c := chain.NewChain(store)
	d := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)

	r := mustAppend(t, c, "001", "A", "CIF1", d, "F1", 21.00, 121.00)

	canon := testCanonicalize(r)
	if len(canon) == 0 {
		t.Fatal("empty canonical string")
	}

	// Check structural elements: pipe-separated
	parts := strings.Split(canon, "|")
	if len(parts) != 8 {
		t.Errorf("canonical has %d pipe-delimited parts, want 8", len(parts))
	}

	// Parts breakdown: CIF | SERIES-NUM | DD-MM-YYYY | TYPE | TAX | TOTAL | PREVFP | TS
	if len(parts) >= 3 {
		expectedDate := d.Format("02-01-2006")
		if parts[2] != expectedDate {
			t.Errorf("date part: got %q, want %q", parts[2], expectedDate)
		}
	}
	if len(parts) >= 5 {
		expectedTax := fmt.Sprintf("%.2f", 21.00)
		if parts[4] != expectedTax {
			t.Errorf("tax part: got %q, want %q", parts[4], expectedTax)
		}
	}

	// Consistency: same record produces same canonical output
	rCopy := r
	if testCanonicalize(rCopy) != canon {
		t.Error("identical records produce different canonical strings")
	}

	// Verify chain matches our canonicalization
	ourFP := testFingerprint(r)
	if r.Fingerprint != ourFP {
		t.Errorf("fingerprint from canonical: got %q, want %q", ourFP, r.Fingerprint)
	}

	// Save directly via store and verify roundtrip
	all, err := store.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(all) > 0 {
		rt := testCanonicalize(all[0])
		if rt != canon {
			t.Errorf("canonical changes after store roundtrip:\n  before: %q\n  after:  %q", canon, rt)
		}
	}

	// Verify PrevFP truncation: create a record with long PrevFP
	t.Run("prevFP truncation", func(t *testing.T) {
		// Use 70 chars where the last 6 are a unique suffix not found in the first 64
		longFP := strings.Repeat("A", 64) + "!!UNIQ"
		truncR := chain.Record{
			InvoiceNumber:       "TRUNC",
			InvoiceSeries:       "A",
			EmisorCIF:           "CIF1",
			IssueDate:           d,
			InvoiceType:         "F1",
			TaxAmount:           10.0,
			Total:               100.0,
			PreviousFingerprint: longFP,
			Timestamp:           time.Date(2026, 5, 15, 10, 0, 0, 0, time.UTC),
		}
		canon := testCanonicalize(truncR)
		if len(canon) == 0 {
			t.Fatal("empty canonical for truncation record")
		}
		// Truncated PrevFP (first 64 of 70) should be in canonical output
		if !strings.Contains(canon, longFP[:64]) {
			t.Error("canonical should contain first 64 chars of PrevFP")
		}
		// Full 70-char PrevFP suffix should NOT be in canonical output
		if strings.Contains(canon, "!!UNIQ") {
			t.Error("canonical should NOT contain chars beyond 64 of PrevFP")
		}
	})
}
