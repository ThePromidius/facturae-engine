# Signing Module

**Package:** `internal/signing`

**Files:** `signer.go`, `mock.go`, `p12.go`, `pkcs12.go`

This module provides the digital signing capability for FacturaE XML documents. It implements a XAdES-BES enveloped signature using RSA-SHA256 with certificate embedding.

---

## Interface: `Signer`

```go
type Signer interface {
    Sign(xmlData []byte) ([]byte, error)
    Algorithm() string
}
```

| Method | Description |
|--------|-------------|
| `Sign(xmlData)` | Returns the XML with an enveloped XAdES-BES signature block inserted before the closing root tag |
| `Algorithm()` | Returns a human-readable description of the signing algorithm |

---

## Implementations

### `MockSigner`

**File:** `mock.go`

A no-op signer for development/testing. Appends an XML comment with the SHA-256 digest instead of a real signature.

```go
type MockSigner struct{}

func (MockSigner) Algorithm() string { return "mock/no-op" }
```

**Sign behavior:** Computes `SHA-256(xmlData)`, appends:
```xml
<!-- XAdES-BES MOCK SIGNATURE | SHA-256: <hex> -->
```

### `P12Signer`

**File:** `p12.go`

The real signer using an RSA private key and X.509 certificate.

```go
type P12Signer struct {
    privateKey  *rsa.PrivateKey
    certificate *x509.Certificate
}

func (s *P12Signer) Algorithm() string {
    return "RSA-SHA256 / XAdES-BES (simplified C14N)"
}
```

**Sign behavior (`Sign` method):**

1. Compute `SHA-256(xmlData)` → base64 digest
2. Build `SignedInfo` XML with:
   - Canonicalization: `http://www.w3.org/TR/2001/REC-xml-c14n-20010315`
   - Signature method: `rsa-sha256`
   - Digest method: `sha256`
   - Transform: `enveloped-signature`
3. RSA-SHA256 sign the `SignedInfo` → base64 signature value
4. Embed the X.509 certificate as base64 DER in `KeyInfo/X509Data`
5. Insert the `ds:Signature` block just before `</fe:Facturae>`

```go
func NewP12SignerFromPEM(keyPEM, certPEM []byte) (*P12Signer, error)
```

Creates a `P12Signer` from PEM-encoded RSA private key and X.509 certificate:
- Supports PKCS#8 and PKCS#1 private key formats
- Only RSA keys are accepted
- The certificate is embedded in the signature output

---

## PKCS#12 Loading

### `LoadFromP12File`

**File:** `pkcs12.go`

```go
func LoadFromP12File(p12Path, password string) (*P12Signer, error)
```

Extracts the private key and certificate from a `.p12`/`.pfx` file using OpenSSL:

1. Runs `openssl pkcs12 -in <path> -nocerts -nodes -passin pass:<password>` for the private key
2. Runs `openssl pkcs12 -in <path> -nokeys -clcerts -passin pass:<password>` for the certificate
3. Passes the PEM outputs to `NewP12SignerFromPEM`

Requires OpenSSL in PATH. On failure, prints a helpful error with manual conversion instructions.

### `LoadTLSCertFromP12`

```go
func LoadTLSCertFromP12(p12Path, password string) (*tls.Certificate, error)
```

Extracts a `tls.Certificate` from a `.p12` file, used as the mTLS client certificate for AEAT connections. Uses the same OpenSSL extraction approach.

### `OpenSSLAvailable`

```go
func OpenSSLAvailable() bool
```

Checks if OpenSSL is installed and in PATH.

### `runOpenSSL`

```go
func runOpenSSL(args ...string) (string, error)
```

Helper that executes `openssl` with given arguments, captures stdout, and returns stderr on failure.

---

## Usage Examples

```go
// From PEM files
keyPEM, _ := os.ReadFile("key.pem")
certPEM, _ := os.ReadFile("cert.pem")
signer, err := signing.NewP12SignerFromPEM(keyPEM, certPEM)

// From PKCS#12 (requires OpenSSL)
signer, err := signing.LoadFromP12File("cert.p12", "password")

// Use
signed, err := signer.Sign(xmlData)
```

---

## Cross-References

- Used by: `api/handlers.go` (`handleInvoice`)
- Loaded by: `cmd/facturae-engine/main.go` (CLI flag parsing)
- TLS certificate for: [AEAT module](aeat.md)
- Signs output of: [FacturaE module](facturae.md)
