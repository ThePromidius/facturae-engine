# API Reference

The service listens on the configured address (default `127.0.0.1:8080` for TCP, or a Unix socket path). All responses are synchronous except AEAT submission which runs in a goroutine.

---

## POST /invoice

Convert a JSON invoice to signed FacturaE XML with Verifactu chaining.

### Request

```
POST /invoice
Content-Type: application/json
```

**Body:**

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
    "nombre": "Cliente S.A.",
    "direccion": "Av. Empresa 50",
    "cp": "08001",
    "ciudad": "Barcelona",
    "provincia": "Barcelona",
    "pais": "ESP"
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
    "metodo": "transferencia",
    "fecha_pago": "2024-03-20T00:00:00Z"
  }
}
```

### Responses

**200 OK** — Signed XML returned

```
Content-Type: application/xml; charset=utf-8
X-Verifactu-Fingerprint: a1b2c3d4e5f6...
X-Chain-Length: 42
X-Verifactu-QR-URL: https://www2.agenciatributaria.gob.es/wlpl/VERIFACTU/ConsultaPublica?nif=...
X-Verifactu-QR-DataURI: data:image/png;base64,iVBOR... (truncated in header)

<?xml version="1.0" encoding="UTF-8"?>
<fe:Facturae xmlns:fe="http://www.facturae.gob.es/formato/Versiones/Facturaev3_2_2.xml" ...>
  ...
  <ds:Signature xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
    <ds:SignedInfo>...</ds:SignedInfo>
    <ds:SignatureValue>...</ds:SignatureValue>
    <ds:KeyInfo>...</ds:KeyInfo>
  </ds:Signature>
</fe:Facturae>
```

**400 Bad Request** — Validation error

```json
{
  "error": "validation failed:\n  * factura.numero: campo obligatorio\n  * lineas: debe incluir al menos una linea de factura"
}
```

**422 Unprocessable Entity** — Structural, XML, or XSD validation failed

```json
{
  "error": "structural error at Invoices.Invoice[0].InvoiceTotals.InvoiceTotal: must be > 0"
}
```

**500 Internal Server Error** — Building, signing, or chain error

```json
{
  "error": "Error construyendo FacturaE: building lines: linea 0: cantidad debe ser > 0"
}
```

### Response Headers

| Header | Description |
|--------|-------------|
| `Content-Type` | `application/xml; charset=utf-8` |
| `X-Verifactu-Fingerprint` | SHA-256 fingerprint of the current chained record (hex) |
| `X-Chain-Length` | Total number of records in the chain |
| `X-Verifactu-QR-URL` | Verifactu verification URL (if QR generation succeeded) |
| `X-Verifactu-QR-DataURI` | Truncated PNG data URI (if QR generation succeeded) |

---

## GET /health

Returns engine status.

### Request

```
GET /health
```

### Response

**200 OK**

```json
{
  "status": "ok",
  "chain_length": 42,
  "signer": "RSA-SHA256 / XAdES-BES (simplified C14N)",
  "timestamp": "2024-03-15T10:30:00Z"
}
```

| Field | Description |
|-------|-------------|
| `status` | Always `"ok"` |
| `chain_length` | Number of records in the Verifactu chain |
| `signer` | Current signing algorithm description |
| `timestamp` | Current UTC time (RFC3339) |

---

## GET /chain

Returns all Verifactu chain records.

### Request

```
GET /chain
```

### Response

**200 OK**

```json
{
  "count": 2,
  "records": [
    {
      "InvoiceNumber": "F2024-001",
      "InvoiceSeries": "A",
      "EmisorCIF": "B12345678",
      "IssueDate": "2024-03-15T00:00:00Z",
      "Total": 1210.00,
      "PreviousFingerprint": "",
      "Fingerprint": "abc123...",
      "Timestamp": "2024-03-15T10:30:00Z"
    },
    {
      "InvoiceNumber": "F2024-002",
      "InvoiceSeries": "A",
      "EmisorCIF": "B12345678",
      "IssueDate": "2024-03-16T00:00:00Z",
      "Total": 2420.00,
      "PreviousFingerprint": "abc123...",
      "Fingerprint": "def456...",
      "Timestamp": "2024-03-16T10:30:00Z"
    }
  ]
}
```

### Errors

**405 Method Not Allowed** — Only GET is accepted

```json
{
  "error": "Solo se permite GET"
}
```

---

## GET /qr

Generate a QR code PNG for arbitrary text (general-purpose, not Verifactu-specific).

### Request

```
GET /qr?text=https://example.com/verify?nif=B12345678
```

### Response

**200 OK** — PNG image

```
Content-Type: image/png
Cache-Control: public, max-age=86400
```

(PNG binary data)

### Errors

**400 Bad Request** — Missing `text` parameter

```json
{
  "error": "missing ?text= parameter"
}
```

**422 Unprocessable Entity** — Text too long (>300 chars)

```json
{
  "error": "QR generation failed: qr: encoding text too long for simplified encoder (350 chars, max 300)"
}
```

---

## Common Headers

| Header | Applies To | Description |
|--------|-----------|-------------|
| `Content-Type` | All | `application/json` for errors, `application/xml` for invoice, `image/png` for QR |
| `Cache-Control` | `GET /qr` | `public, max-age=86400` (1 day) |

## Error Format

All error responses return JSON with a single `error` key:

```json
{
  "error": "human-readable error message"
}
```

## HTTP Status Code Summary

| Code | Meaning |
|------|---------|
| 200 | Success |
| 400 | Bad request (invalid JSON, validation failure, missing parameter) |
| 405 | Method not allowed |
| 422 | Unprocessable entity (structural, XML, or XSD validation failure) |
| 500 | Internal server error (building, signing, chain store failure) |
