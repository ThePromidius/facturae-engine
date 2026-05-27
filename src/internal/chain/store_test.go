package chain

import (
	"fmt"
	"os"
	"testing"
	"time"
)

// storeFactory is a function that creates a fresh Store for testing.
// It returns the store, a cleanup function, and any error.
type storeFactory func(name string) (Store, func(), error)

func memoryFactory(_ string) (Store, func(), error) {
	return NewMemoryStore(), func() {}, nil
}

func sqliteFactory(name string) (Store, func(), error) {
	path := fmt.Sprintf("%s%s_test_%d.db", os.TempDir(), string(os.PathSeparator), time.Now().UnixNano())
	store, err := NewSQLStore("sqlite", path)
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() {
		store.Close()
		os.Remove(path)
	}
	return store, cleanup, nil
}

func postgresFactory(name string) (Store, func(), error) {
	host := os.Getenv("PGHOST")
	port := os.Getenv("PGPORT")
	user := os.Getenv("PGUSER")
	password := os.Getenv("PGPASSWORD")
	dbname := os.Getenv("PGDATABASE")

	if port == "" {
		port = "5432"
	}
	if user == "" {
		user = "postgres"
	}
	if password == "" {
		password = "postgres"
	}
	if dbname == "" {
		dbname = "facturae_test"
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)
	store, err := NewSQLStore("postgres", dsn)
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() {
		store.Close()
	}
	return store, cleanup, nil
}

// stores returns the map of available store factories. PostgreSQL is only
// included when PGHOST is set (e.g. inside Docker or CI).
func stores() map[string]storeFactory {
	m := map[string]storeFactory{
		"memory": memoryFactory,
		"sqlite": sqliteFactory,
	}
	if os.Getenv("PGHOST") != "" {
		m["postgres"] = postgresFactory
	}
	return m
}

func TestStores_AppendAndVerify(t *testing.T) {
	factories := stores()

	for name, factory := range factories {
		t.Run(name, func(t *testing.T) {
			store, cleanup, err := factory("append_verify")
			if err != nil {
				t.Fatalf("failed to create %s store: %v", name, err)
			}
			defer cleanup()

			chain := NewChain(store)
			issueDate := time.Date(2026, 5, 13, 0, 0, 0, 0, time.UTC)

			// Empty chain verify
			if err := chain.Verify(); err != nil {
				t.Fatalf("empty chain Verify: %v", err)
			}

			// Append first
			r1, err := chain.Append("001", "SERIE", "A12345678", issueDate, "F1", 21.00, 121.00)
			if err != nil {
				t.Fatalf("Append #1: %v", err)
			}
			if r1.PreviousFingerprint != "" {
				t.Errorf("first record: expected empty PreviousFingerprint, got %q", r1.PreviousFingerprint)
			}
			if len(r1.Fingerprint) != 64 {
				t.Errorf("first record: expected 64-char fingerprint, got %d", len(r1.Fingerprint))
			}
			if r1.InvoiceType != "F1" {
				t.Errorf("first record: InvoiceType = %q, want %q", r1.InvoiceType, "F1")
			}
			if r1.TaxAmount != 21.00 {
				t.Errorf("first record: TaxAmount = %.2f, want %.2f", r1.TaxAmount, 21.00)
			}

			// Verify after first
			if err := chain.Verify(); err != nil {
				t.Fatalf("Verify after #1: %v", err)
			}

			// Append second
			r2, err := chain.Append("002", "SERIE", "A12345678", issueDate, "F1", 42.00, 242.00)
			if err != nil {
				t.Fatalf("Append #2: %v", err)
			}
			if r2.PreviousFingerprint != r1.Fingerprint {
				t.Errorf("second record: PreviousFingerprint mismatch:\n  got:  %q\n  want: %q",
					r2.PreviousFingerprint, r1.Fingerprint)
			}
			if r2.Fingerprint == r1.Fingerprint {
				t.Error("second record: fingerprint must differ from first")
			}
			if r2.InvoiceType != "F1" {
				t.Errorf("second record: InvoiceType = %q, want %q", r2.InvoiceType, "F1")
			}
			if r2.TaxAmount != 42.00 {
				t.Errorf("second record: TaxAmount = %.2f, want %.2f", r2.TaxAmount, 42.00)
			}

			// Verify after second
			if err := chain.Verify(); err != nil {
				t.Fatalf("Verify after #2: %v", err)
			}

			// Check All returns both records in order
			all, err := store.All()
			if err != nil {
				t.Fatalf("All: %v", err)
			}
			if len(all) != 2 {
				t.Fatalf("All: got %d records, want 2", len(all))
			}
			if all[0].Fingerprint != r1.Fingerprint {
				t.Errorf("All[0] fingerprint mismatch")
			}
			if all[1].Fingerprint != r2.Fingerprint {
				t.Errorf("All[1] fingerprint mismatch")
			}
		})
	}
}

func TestStores_MultipleCIFs(t *testing.T) {
	factories := stores()

	for name, factory := range factories {
		t.Run(name, func(t *testing.T) {
			store, cleanup, err := factory("multi_cif")
			if err != nil {
				t.Fatalf("create store: %v", err)
			}
			defer cleanup()

			chain := NewChain(store)
			issueDate := time.Date(2026, 5, 13, 0, 0, 0, 0, time.UTC)

			// Two different CIFs
			r1, _ := chain.Append("001", "A", "CIF1", issueDate, "F1", 10.0, 100.0)
			r2, _ := chain.Append("001", "A", "CIF2", issueDate, "F1", 20.0, 200.0)

			// Each CIF's chain starts fresh
			if r1.PreviousFingerprint != "" {
				t.Errorf("CIF1 first: expected empty PrevFP, got %q", r1.PreviousFingerprint)
			}
			if r2.PreviousFingerprint != "" {
				t.Errorf("CIF2 first: expected empty PrevFP, got %q", r2.PreviousFingerprint)
			}

			// Append second invoice for CIF1
			r3, _ := chain.Append("002", "A", "CIF1", issueDate, "F1", 10.0, 150.0)
			if r3.PreviousFingerprint != r1.Fingerprint {
				t.Errorf("CIF1 second: expected PrevFP=%q, got %q", r1.Fingerprint, r3.PreviousFingerprint)
			}

			// CIF2 still independent
			r4, _ := chain.Append("002", "A", "CIF2", issueDate, "F1", 20.0, 250.0)
			if r4.PreviousFingerprint != r2.Fingerprint {
				t.Errorf("CIF2 second: expected PrevFP=%q, got %q", r2.Fingerprint, r4.PreviousFingerprint)
			}

			if err := chain.Verify(); err != nil {
				t.Fatalf("Verify after all appends: %v", err)
			}
		})
	}
}

func TestStores_CanonicalFieldsRoundTrip(t *testing.T) {
	factories := stores()

	for name, factory := range factories {
		t.Run(name, func(t *testing.T) {
			store, cleanup, err := factory("canon_rt")
			if err != nil {
				t.Fatalf("create store: %v", err)
			}
			defer cleanup()

			issueDate := time.Date(2026, 5, 13, 0, 0, 0, 0, time.UTC)
			chain := NewChain(store)

			// Various invoice types
			records := []struct {
				num, series, cif, invType string
				taxAmt, total             float64
			}{
				{"001", "A", "CIF1", "F1", 21.00, 121.00},
				{"002", "A", "CIF1", "F2", 10.50, 60.50},
				{"003", "B", "CIF1", "R1", 5.25, 30.25},
			}

			var fps []string
			for _, r := range records {
				rec, err := chain.Append(r.num, r.series, r.cif, issueDate, r.invType, r.taxAmt, r.total)
				if err != nil {
					t.Fatalf("Append %s: %v", r.num, err)
				}
				fps = append(fps, rec.Fingerprint)

				if rec.InvoiceType != r.invType {
					t.Errorf("record %s: InvoiceType = %q, want %q", r.num, rec.InvoiceType, r.invType)
				}
				if rec.TaxAmount != r.taxAmt {
					t.Errorf("record %s: TaxAmount = %.2f, want %.2f", r.num, rec.TaxAmount, r.taxAmt)
				}
			}

			// Verify all
			if err := chain.Verify(); err != nil {
				t.Fatalf("Verify: %v", err)
			}

			// Read back and compare fingerprints
			all, err := store.All()
			if err != nil {
				t.Fatalf("All: %v", err)
			}
			for i, rec := range all {
				if rec.Fingerprint != fps[i] {
					t.Errorf("record %d: fingerprint changed after round-trip", i)
				}
			}
		})
	}
}

func TestStores_EmptyChain(t *testing.T) {
	factories := stores()

	for name, factory := range factories {
		t.Run(name, func(t *testing.T) {
			store, cleanup, err := factory("empty")
			if err != nil {
				t.Fatalf("create store: %v", err)
			}
			defer cleanup()

			chain := NewChain(store)

			if err := chain.Verify(); err != nil {
				t.Errorf("empty chain Verify: %v", err)
			}

			_, ok := chain.Last()
			if ok {
				t.Error("empty chain Last(): expected false, got true")
			}

			all, err := store.All()
			if err != nil {
				t.Fatalf("All: %v", err)
			}
			if len(all) != 0 {
				t.Errorf("All: got %d records, want 0", len(all))
			}
		})
	}
}
