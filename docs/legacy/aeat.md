# AEAT Module

**Package:** `internal/aeat`

**Files:** `client.go`, `soap.go`, `types.go`

This module implements the AEAT (Agencia Tributaria) Verifactu web service client. It submits signed FacturaE XML documents to the AEAT for registration and receives a CSV (Secure Verification Code) in response.

---

## Types

### `Environment`

```go
type Environment string

const (
    EnvTest Environment = "test"
    EnvProd Environment = "prod"
)
```

### Endpoints

```go
var Endpoints = map[Environment]string{
    EnvTest: "https://prewww1.aeat.es/wlpl/VERIFACTU-CONT/ws/VeriFactuSOAP",
    EnvProd: "https://www1.aeat.es/wlpl/VERIFACTU-CONT/ws/VeriFactuSOAP",
}
```

### `SOAPAction`

```go
const SOAPAction = "http://www.agenciatributaria.gob.es/AEAT/VERIFACTU/RegistroFacturacion"
```

---

## Client

```go
type Client struct {
    env      Environment
    http     *http.Client
    maxTries int
}

func NewClient(env Environment, cert *tls.Certificate) *Client
```

Creates a new AEAT client:
- Sets the environment (test/prod)
- Optionally configures mTLS with a client certificate
- Default HTTP timeout: 45 seconds
- Default max retries: 3

### Configuration

```go
func (c *Client) SetHTTPClient(client *http.Client)
func (c *Client) SetMaxTries(n int)
```

### `Submit`

```go
func (c *Client) Submit(ctx context.Context, signedXML []byte, emisorCIF string) (*SubmitResult, error)
```

Submits a signed FacturaE XML to the AEAT Verifactu SOAP web service:

1. Builds the SOAP envelope using `buildSOAPEnvelope`
2. Looks up the endpoint for the configured environment
3. Sends an HTTP POST with `Content-Type: text/xml; charset=utf-8` and the `SOAPAction` header
4. On failure, retries up to `maxTries` times with quadratic backoff: `attempt²` seconds between retries
5. Respects context cancellation
6. Parses the SOAP response and returns a `SubmitResult`

### `doRequest`

```go
func (c *Client) doRequest(ctx context.Context, endpoint string, envelope []byte) (*SubmitResult, error)
```

Sends a single SOAP request and parses the response.

---

## SOAP Envelope

### `buildSOAPEnvelope`

```go
func buildSOAPEnvelope(signedXML []byte, emisorCIF string) []byte
```

Constructs the SOAP 1.1 envelope:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope
    xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"
    xmlns:sum="https://www2.agenciatributaria.gob.es/static_files/common/internet/dep/aplicaciones/es/aeat/tike/cont/ws/SuministroLR.xsd">
  <soapenv:Header>
    <sum:Cabecera>
      <sum:ObligadoEmision>
        <sum:NIF>B12345678</sum:NIF>
      </sum:ObligadoEmision>
    </sum:Cabecera>
  </soapenv:Header>
  <soapenv:Body>
    <sum:RegFactuSistemaFacturacion>
      <!-- signed FacturaE XML inserted here as inner XML -->
    </sum:RegFactuSistemaFacturacion>
  </soapenv:Body>
</soapenv:Envelope>
```

---

## Response Types

### `SubmitResult`

```go
type SubmitResult struct {
    CSV         string // Secure Verification Code
    Estado      string // "Correcto", "AceptadoConErrores", or rejection
    Descripcion string // Human-readable description
    HTTPStatus  int    // HTTP status code
}

func (r SubmitResult) IsAccepted() bool
```

`IsAccepted()` returns `true` if `Estado` is `"Correcto"` or `"AceptadoConErrores"`.

### SOAP Response Structure

```go
type soapResponse struct {
    XMLName xml.Name `xml:"Envelope"`
    Body    struct {
        Respuesta struct {
            CSV                    string `xml:"CSV"`
            EstadoEnvio            string `xml:"EstadoEnvio"`
            DescripcionEstadoEnvio string `xml:"DescripcionEstadoEnvio"`
        } `xml:"RespuestaRegFactuSistemaFacturacion"`
    } `xml:"Body"`
}
```

---

## Verifactu Message Types

The module also defines the `SuministroLR` (Suministro de Libros de Registro) message types for reference:

```go
type SuministroLR struct {
    Cabecera Cabecera  `xml:"sum:Cabecera"`
    Registro Registro  `xml:"sum:RegistroFacturacion"`
}

type Registro struct {
    IDFactura      IDFactura  `xml:"sum:IDFactura"`
    FechaOperacion string     `xml:"sum:FechaOperacion"`
    TipoFactura    string     `xml:"sum:TipoFactura"`
    CuotaTotal     float64    `xml:"sum:CuotaTotal"`
    ImporteTotal   float64    `xml:"sum:ImporteTotal"`
    Huella         string     `xml:"sum:Huella"`
    FechaHoraHito  string     `xml:"sum:FechaHoraHito"`
    SistemaInformatico Sistema `xml:"sum:SistemaInformatico"`
}
```

---

## Retry Logic

The `Submit` method implements exponential (quadratic) backoff:

| Attempt | Wait Time |
|---------|-----------|
| 1 | 1 second (not waited — only on failure) |
| 2 | 4 seconds |
| 3 | 9 seconds |

Total maximum wait: ~13 seconds before all 3 attempts are exhausted. The context timeout (default 30 seconds) provides an additional safety boundary.

---

## Cross-References

- Client instantiation: `cmd/facturae-engine/main.go`
- Used by: `api/handlers.go` (async submission in `handleInvoice`)
- TLS cert loading: [Signing module](signing.md) (`LoadTLSCertFromP12`)
- QR verification URL format: [QR module](qr.md)
