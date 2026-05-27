# Contributing to FacturaE Engine

Thank you for your interest in the FacturaE Engine — a Go sidecar for Spanish
electronic invoicing with Verifactu compliance, XAdES-BES signing, AEAT
submission, and FACe integration.

**Table of Contents**

- [Documentation Map](#documentation-map)
- [Quick Start](#quick-start)
- [Development Workflow](#development-workflow)
- [Building](#building)
- [Running the Service](#running-the-service)
- [Testing](#testing)
- [Code Conventions](#code-conventions)
- [How to Add a Feature](#how-to-add-a-feature)
- [How to Add an API Endpoint](#how-to-add-an-api-endpoint)
- [How to Add a Chain Backend](#how-to-add-a-chain-backend)
- [Submitting Changes](#submitting-changes)
- [Reporting Issues](#reporting-issues)
- [License](#license)

---

## Documentation Map

All documentation lives in the `docs/` directory as reStructuredText (Sphinx).
The most relevant guides for contributors are:

| Guide | Path | Covers |
|-------|------|--------|
| Development Guide | [docs/development/index.rst](docs/development/index.rst) | Setup, build, run modes, env vars, signer resolution, adding features, conventions |
| Testing Reference | [docs/development/testing.rst](docs/development/testing.rst) | Every test package, run commands, test data |
| Certificate Management | [docs/operations/certificates.rst](docs/operations/certificates.rst) | FNMT certs, self-signed dev certs, `-dev-p12` flag, Docker mount |
| AEAT Submission | [docs/integration/aeat_submission.rst](docs/integration/aeat_submission.rst) | SOAP protocol, endpoints (test/prod), retry logic |
| API Reference | [docs/integration/api_reference.rst](docs/integration/api_reference.rst) | All endpoints, request/response formats |
| Deployment Guide | [docs/operations/deployment.rst](docs/operations/deployment.rst) | Docker, production config, env vars |
| Compliance Checklist | [docs/compliance/checklist.rst](docs/compliance/checklist.rst) | Legal requirements mapped to test proofs |
| Sphinx docs index | [docs/index.rst](docs/index.rst) | Full documentation portal |

---

## Quick Start

### Prerequisites

- **Go 1.25+** (see `go.mod`)
- **OpenSSL** — required for PKCS#12 certificate parsing (`-p12` flag)
- **xmllint** (`libxml2-utils`) — optional, XSD validation falls back gracefully
- **Docker** — optional, for containerised development and dev cert generation

### Setup

```bash
git clone https://github.com/ThePromidius/facturae-engine.git
cd facturae-engine
go mod download
```

---

## Development Workflow

1. Make your changes in the relevant `src/internal/<module>/` package.
2. Write or update tests (table-driven, mock external dependencies).
3. Run `go vet ./src/...` and `go test ./src/...` to verify nothing is broken.
4. Wire new dependencies in `src/cmd/facturae-engine/main.go` if needed.
5. Document your changes (API, module docs, or code comments).
6. Submit a pull request.

---

## Building

```bash
# Standard build
go build -o facturae-engine ./src/cmd/facturae-engine

# Cross-compile for Linux
GOOS=linux GOARCH=amd64 go build -o facturae-engine-linux ./src/cmd/facturae-engine

# Optimised production build (stripped, no CGO)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags="-w -s" -o facturae-engine ./src/cmd/facturae-engine
```

---

## Running the Service

The engine selects a signer using the **first matching rule**:

1. `-p12` — real PKCS#12 signer (requires OpenSSL, enables AEAT TLS)
2. `-dev-p12` — self-signed dev cert (auto-generated if missing, no AEAT TLS)
3. `-key + -cert` — PEM-based signer
4. None — MockSigner (SHA-256 comment, no real signature)

```bash
# Mock signing (no cert, no AEAT)
go run ./src/cmd/facturae-engine

# Self-signed dev cert (auto-generated)
go run ./src/cmd/facturae-engine -dev-p12 dev.p12

# Real certificate
go run ./src/cmd/facturae-engine -p12 cert.p12 -p12pass secret

# With AEAT test environment
go run ./src/cmd/facturae-engine -dev-p12 dev.p12 -aeat test

# With SQL persistence
go run ./src/cmd/facturae-engine -db postgres -dsn "postgres://user:pass@localhost:5432/facturae?sslmode=disable"
```

See the [Development Guide](docs/development/index.rst) for the full list of
flags and environment variables.

---

## Testing

```bash
# All tests
go test ./src/...

# Verbose output
go test -v ./src/...

# Specific package
go test -v ./src/internal/api/...

# Specific test function
go test -v -run TestInvoice_HappyPath_Returns200 ./src/internal/api/...

# Coverage
go test -coverprofile=coverage.out ./src/...
go tool cover -html=coverage.out
```

### Test Packages

| Package | File | Tests | Focus |
|---------|------|-------|-------|
| `api` | `server_test.go` | 14 | HTTP handlers, server lifecycle, QR endpoint |
| `api` (integration) | `integration_test.go` | 7 | Full POST /invoice pipeline |
| `aeat` | `client_test.go`, `aeat_test.go` | 10 | SOAP client, XML marshalling, retry logic |
| `aeat` (integration) | `integration_test.go` | 1 | Mock AEAT round-trip |
| `signing` | `signer_test.go`, `pkcs12_test.go` | 14 | MockSigner, P12Signer, PKCS#12 loading |
| `chain` | `chain_test.go` | 2 | Canonicalisation, fingerprinting |
| `facturae` | `builder_test.go`, `interop_test.go` | 22 | XML builder, validation, UBL interop |
| `invoice` | `validation_test.go` | 9 | Request validation rules |
| `legal` | `compliance_test.go`, `legal_test.go` | 5 | Verifactu compliance checks |
| `schema` | `manager_test.go`, `verifactu_test.go` | 9 | XSD download, cache, validation |
| `qr` | `qr_test.go` | 1 | QR code generation |

For detailed breakdowns of every test function, see the
[Testing Reference](docs/development/testing.rst).

### Test Data

Sample invoice JSONs are in `src/testdata/`:

```bash
curl -X POST http://127.0.0.1:8080/invoice \
  -H "Content-Type: application/json" \
  -d @src/testdata/invoice_simple.json
```

- `invoice_simple.json` — 2 lines at 21% IVA (total 1815.00)
- `invoice_multi_iva.json` — 4 lines at 4/10/21/0% (total 4176.60)

---

## Code Conventions

- **Idiomatic Go** — prefer standard library over external packages.
- **API errors in Spanish** — user-facing error messages are in Spanish.
- **Internal logs in English** — system/debug logs are in English.
- **Godoc** — exported types and functions must have godoc comments.
- **Table-driven tests** — use the `t.Run()` subtest pattern.
- **Mock externals** — use mock implementations for AEAT, FACe, and other
  external services (see `internal/aeat/client_test.go` for an example).
- **Zero heavy frameworks** — the project uses only stdlib `net/http`,
  `database/sql`, and minimal third-party drivers (`pgx`, `modernc/sqlite`).
- **Run `go vet ./src/...`** before every commit.

---

## How to Add a Feature

1. **Add types** in the relevant `src/internal/<module>/` package.
2. **Write tests** following the table-driven pattern. Every exported function
   should have a corresponding test.
3. **Wire dependencies** in `src/cmd/facturae-engine/main.go` by adding a CLI
   flag (with env var fallback) and initialising the component.
4. **Expose via API** — add a handler in `src/internal/api/handlers.go` and
   register the route in `src/internal/api/server.go`.
5. **Document** the new module in the Sphinx docs.
6. **Run all tests** before submitting.

### How to Add an API Endpoint

```go
// 1. Add handler on Server
func (s *Server) handleMyEndpoint(w http.ResponseWriter, r *http.Request) {
    // ...
    respondJSON(w, http.StatusOK, result)
}

// 2. Register route in server.go
mux.HandleFunc("/my-endpoint", s.handleMyEndpoint)

// 3. Document in docs/integration/api_reference.rst
// 4. Add tests in src/internal/api/server_test.go
```

### How to Add a Chain Backend

```go
// 1. Implement the Store interface (internal/chain/store.go)
type MyStore struct { ... }

func (s *MyStore) Save(r Record) error                { ... }
func (s *MyStore) Last(emisorCIF string) (Record, bool, error) { ... }
func (s *MyStore) All() ([]Record, error)             { ... }
func (s *MyStore) Close() error                       { ... }

// 2. Wire in main.go
case "mybackend":
    cs = chain.NewMyStore(dsn)

// 3. Add tests in src/internal/chain/
```

---

## Submitting Changes

1. **Fork** the repository on GitHub.
2. **Create a feature branch** from `main`:
   ```bash
   git checkout -b feat/my-feature
   ```
3. **Commit your changes** with clear messages:
   ```bash
   git commit -m "feat: add support for X"
   ```
   Use conventional commit prefixes: `feat:`, `fix:`, `docs:`, `refactor:`,
   `test:`, `chore:`.
4. **Push** to your fork:
   ```bash
   git push origin feat/my-feature
   ```
5. **Open a pull request** with:
   - A clear summary of the change and its motivation.
   - Links to any related issues.
   - Confirmation that all tests pass.
   - Screenshots or curl examples if the change adds a visible behaviour.

---

## Reporting Issues

### Bug Reports

Include the following information:

- **Go version:** `go version`
- **OS:** Windows / Linux / macOS
- **Flags used:** full command line or env vars
- **Steps to reproduce:** what you did
- **Expected behaviour:** what should happen
- **Actual behaviour:** what happens (include logs, error messages)
- **Test data:** the invoice JSON you used (redact sensitive info)

### Feature Requests

Describe:

- **Use case:** what problem you are solving
- **Proposed behaviour:** how you expect it to work
- **Context:** any relevant links (AEAT specs, BOE publications, etc.)

---

## License

This project is licensed under the **Business Source License 1.1**.

- **Free for Individual & Internal Use** — you can use this code for your own
  projects or internal company tools for free.
- **Commercial License Required** for offering this software as a Managed
  Service (SaaS) or including it in a Commercial Product.
- **Open Source Conversion** — the license automatically converts to
  **Apache License 2.0** on **January 1st, 2029**.

See the [LICENSE](LICENSE) file for full terms.

---

Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
