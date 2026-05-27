package chain

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestSQLiteRoundtripPreservesCanonical(t *testing.T) {
	dbPath := fmt.Sprintf("%s/test_roundtrip_%d.db", os.TempDir(), time.Now().UnixNano())
	defer os.Remove(dbPath)

	store, err := NewSQLStore("sqlite", dbPath)
	if err != nil {
		t.Fatalf("NewSQLStore: %v", err)
	}
	defer store.Close()

	issueDate := time.Date(2026, 5, 13, 0, 0, 0, 0, time.UTC)
	ts := time.Date(2026, 5, 13, 10, 30, 5, 123456789, time.UTC)

	original := Record{
		InvoiceNumber:       "001",
		InvoiceSeries:       "SERIE",
		EmisorCIF:           "A12345678",
		IssueDate:           issueDate,
		InvoiceType:         "F1",
		TaxAmount:           21.00,
		Total:               121.00,
		PreviousFingerprint: "",
		Fingerprint:         "TESTFP",
		Timestamp:           ts,
	}

	wantCanon := canonicalize(original)

	if err := store.Save(original); err != nil {
		t.Fatalf("Save: %v", err)
	}

	all, err := store.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("got %d records, want 1", len(all))
	}

	gotCanon := canonicalize(all[0])

	if gotCanon != wantCanon {
		t.Errorf("canonical string differs after round-trip\nwant: %q\ngot:  %q", wantCanon, gotCanon)
		t.Logf("Original Timestamp: %v (unix=%d, nano=%d)", original.Timestamp, original.Timestamp.Unix(), original.Timestamp.Nanosecond())
		t.Logf("Loaded  Timestamp: %v (unix=%d, nano=%d)", all[0].Timestamp, all[0].Timestamp.Unix(), all[0].Timestamp.Nanosecond())
		t.Logf("Original IssueDate: %v", original.IssueDate)
		t.Logf("Loaded  IssueDate: %v", all[0].IssueDate)
		t.Logf("Original InvoiceType: %q TaxAmount: %v", original.InvoiceType, original.TaxAmount)
		t.Logf("Loaded  InvoiceType: %q TaxAmount: %v", all[0].InvoiceType, all[0].TaxAmount)
	}

	if all[0].Fingerprint != original.Fingerprint {
		t.Errorf("Fingerprint field changed: want %q, got %q", original.Fingerprint, all[0].Fingerprint)
	}
}

func TestSQLiteRoundtripPreservesCanonical_Multiple(t *testing.T) {
	dbPath := fmt.Sprintf("%s/test_roundtrip2_%d.db", os.TempDir(), time.Now().UnixNano())
	defer os.Remove(dbPath)

	store, err := NewSQLStore("sqlite", dbPath)
	if err != nil {
		t.Fatalf("NewSQLStore: %v", err)
	}
	defer store.Close()

	chain := NewChain(store)

	issueDate := time.Date(2026, 5, 13, 0, 0, 0, 0, time.UTC)

	// This approximates what handleInvoice does
	rec, err := chain.Append("001", "SERIE", "A12345678", issueDate, 121.00)
	if err != nil {
		t.Fatalf("Append #1: %v", err)
	}

	// Verify reads back and checks
	if err := chain.Verify(); err != nil {
		t.Fatalf("Verify after append: %v", err)
	}

	// Second invoice
	rec2, err := chain.Append("002", "SERIE", "A12345678", issueDate, 200.00)
	if err != nil {
		t.Fatalf("Append #2: %v", err)
	}

	if err := chain.Verify(); err != nil {
		t.Fatalf("Verify after 2nd append: %v", err)
	}

	t.Logf("Record 1: %s -> %s", rec.InvoiceNumber, rec.Fingerprint[:16])
	t.Logf("Record 2: %s -> %s (prev: %s)", rec2.InvoiceNumber, rec2.Fingerprint[:16], rec2.PreviousFingerprint[:16])
}
