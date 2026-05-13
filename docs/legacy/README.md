# Facturae Engine v2

A Go sidecar service for Spanish electronic invoicing (FacturaE format) with Verifactu compliance, digital signing, AEAT submission, QR generation, and FACe public administration submission.

## Overview

`facturae-engine-v2` is a lightweight HTTP server that accepts JSON invoice requests, converts them to the official Spanish FacturaE XML format (v3.2.2 / v3.2.1), applies XAdES-BES digital signatures, maintains a Verifactu-compliant chained fingerprint record, optionally submits invoices to the AEAT Verifactu web service, generates Verifactu QR codes, and integrates with FACe for public administration invoicing.

## Documentation

| Document | Description |
|----------|-------------|
| [Architecture](architecture.md) | System architecture, module dependency graph, processing pipeline |
| [Workflow](workflow.md) | Step-by-step invoice processing pipeline |
| [API Reference](api.md) | Full API: endpoints, request/response formats, status codes |
| **Modules** | |
| [Invoice](modules/invoice.md) | JSON request types and validation rules |
| [FacturaE](modules/facturae.md) | FacturaE XML model, builder, line/tax computation |
| [Signing](modules/signing.md) | XAdES-BES signing: interface, mock, P12, PEM |
| [Chain](modules/chain.md) | Verifactu fingerprint chain: store interface, memory, SQL |
| [AEAT](modules/aeat.md) | AEAT SOAP client, submission, retry logic |
| [QR](modules/qr.md) | QR code generation for Verifactu |
| [Schema](modules/schema.md) | XSD cache management and validation |
| [FACe](modules/face.md) | FACe SOAP client with WS-Security |
| **Guides** | |
| [Development](guides/development.md) | Setup, building, testing, adding features |
| [Deployment](guides/deployment.md) | Docker, configuration, production considerations |

## Quick Start

```bash
# Build
go build -o facturae-engine ./src/cmd/facturae-engine

# Run with mock signing (development)
./facturae-engine -socket 127.0.0.1:8080 -schemas ./schemas

# Run with real certificate
./facturae-engine -p12 cert.p12 -p12pass secret -socket 127.0.0.1:8080

# Run with AEAT auto-submission
./facturae-engine -p12 cert.p12 -p12pass secret -aeat test -socket 127.0.0.1:8080
```

## License

This project is licensed under the Business Source License 1.1 - see the `LICENSE` file in the root for details.
