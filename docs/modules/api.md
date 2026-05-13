---

## Module Documentation

### `internal/api`

-   **Files:** `server.go`, `handlers.go`
-   **Purpose:** Provides the HTTP server that exposes FacturaE invoice processing endpoints: `/invoice` (POST), `/chain` (GET), `/health` (GET), and `/qr` (GET). It wires together signing, schema validation, Verifactu chain hashing, and optional AEAT submission.

### Server Configuration (`Config`)

```go
type Config struct {
    SocketPath string
    SchemaDir  string
    Signer     signing.Signer
    AEATClient *aeat.Client
    ChainStore chain.Store
}
```

### `Server` Struct

```go
type Server struct {
    cfg     Config
    chain   *chain.Chain
    schemas *schema.Manager
    signer  signing.Signer
    aeat    *aeat.Client
    http    *http.Server
}

func New(cfg Config) (*Server, error)
```

Creates a new server instance, initializing the schema manager, Verifactu chain, and HTTP routes. Defaults to an in-memory chain store if none is provided.

### Server Lifecycle

```go
func (s *Server) ListenAndServe() error
func (s *Server) Shutdown(ctx context.Context) error
```

`ListenAndServe` starts the HTTP server, supporting both TCP and Unix domain sockets. `Shutdown` performs a graceful shutdown.

---

## Handlers

### `handleInvoice`

-   **Endpoint:** `POST /invoice`
-   **Function:** `(s *Server) handleInvoice(w http.ResponseWriter, r *http.Request)`

Processes an invoice request:
1.  Decode JSON → `invoice.Request`.
2.  Validate `invoice.Request`.
3.  Inject default metadata (`invoice.DefaultMeta`).
4.  Build `FacturaE` struct (`facturae.Build`).
5.  Validate `FacturaE` struct (`facturae.ValidateStruct`).
6.  Marshal to XML (`xml.MarshalIndent`).
7.  Validate XML bytes (`facturae.ValidateXMLBytes`).
8.  Get XSD path and validate against schema (`schema.ValidateXML`).
9.  Sign the XML (`s.signer.Sign`).
10. Append to Verifactu chain (`s.chain.Append`).
11. Generate QR code URL and data URI (`qr.GenerateDataURI`).
12. Optionally submit to AEAT (`s.aeat.Submit`) in a goroutine.
13. Write signed XML response with headers (`X-Verifactu-Fingerprint`, `X-Chain-Length`, QR headers).

Returns appropriate HTTP status codes (200, 400, 422, 500) with JSON error bodies for failures.

### `handleHealth`

-   **Endpoint:** `GET /health`
-   **Function:** `(s *Server) handleHealth(w http.ResponseWriter, r *http.Request)`

Returns server status in JSON format: `{"status": "ok", "chain_length": N, "signer": "algo", "timestamp": "..."}`.

### `handleChain`

-   **Endpoint:** `GET /chain`
-   **Function:** `(s *Server) handleChain(w http.ResponseWriter, r *http.Request)`

Returns all Verifactu chain records as a JSON array: `{"count": N, "records": [...]}`.

### `handleQR`

-   **Endpoint:** `GET /qr?text=<value>`
-   **Function:** `(s *Server) handleQR(w http.ResponseWriter, r *http.Request)`

Generates a PNG QR code for the provided `text` query parameter.

---

## Cross-References

-   Initialized by: `cmd/server/main.go`
-   Uses: `internal/invoice`, `internal/facturae`, `internal/signing`, `internal/chain`, `internal/schema`, `internal/qr`, `internal/aeat`
