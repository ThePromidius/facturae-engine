# Development Guide

## Prerequisites

- **Go 1.25+**
- **OpenSSL** (for PKCS#12 support — optional for development with mock signing)
- **xmllint** (`libxml2-utils` package — optional, XSD validation falls back gracefully)
- **Docker** (optional, for containerized development)
- **PostgreSQL** (optional, for SQL chain storage testing)

## Project Structure

```
LICENSE                 # Business Source License 1.1
README.md               # Main documentation
Dockerfile              # Production build
docker-compose.yml      # Orchestration
go.mod                  # Dependencies
src/
  cmd/facturae-engine/  # CLI entry point
  internal/
    api/                # HTTP server + handlers
    invoice/            # JSON contract + validation
    facturae/           # FacturaE XML model + builder
    signing/            # XAdES-BES signing
    chain/              # Verifactu fingerprint chain
    aeat/               # AEAT SOAP client
    qr/                 # QR code generation
    schema/             # XSD cache + validation
    face/               # FACe SOAP client
  testdata/             # Test fixtures
docs/                   # Documentation
```

## Setup

```bash
# Clone and enter
git clone https://github.com/ThePromidius/facturae-engine.git
cd facturae-engine

# Download dependencies
go mod download

# Set up development certificates (if needed)
openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 365 -nodes

# Verify build
go build ./src/cmd/facturae-engine
```

## Building

```bash
# Standard build
go build -o facturae-engine ./src/cmd/facturae-engine

# Cross-compile for Linux
GOOS=linux GOARCH=amd64 go build -o facturae-engine-linux ./src/cmd/facturae-engine

# Optimized production build
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags="-w -s" -o facturae-engine ./src/cmd/facturae-engine
```

## Running

### Development mode (mock signing, no AEAT)

```bash
go run ./src/cmd/facturae-engine -socket 127.0.0.1:8080 -schemas ./src/testdata
```

### With real certificate

```bash
go run ./src/cmd/facturae-engine \
  -socket 127.0.0.1:8080 \
  -schemas ./src/testdata \
  -p12 cert.p12 -p12pass secret
```

### Using Environment Variables

```bash
export DB_DRIVER=sqlite
export DB_DSN=./chain.db
go run ./src/cmd/facturae-engine
```

### With AEAT test submission

```bash
go run ./cmd/facturae-engine \
  -socket 127.0.0.1:8080 \
  -schemas ./schemas \
  -p12 cert.p12 -p12pass secret \
  -aeat test
```

### With SQL persistence (PostgreSQL)

```bash
go run ./cmd/facturae-engine \
  -db postgres -dsn "postgres://user:pass@localhost:5432/facturae?sslmode=disable" \
  -socket 127.0.0.1:8080
```

### With SQLite persistence

```bash
go run ./cmd/facturae-engine \
  -db sqlite -dsn "./data/chain.db" \
  -socket 127.0.0.1:8080
```

## Testing

```bash
# Run all tests
go test ./src/...

# Run tests with verbose output
go test -v ./src/...
```


### Test packages

| Package | Focus |
|---------|-------|
| `internal/api` | HTTP handlers, server lifecycle, QR endpoint |
| `internal/api` (integration) | Full POST /invoice flow |
| `internal/chain` | Memory and SQL store operations, chain verification |
| `internal/facturae` | Builder output, line computation, validation |
| `internal/invoice` | Request validation rules |
| `internal/qr` | QR code generation |
| `internal/schema` | Schema manager |
| `internal/signing` | Signer interface, mock, P12 signer, PKCS#12 loading |

## Adding a New Feature

1. **Add/change types** in the relevant internal package
2. **Write tests** following existing patterns (table-driven tests preferred)
3. **Wire dependencies** in `src/cmd/facturae-engine/main.go`
4. **Expose via API** by adding a handler in `internal/api/handlers.go` and registering the route in `internal/api/server.go`
5. **Document** the new module in `docs/modules/`
6. **Run all tests** before submitting

### Adding a new API endpoint

```go
// 1. Add handler method on Server
func (s *Server) handleMyEndpoint(w http.ResponseWriter, r *http.Request) {
    // ...
}

// 2. Register route in server.go New()
mux.HandleFunc("/my-endpoint", s.handleMyEndpoint)

// 3. Document in docs/api.md
```

### Adding a new chain backend

```go
// 1. Implement the Store interface
type MyStore struct { ... }

func (s *MyStore) Save(r Record) error { ... }
func (s *MyStore) Last(emisorCIF string) (Record, bool, error) { ... }
func (s *MyStore) All() ([]Record, error) { ... }
func (s *MyStore) Close() error { ... }

// 2. Wire in main.go
case "mybackend":
    cs = chain.NewMyStore(dsn)
```

## Code Conventions

- Use idiomatic Go (standard library prefered)
- Error messages in Spanish for API responses (user-facing), English for internal logs
- Exported types and functions have godoc comments
- Table-driven tests
- Mock implementations for external dependencies
- Zero dependencies on heavy frameworks — the project uses only stdlib `net/http`, `database/sql`, and minimal third-party drivers

## Environment Variables

The engine supports configuration via environment variables (ideal for Docker) or CLI flags. Flags take precedence.

| Flag | Environment Variable | Default | Description |
|------|----------------------|---------|-------------|
| `-socket` | `ENGINE_SOCKET` | `127.0.0.1:8080` | Listen address |
| `-schemas` | `ENGINE_SCHEMAS` | `./schemas` | XSD cache directory |
| `-p12` | `CERT_P12_PATH` | | PKCS#12 file path |
| `-p12pass` | `CERT_P12_PASS` | | PKCS#12 password |
| `-key` | `CERT_KEY_PATH` | | PEM private key path |
| `-cert` | `CERT_PEM_PATH` | | PEM certificate path |
| `-aeat` | `AEAT_ENV` | | AEAT environment (`test`/`prod`) |
| `-db` | `DB_DRIVER` | `memory` | Chain store driver (`memory`/`postgres`/`sqlite`) |
| `-dsn` | `DB_DSN` | | Database connection string |
