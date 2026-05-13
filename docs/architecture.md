# System Architecture

## Module Dependency Graph

```
cmd/facturae-engine (main)
    |
    +--> api (server.go, handlers.go)
    |       |
    |       +--> invoice (request.go, validation.go)
    |       +--> facturae (types.go, builder.go, lines.go, validation.go)
    |       |       |
    |       |       +--> invoice (core types + validation)
    |       |       +--> math (line computation, rounding)
    |       |
    |       +--> signing (signer.go, mock.go, p12.go, pkcs12.go)
    |       +--> chain (types.go, store.go, memory.go, sql.go)
    |       |       |
    |       |       +--> crypto/sha256 (fingerprint computation)
    |       |       +--> database/sql (persistent storage)
    |       |
    |       +--> schema (manager.go, validator.go)
    |       |       |
    |       |       +--> net/http (XSD download)
    |       |       +--> exec (xmllint)
    |       |
    |       +--> aeat (client.go, soap.go, types.go)
    |       |       |
    |       |       +--> net/http (SOAP submission)
    |       |       +--> encoding/xml (SOAP envelope)
    |       |
    |       +--> qr (qr.go, encoder.go)
    |
    +--> signing (used directly by main for TLS cert loading)
    +--> aeat (used directly by main)
    +--> chain (used directly by main for store selection)
```

## Processing Pipeline

```
JSON Request  ──►  Invoice Validation  ──►  FacturaE Build
                       │                           │
                       │                    Line/Tax Computation
                       │                           │
                       ▼                           ▼
                 Validation Error           FacturaE Struct
                                                  │
                                                  ▼
                                          Structural Validation
                                                  │
                                                  ▼
                                          XML Marshal + Header
                                                  │
                                                  ▼
                                       XML Parse Validation (root, ns)
                                                  │
                                                  ▼
                                       XSD Schema Validation (xmllint)
                                                  │
                                                  ▼
                                       Digital Signing (XAdES-BES)
                                                  │
                                                  ▼
                                  ┌─────────────────┬─────────────────┐
                                  │                 │                 │
                                  ▼                 ▼                 ▼
                           Chain Append     (optional) QR Gen   (optional)
                           (fingerprint)    (VerificationURL,   AEAT Submit
                                              PNG, DataURI)     (SOAP, retry)
                                  │
                                  ▼
                           XML Response + Headers
                           (Content-Type, Fingerprint,
                            Chain-Length, QR URLs)
```

## Component Responsibility Matrix

| Component | Responsibility |
|-----------|---------------|
| `cmd/facturae-engine` | CLI flag parsing, dependency injection, signal handling |
| `api` | HTTP server lifecycle, route multiplexing, request/response handling |
| `invoice` | JSON contract types, field-level validation |
| `facturae` | FacturaE XML struct model, builder from invoice, structural + XML validation |
| `signing` | XAdES-BES signing via `Signer` interface, PKCS#12 loading |
| `chain` | Verifactu chained fingerprint: record model, store abstraction, `MemoryStore`, `SQLStore` |
| `aeat` | SOAP client for AEAT Verifactu web service, retry with exponential backoff |
| `qr` | QR code generation: `VerificationURL`, PNG rendering, data URI |
| `schema` | XSD file cache manager, version-aware download, `xmllint` wrapper |
| `face` | SOAP client for FACe (public admin), WS-Security envelope signing |

## Data Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Client / Caller                             │
│                  POST /invoice  (JSON body)                         │
└─────────────────────────┬───────────────────────────────────────────┘
                          │
                          ▼
┌──────────────────────────────────────────────────────────┐
│                    api.Server                             │
│  ┌──────────────────────────────────────────────────────┐│
│  │ handleInvoice                                        ││
│  │  1. Decode JSON → invoice.Request                    ││
│  │  2. invoice.Validate(req)                            ││
│  │  3. facturae.Build(req)                              ││
│  │  4. facturae.ValidateStruct(f)                       ││
│  │  5. xml.MarshalIndent(f)                             ││
│  │  6. facturae.ValidateXMLBytes(xmlBytes)              ││
│  │  7. schema.ValidateXML(xmlBytes, xsdPath)            ││
│  │  8. signer.Sign(xmlFull)                             ││
│  │  9. chain.Append(...) → Record                       ││
│  │  10. qr.GenerateDataURI(url)                         ││
│  │  11. (async) aeat.Submit(signed, cif)                ││
│  │  12. Write signed XML + headers                      ││
│  └──────────────────────────────────────────────────────┘│
└──────────────────────────────────────────────────────────┘
                          │
                          ▼
                  Signed XML Response
                  (Content-Type: application/xml)
                  Headers: X-Verifactu-Fingerprint
                           X-Chain-Length
                           X-Verifactu-QR-URL (optional)
                           X-Verifactu-QR-DataURI (optional)
```

## Service Boundary

The engine is designed as a **sidecar** — it runs alongside other services, listens on localhost (TCP or Unix socket), and handles one concern: converting JSON invoice data into signed, validatable FacturaE XML with full Verifactu compliance. It does not handle business logic, UI, or long-term invoice storage beyond the fingerprint chain.
