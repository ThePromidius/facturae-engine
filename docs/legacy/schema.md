# Schema Module

**Package:** `internal/schema`

**Files:** `manager.go`, `validator.go`

This module manages XSD (XML Schema Definition) files for FacturaE validation. It caches downloaded schemas and provides a validation wrapper around the `xmllint` command-line tool.

---

## Manager

### `KnownVersions`

```go
var KnownVersions = map[string]string{
    "3.2.2": "https://www.facturae.gob.es/content/dam/facturae/formato/versiones/Facturaev3_2_2.xml",
    "3.2.1": "https://www.facturae.gob.es/content/dam/facturae/formato/versiones/Facturaev3_2_1.xml",
}
```

Maps FacturaE version strings to their official XSD download URLs.

### `Manager`

```go
type Manager struct {
    cacheDir string
    client   *http.Client
    mu       sync.Mutex
}

func NewManager(cacheDir string) (*Manager, error)
```

Creates a new schema manager, ensuring the cache directory exists (created if missing with `0755` permissions). The HTTP client has a 30-second timeout.

### `SchemaPath`

```go
func (m *Manager) SchemaPath(version string) (string, error)
```

Returns the local filesystem path for the requested version's XSD:

1. Check if `facturae_<version>.xsd` exists in the cache directory
2. If cached, return the path immediately
3. If not cached, look up the URL in `KnownVersions`
4. Download the XSD to a temporary file, then atomically rename to the final path
5. Return the local path

Thread-safe via `sync.Mutex`.

### `IsCached`

```go
func (m *Manager) IsCached(version string) bool
```

Returns `true` if the XSD for the given version is already cached locally.

### `ClearCache`

```go
func (m *Manager) ClearCache() error
```

Removes all `.xsd` files from the cache directory.

### Download Process

```go
func (m *Manager) download(url, dest string) error
```

1. HTTP GET the URL
2. Write to a temporary file in the cache directory (`*.xsd.tmp`)
3. Atomically rename to the final path

---

## Validator

### `ErrXmllintMissing`

```go
var ErrXmllintMissing = errors.New("xmllint no esta instalado")
```

Sentinel error returned when `xmllint` is not found in PATH.

### `ValidateXML`

```go
func ValidateXML(xmlBytes []byte, schemaPath string) error
```

Validates XML against an XSD schema using `xmllint`:

1. Check that `xmllint` is available (returns `ErrXmllintMissing` if not)
2. Check that the schema file exists
3. Execute: `xmllint --noout --schema <schemaPath> -` (reads XML from stdin)
4. Returns `nil` on success, or the `xmllint` error output on failure

```go
if err := schema.ValidateXML(xmlBytes, xsdPath); err != nil {
    if errors.Is(err, schema.ErrXmllintMissing) {
        // Graceful fallback: xmllint not available
    } else {
        // Schema validation failed
    }
}
```

---

## Usage in Pipeline

```go
// Get the schema path (download if needed)
xsdPath, err := s.schemas.SchemaPath("3.2.2")

// Validate the generated XML
err := schema.ValidateXML(xmlBytes, xsdPath)
if errors.Is(err, schema.ErrXmllintMissing) {
    // xmllint not available — skip strict validation
} else if err != nil {
    // XSD validation failed — return 422
}
```

---

## Cross-References

- Used by: `api/handlers.go` (`handleInvoice` — XSD validation step)
- Initialized by: `api/server.go` (`New`)
- Cache directory config: `cmd/facturae-engine/main.go` (`-schemas` flag)
- Validates output of: [FacturaE module](facturae.md)
