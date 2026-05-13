# Invoice Processing Pipeline

This document describes the complete lifecycle of an invoice from JSON input to signed XML output, covering all validation, signing, and chaining stages.

---

## Stage 1: JSON Decoding

**File:** `internal/api/handlers.go:24-28`, `internal/invoice/request.go`

The handler decodes the HTTP request body into an `invoice.Request` struct. If JSON parsing fails, a `400 Bad Request` is returned.

```json
{
  "meta": {
    "version_formato": "3.2.2",
    "moneda": "EUR"
  },
  "factura": {
    "numero": "F2024-001",
    "serie": "A",
    "fecha": "2024-03-15T00:00:00Z"
  },
  "emisor": {
    "cif": "B12345678",
    "nombre": "Proveedor S.L.",
    "direccion": "Calle Mayor 1",
    "cp": "28001",
    "ciudad": "Madrid",
    "provincia": "Madrid",
    "pais": "ESP"
  },
  "receptor": {
    "cif": "A87654321",
    "nombre": "Cliente S.A."
  },
  "lineas": [
    {
      "desc": "Servicio de consultoria",
      "cantidad": 10,
      "precio_unitario": 100.00,
      "iva_tipo": 21.0
    }
  ],
  "pago": {
    "estado": "pagado",
    "metodo": "transferencia"
  }
}
```

## Stage 2: Invoice Validation

**File:** `internal/invoice/validation.go`

`invoice.Validate()` runs field-level validation:

| Field | Rule |
|-------|------|
| `factura.numero` | Required, non-empty |
| `factura.fecha` | Required, valid time |
| `emisor.cif` | Required |
| `emisor.nombre` | Required |
| `emisor.direccion` | Required for issuer |
| `emisor.cp` | Required for issuer |
| `emisor.ciudad` | Required for issuer |
| `emisor.pais` | Required for issuer |
| `receptor.cif` | Required |
| `receptor.nombre` | Required |
| `lineas` | At least one required |
| `lineas[n].desc` | Required |
| `lineas[n].cantidad` | Must be > 0 |
| `lineas[n].precio_unitario` | Must be ≥ 0 |
| `lineas[n].iva_tipo` | Must be 0–100 |

Errors return `400 Bad Request` with a structured `ValidationError`.

## Stage 3: Default Meta Injection

**File:** `internal/invoice/request.go:55-63`

`invoice.DefaultMeta()` fills missing fields:
- `version_formato` defaults to `"3.2.2"`
- `moneda` defaults to `"EUR"`

## Stage 4: FacturaE Struct Building

**File:** `internal/facturae/builder.go`

`facturae.Build()` maps the invoice `Request` to the `FacturaE` XML struct:

1. **Line computation** (`buildLines` in `lines.go`): For each line, computes:
   - `GrossAmount = cantidad × precio_unitario` (rounded to 2 decimals)
   - `TaxAmount = GrossAmount × iva_tipo / 100` (rounded)
   - Aggregates taxes by rate into a `TaxOutput` summary

2. **FileHeader**: Modality `"I"` (Individual), issuer type `"EM"`, batch identifier composed of CIF + series + number

3. **Parties**: Seller and buyer `Party` with tax identification and legal entity

4. **Invoice**: Header (number, series, document type `FC`, class `OR`), issue data (date, currency, language `es`), taxes, totals, items

## Stage 5: Structural Validation

**File:** `internal/facturae/validation.go`

`facturae.ValidateStruct()` checks the built `FacturaE` struct for structural integrity:

- FileHeader: SchemaVersion, Modality required; InvoicesCount > 0; TotalAmount > 0
- Parties: Both CIF numbers required
- Invoices: At least one invoice required
- Per invoice: InvoiceNumber required, IssueDate required, at least one line, InvoiceTotal ≥ TotalGrossAmount, InvoiceTotal > 0
- Per line: Description required, Quantity > 0, GrossAmount ≥ 0

Returns `422 Unprocessable Entity` on failure.

## Stage 6: XML Marshalling

**File:** `internal/api/handlers.go:48-53`

The struct is marshalled with `xml.MarshalIndent` and the `<?xml version="1.0" encoding="UTF-8"?>` header is prepended.

## Stage 7: XML Parse Validation

**File:** `internal/facturae/validation.go:77-118`

`facturae.ValidateXMLBytes()` performs basic XML sanity checks:
- Minimum length (500 bytes)
- Valid root element named `Facturae`
- `xmlns:fe` namespace declaration present

Returns `422 Unprocessable Entity` on failure.

## Stage 8: XSD Schema Validation

**File:** `internal/schema/validator.go`

`schema.ValidateXML()` runs `xmllint --noout --schema <xsd> -` against the generated XML:

1. `schema.Manager.SchemaPath()` locates or downloads the correct XSD
2. If `xmllint` is not installed, falls back with a warning (`ErrXmllintMissing`)
3. On schema violation, returns the `xmllint` error output

Returns `422 Unprocessable Entity` on failure.

## Stage 9: Digital Signing

**File:** `internal/signing/signer.go`, `internal/signing/p12.go`

`s.signer.Sign(xmlFull)` applies XAdES-BES enveloped signature:

1. SHA-256 digest of the full XML
2. Build `SignedInfo` with canonicalization, signature method (`rsa-sha256`), digest method, and enveloped signature transform
3. RSA-SHA256 signature of `SignedInfo`
4. Build `Signature` element with `SignatureValue` and embedded `X509Certificate`
5. Insert the `Signature` block just before `</fe:Facturae>`

If using `MockSigner`, a SHA-256 comment is appended instead.

## Stage 10: Chain Append (Verifactu Fingerprint)

**File:** `internal/chain/store.go:23-53`

`chain.Append()` creates a chained fingerprint record:

1. Fetch the previous record by `emisorCIF`
2. Canonicalize: `emisorCIF|series|number|date|total|previousFingerprint`
3. SHA-256 of canonical string → hex fingerprint
4. Store the record (memory or SQL)

Headers returned:
- `X-Verifactu-Fingerprint`: current record fingerprint
- `X-Chain-Length`: total chain length

## Stage 11: QR Code Generation (Optional)

**File:** `internal/api/handlers.go:100-111`

A Verifactu verification URL is built with `qr.VerificationURL()` using the emisor CIF, invoice number, series, date, total, and fingerprint. If PNG generation succeeds, headers are set:
- `X-Verifactu-QR-URL`: verification URL
- `X-Verifactu-QR-DataURI`: truncated data URI for embedding

## Stage 12: AEAT Auto-Submission (Optional)

**File:** `internal/api/handlers.go:113-131`

If an AEAT client is configured, the signed XML is submitted to the AEAT Verifactu web service asynchronously:

1. Build SOAP envelope with emisor CIF header and signed XML body
2. POST to the environment endpoint (test or prod)
3. On failure, retry up to 3 times with quadratic backoff (`attempt²` seconds)
4. Log acceptance (with CSV code) or rejection

## Stage 13: Response

**File:** `internal/api/handlers.go:133-137`

The signed XML is returned with:
- `Content-Type: application/xml; charset=utf-8`
- `X-Verifactu-Fingerprint: <hex>`
- `X-Chain-Length: <int>`
- `X-Verifactu-QR-URL: <url>` (if enabled)
- `X-Verifactu-QR-DataURI: <data:uri>` (if enabled)
- Status `200 OK`
