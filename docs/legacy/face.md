# FACe Module

**Package:** `internal/face`

**Files:** `client.go`, `signing.go`

This module implements a SOAP client for FACe (Punto General de Entrada de Facturas Electrónicas de la Administración General del Estado), the Spanish public administration e-invoicing gateway. It provides WS-Security (X.509 Binary Security Token) signing of SOAP envelopes.

---

## Client

### `Client`

```go
type Client struct {
    endpoint string
    http     *http.Client
    cert     *tls.Certificate
    privKey  *rsa.PrivateKey
}

func NewClient(endpoint string, cert *tls.Certificate, privKey *rsa.PrivateKey) *Client
```

Creates a new FACe client with:
- mTLS transport configuration (if certificate provided)
- 60-second HTTP timeout
- Private key and certificate for WS-Security envelope signing

### `Enviar`

```go
func (c *Client) Enviar(ctx context.Context, signedXML []byte, filename string) (string, error)
```

Submits a signed FacturaE XML to the FACe web service:

1. Base64-encodes the signed XML
2. Builds the SOAP body (`enviarFactura` RPC call) with:
   - Base64-encoded invoice content
   - Filename
   - MIME type (`application/xml`)
3. Signs the SOAP body with WS-Security (see below)
4. POSTs to the endpoint with `Content-Type: text/xml; charset=utf-8` and `SOAPAction: https://webservice.face.gob.es/sspp#enviarFactura`
5. Returns a reception ID on success

```go
func (c *Client) Enviar(ctx context.Context, signedXML []byte, filename string) (string, error)
```

**SOAP request structure:**

```xml
<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/">
  <soapenv:Header>
    <wsse:Security>
      <wsse:BinarySecurityToken ...>...</wsse:BinarySecurityToken>
      <ds:Signature>...</ds:Signature>
    </wsse:Security>
  </soapenv:Header>
  <soapenv:Body wsu:Id="id-body-...">
    <sspp:enviarFactura xmlns:sspp="https://webservice.face.gob.es/sspp">
      <factura>
        <factura>base64-encoded-signed-xml</factura>
        <nombre>filename.xml</nombre>
        <mime>application/xml</mime>
      </factura>
    </sspp:enviarFactura>
  </soapenv:Body>
</soapenv:Envelope>
```

---

## WS-Security Signing

### `SignSOAPBody`

```go
func SignSOAPBody(bodyContent []byte, privKey *rsa.PrivateKey, certDER []byte) ([]byte, error)
```

Constructs a WS-Security envelope with X.509 Binary Security Token:

1. **Body wrapping:** Wraps the SOAP body content with `wsu:Id` attribute (nanosecond-precision ID)
2. **Digest:** SHA-256 of the tagged body → base64
3. **SignedInfo** with:
   - Canonicalization: Exclusive XML Canonicalization (`http://www.w3.org/2001/10/xml-exc-c14n#`)
   - Signature method: RSA-SHA256
   - Reference to the body element via `#<bodyID>`
   - Transform: Exclusive XML Canonicalization
   - Digest method: SHA-256
4. **Signature:** RSA-SHA256 sign the `SignedInfo` → base64
5. **BinarySecurityToken** with:
   - Encoding: Base64Binary
   - Value type: X509 v3
   - Content: The certificate DER as base64
6. **KeyInfo** references the security token by `URI="#<tokenID>"`
7. Assembles the full SOAP envelope: Header (with Security element) + Body

```go
func SignSOAPBody(bodyContent []byte, privKey *rsa.PrivateKey, certDER []byte) ([]byte, error)
```

**Namespace URIs used:**

| Prefix | URI |
|--------|-----|
| `wsse` | `http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd` |
| `wsu` | `http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd` |
| `ds` | `http://www.w3.org/2000/09/xmldsig#` |
| `soapenv` | `http://schemas.xmlsoap.org/soap/envelope/` |

---

## Cross-References

- Client certificate: [Signing module](signing.md) (`LoadTLSCertFromP12`)
- The signed XML input is the output of the full pipeline (built by [FacturaE module](facturae.md) and signed by [Signing module](signing.md))
