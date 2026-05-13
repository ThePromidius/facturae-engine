package chain

import (
	"testing"
	"time"
)

func TestCanonicalize(t *testing.T) {
	issueDate := time.Date(2026, 5, 13, 0, 0, 0, 0, time.UTC)
	timestamp := time.Date(2026, 5, 13, 10, 30, 0, 0, time.FixedZone("CEST", 2*60*60))

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
	// Expected order: EmisorCIF|Series+Number|IssueDate|Type|Tax|Total|PrevHash|Timestamp
	want := "A12345678|SERIE123|2026-05-13|F1|21.00|121.00|PREV-HASH|2026-05-13T10:30:00+02:00"

	if got != want {
		t.Errorf("canonicalize() = %v, want %v", got, want)
	}
}

func TestFingerprint(t *testing.T) {
	input := "test-string"
	// echo -n "test-string" | sha256sum
	want := "d5558e7090382ba187f55180f680970a27320f269600e0086706f9d7840134f5"
	got := fingerprint(input)

	if got != want {
		t.Errorf("fingerprint() = %v, want %v", got, want)
	}
}
