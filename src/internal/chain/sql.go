// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

// Package chain implements a Verifactu-compliant invoice chain that links
// consecutive records via SHA-256 fingerprints.
package chain

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

// SQLStore is a SQL-backed implementation of the Store interface, supporting
// both PostgreSQL (via pgx) and SQLite (via modernc.org/sqlite).
type SQLStore struct {
	db     *sql.DB
	driver string
}

// NewSQLStore opens a SQL database connection using the given driver ("postgres"
// or "sqlite") and DSN, then initialises the invoice_chain schema.
func NewSQLStore(driver, dsn string) (*SQLStore, error) {
	sqlDriver := driver
	if driver == "postgres" {
		sqlDriver = "pgx"
	}

	db, err := sql.Open(sqlDriver, dsn)
	if err != nil {
		return nil, fmt.Errorf("sql: open %s: %w", driver, err)
	}

	s := &SQLStore{
		db:     db,
		driver: driver,
	}

	if err := s.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("sql: init schema: %w", err)
	}

	return s, nil
}

// initSchema creates the invoice_chain table and an index on (emisor_cif, id)
// if they do not already exist. It uses driver-specific SQL syntax.
func (s *SQLStore) initSchema() error {
	var query string
	if s.driver == "postgres" {
		query = `
		CREATE TABLE IF NOT EXISTS invoice_chain (
			id SERIAL PRIMARY KEY,
			invoice_number TEXT NOT NULL,
			invoice_series TEXT,
			emisor_cif TEXT NOT NULL,
			issue_date DATE NOT NULL,
			total NUMERIC NOT NULL,
			invoice_type TEXT NOT NULL DEFAULT '',
			tax_amount NUMERIC NOT NULL DEFAULT 0.00,
			prev_fingerprint TEXT,
			fingerprint TEXT NOT NULL,
			processed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_emisor_id ON invoice_chain(emisor_cif, id DESC);
		`
	} else {
		query = `
		CREATE TABLE IF NOT EXISTS invoice_chain (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			invoice_number TEXT NOT NULL,
			invoice_series TEXT,
			emisor_cif TEXT NOT NULL,
			issue_date DATE NOT NULL,
			total REAL NOT NULL,
			invoice_type TEXT NOT NULL DEFAULT '',
			tax_amount REAL NOT NULL DEFAULT 0.00,
			prev_fingerprint TEXT,
			fingerprint TEXT NOT NULL,
			processed_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_emisor_id ON invoice_chain(emisor_cif, id DESC);
		`
	}

	_, err := s.db.Exec(query)
	if err != nil {
		return err
	}

	// Migrate existing tables that lack invoice_type / tax_amount
	s.db.Exec("ALTER TABLE invoice_chain ADD COLUMN invoice_type TEXT NOT NULL DEFAULT ''")
	s.db.Exec("ALTER TABLE invoice_chain ADD COLUMN tax_amount REAL NOT NULL DEFAULT 0.00")

	return nil
}

// Save inserts a record into the invoice_chain table using parameterised
// queries compatible with the configured driver.
func (s *SQLStore) Save(r Record) error {
	query := `
	INSERT INTO invoice_chain 
	(invoice_number, invoice_series, emisor_cif, issue_date, total, invoice_type, tax_amount, prev_fingerprint, fingerprint, processed_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	if s.driver == "sqlite" {
		query = `
		INSERT INTO invoice_chain 
		(invoice_number, invoice_series, emisor_cif, issue_date, total, invoice_type, tax_amount, prev_fingerprint, fingerprint, processed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
	}

	_, err := s.db.Exec(query,
		r.InvoiceNumber, r.InvoiceSeries, r.EmisorCIF,
		r.IssueDate.Format("2006-01-02"),
		r.Total, r.InvoiceType, r.TaxAmount,
		r.PreviousFingerprint, r.Fingerprint, r.Timestamp,
	)
	return err
}

// Last queries the most recent record for the given emisorCIF, ordered by id
// DESC. Returns false if no record exists for that issuer.
func (s *SQLStore) Last(emisorCIF string) (Record, bool, error) {
	query := `
	SELECT invoice_number, invoice_series, emisor_cif, issue_date, total, invoice_type, tax_amount, prev_fingerprint, fingerprint, processed_at
	FROM invoice_chain
	WHERE emisor_cif = $1
	ORDER BY id DESC
	LIMIT 1
	`
	if s.driver == "sqlite" {
		query = `
		SELECT invoice_number, invoice_series, emisor_cif, issue_date, total, invoice_type, tax_amount, prev_fingerprint, fingerprint, processed_at
		FROM invoice_chain
		WHERE emisor_cif = ?
		ORDER BY id DESC
		LIMIT 1
		`
	}

	var r Record
	err := s.db.QueryRow(query, emisorCIF).Scan(
		&r.InvoiceNumber, &r.InvoiceSeries, &r.EmisorCIF,
		&r.IssueDate, &r.Total, &r.InvoiceType, &r.TaxAmount,
		&r.PreviousFingerprint, &r.Fingerprint, &r.Timestamp,
	)

	if err == sql.ErrNoRows {
		return Record{}, false, nil
	}
	if err != nil {
		return Record{}, false, err
	}

	return r, true, nil
}

// All returns all records ordered by id ASC (insertion order).
func (s *SQLStore) All() ([]Record, error) {
	query := "SELECT invoice_number, invoice_series, emisor_cif, issue_date, total, invoice_type, tax_amount, prev_fingerprint, fingerprint, processed_at FROM invoice_chain ORDER BY id ASC"
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var r Record
		if err := rows.Scan(&r.InvoiceNumber, &r.InvoiceSeries, &r.EmisorCIF, &r.IssueDate, &r.Total, &r.InvoiceType, &r.TaxAmount, &r.PreviousFingerprint, &r.Fingerprint, &r.Timestamp); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}

// Close closes the underlying SQL database connection.
func (s *SQLStore) Close() error {
	return s.db.Close()
}
