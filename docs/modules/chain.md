# Chain Module

**Package:** `internal/chain`

**Files:** `types.go`, `store.go`, `memory.go`, `sql.go`

This module implements the Verifactu-compliant chained fingerprint system. Each invoice record is linked to the previous one via a SHA-256 hash, creating an immutable chain that detects tampering.

---

## Types

### `Record`

```go
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
```

| Field | Description |
|-------|-------------|
| `InvoiceNumber` | The invoice number |
| `InvoiceSeries` | The invoice series code |
| `EmisorCIF` | Issuer's tax ID |
| `IssueDate` | Invoice issue date |
| `Total` | Invoice total amount |
| `PreviousFingerprint` | Fingerprint of the previous record (empty string for first) |
| `Fingerprint` | SHA-256 hex hash of the canonicalized record |
| `Timestamp` | UTC timestamp when the record was created |

### Fingerprint Computation

The fingerprint is computed by:

1. **Canonicalize** the record: `emisorCIF|series|number|YYYY-MM-DD|total|previousFP`
2. **SHA-256** the canonical string → hex digest

```go
func canonicalize(r Record) string
func fingerprint(canonical string) string
```

---

## Interface: `Store`

```go
type Store interface {
    Save(r Record) error
    Last(emisorCIF string) (Record, bool, error)
    All() ([]Record, error)
    Close() error
}
```

| Method | Description |
|--------|-------------|
| `Save(r)` | Persist a record |
| `Last(emisorCIF)` | Return the most recent record for a given CIF; `bool` is false if none exists |
| `All()` | Return all records ordered by insertion |
| `Close()` | Release any held resources |

---

## Implementations

### `MemoryStore`

**File:** `memory.go`

An in-memory, mutex-protected store. Suitable for development and single-instance deployments.

```go
type MemoryStore struct {
    mu      sync.RWMutex
    records []Record
}

func NewMemoryStore() *MemoryStore
```

- `Save`: Appends to a slice with write lock
- `Last`: Reverse-scans the slice for matching CIF
- `All`: Returns a copy of the slice
- `Close`: No-op

### `SQLStore`

**File:** `sql.go`

A SQL-backed store supporting PostgreSQL and SQLite.

```go
type SQLStore struct {
    db     *sql.DB
    driver string
}

func NewSQLStore(driver, dsn string) (*SQLStore, error)
```

Supported drivers:
- `"postgres"` → uses `pgx` driver
- `"sqlite"` → uses `modernc.org/sqlite` driver (pure Go, no CGO)

#### Schema (PostgreSQL)

```sql
CREATE TABLE IF NOT EXISTS invoice_chain (
    id SERIAL PRIMARY KEY,
    invoice_number TEXT NOT NULL,
    invoice_series TEXT,
    emisor_cif TEXT NOT NULL,
    issue_date DATE NOT NULL,
    total NUMERIC NOT NULL,
    prev_fingerprint TEXT,
    fingerprint TEXT NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_emisor_id ON invoice_chain(emisor_cif, id DESC);
```

#### Schema (SQLite)

```sql
CREATE TABLE IF NOT EXISTS invoice_chain (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    invoice_number TEXT NOT NULL,
    invoice_series TEXT,
    emisor_cif TEXT NOT NULL,
    issue_date DATE NOT NULL,
    total REAL NOT NULL,
    prev_fingerprint TEXT,
    fingerprint TEXT NOT NULL,
    processed_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_emisor_id ON invoice_chain(emisor_cif, id DESC);
```

The `SQLStore` uses `$1`-style parameters for PostgreSQL (`pgx`) and `?`-style for SQLite, selected per operation.

---

## Chain Operations

### `Chain`

```go
type Chain struct {
    store Store
}

func NewChain(s Store) *Chain
```

Wraps a `Store` with higher-level operations.

### `Append`

```go
func (c *Chain) Append(
    invoiceNumber, invoiceSeries, emisorCIF string,
    issueDate time.Time,
    total float64,
) (Record, error)
```

Creates and persists a new chained record:
1. Fetch `Last(emisorCIF)` to get the previous fingerprint
2. Build `Record` with the previous fingerprint (empty string if first)
3. Compute current fingerprint
4. Save to store
5. Return the new record

### `Last`

```go
func (c *Chain) Last() (Record, bool)
```

Returns the most recent record across all emisors. This is a convenience wrapper that fetches all records and returns the last one.

### `Len`

```go
func (c *Chain) Len() int
```

Returns the total number of records in the chain.

### `All`

```go
func (c *Chain) All() []Record
```

Returns all records.

### `Verify`

```go
func (c *Chain) Verify() error
```

Verifies chain integrity:
1. Iterates all records in order
2. For each record, checks that `PreviousFingerprint` matches the previous record's `Fingerprint`
3. Recomputes each record's fingerprint and compares with stored value
4. Returns error on first mismatch (broken chain or tampering)

---

## Cross-References

- Used by: `api/handlers.go` (Append after signing)
- Exposed via: `GET /chain` endpoint
- Backend selection: `cmd/facturae-engine/main.go` (`-db` and `-dsn` flags)
- The QR module uses the fingerprint: [QR module](qr.md)
