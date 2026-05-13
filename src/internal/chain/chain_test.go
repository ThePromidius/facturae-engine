package chain

import (
	"testing"
	"time"
)

func TestCanonicalize(t *testing.T) {
	issueDate := time.Date(2026, 5, 13, 0, 0, 0, 0, time.UTC)
	timestamp := time.Date(2026, 5, 13, 10, 30, 0, 0, time.UTC)

	record := Record{
		EmisorCIF:           "A12345678",
		InvoiceSeries:       "SERIE",
		InvoiceNumber:       "123",
		IssueDate:           issueDate,
		InvoiceType:         "F1",
		TaxAmount:           21.00,
		Total:               121.00,
		PreviousFingerprint: "PREV-HASH",
		Timestamp:           timestamp,
	}

	got := canonicalize(record)
	// New expected order based on latest legal mastermind: 
	// EmisorCIF|Series-Number|DD-MM-YYYY|Type|Tax|Total|PrevHash|ISO8601Z
	want := "A12345678|SERIE-123|13-05-2026|F1|21.00|121.00|PREV-HASH|2026-05-13T10:30:00Z"

	if got != want {
		t.Errorf("canonicalize() = %v, want %v", got, want)
	}
}

func TestFingerprint(t *testing.T) {
	input := "test-string"
	// echo -n "test-string" | sha256sum
	want := "ffe65f1d98fafedea3514adc956c8ada5980c6c5d2552fd61f48401aefd5c00e"
	got := fingerprint(input)

	if got != want {
		t.Errorf("fingerprint() = %v, want %v", got, want)
	}
}
