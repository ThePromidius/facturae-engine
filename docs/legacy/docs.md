---

## Module Documentation

### `internal/face`

-   **Files:** `client.go`, `signing.go`
-   **Purpose:** Implements a SOAP client for FACe (Punto General de Entrada de Facturas Electrónicas de la Administración General del Estado), the Spanish public administration e-invoicing gateway. Supports WS-Security signing.

### `internal/qr`

-   **Files:** `qr.go`, `encoder.go`
-   **Purpose:** Generates QR codes for Verifactu invoice verification. Creates verification URLs, PNG images, and data URIs.

### `internal/schema`

-   **Files:** `manager.go`, `validator.go`
-   **Purpose:** Manages FacturaE XSD schemas; caches downloaded files and validates XML against them using `xmllint`.

### `internal/invoice`

-   **Files:** `request.go`, `validation.go`
-   **Purpose:** Defines the public JSON contract for invoice creation, including validation and default value injection.

### `internal/facturae`

-   **Files:** `types.go`, `builder.go`, `lines.go`, `validation.go`
-   **Purpose:** Implements the FacturaE XML format (Spanish electronic invoice standard). Provides struct models, a builder, tax computation, and validation.

### `internal/chain`

-   **Files:** `types.go`, `store.go`, `memory.go`, `sql.go`
-   **Purpose:** Implements Verifactu-compliant chained fingerprints for tamper detection using SHA-256 hashes.

### `internal/signing`

-   **Files:** `signer.go`, `mock.go`, `p12.go`, `pkcs12.go`
-   **Purpose:** Provides digital signing capabilities (XAdES-BES) using RSA-SHA256 with certificate embedding. Supports mock, P12, and PEM key/cert loading.

### `internal/aeat`

-   **Files:** `client.go`, `soap.go`, `types.go`
-   **Purpose:** Implements a SOAP client for submitting invoices to the AEAT Verifactu web service with retry logic.

### `cmd/server`

-   **Files:** `main.go`
-   **Purpose:** The entry point for the service; parses flags, initializes dependencies, starts the HTTP server, and handles graceful shutdown.

---

## API Endpoints

-   **`POST /invoice`**: Accepts JSON invoice, returns signed FacturaE XML.
-   **`GET /health`**: Returns server status.
-   **`GET /chain`**: Returns all Verifactu chain records.
-   **`GET /qr?text=<value>`**: Generates a QR code PNG.

---

## Cross-References

-   **Input:** `internal/invoice` defines the accepted JSON format.
-   **Output:** `internal/facturae` generates the FacturaE XML.
-   **Signing:** `internal/signing` applies the digital signature.
-   **Chaining:** `internal/chain` adds the Verifactu fingerprint.
-   **Validation:** `internal/schema` performs XSD checks.
-   **Submission:** `internal/aeat` handles AEAT communication.
-   **QR:** `internal/qr` generates verification QR codes.
