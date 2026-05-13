# FacturaE Engine

Welcome to the **FacturaE Engine** documentation.

This project provides a Go-based service for generating, signing, validating, and chaining Spanish electronic invoices (FacturaE format). It is designed to be used as a "Sidecar" service in modern architectures.

It adheres to the **Verifactu** standard, integrates with **AEAT** (Spanish Tax Agency) for submission, and supports **QR code** generation for verification.

---

## Getting Started

### Prerequisites

-   Go 1.25+
-   OpenSSL (for PKCS#12 support)
-   `xmllint` (for XSD validation)

### Installation

```bash
# Clone the repository
git clone https://github.com/ThePromidius/facturae-engine.git
cd facturae-engine

# Download dependencies
go mod download
```

---

## Core Features

-   **JSON to FacturaE XML Conversion:** Accepts invoice data in a structured JSON format.
-   **FacturaE Validation:** Enforces structural integrity and adheres to official XSD schemas.
-   **Digital Signing:** Implements XAdES-BES enveloped signatures using RSA-SHA256.
-   **Verifactu Chaining:** Creates a tamper-evident chain of invoice records using SHA-256 fingerprints.
-   **AEAT Submission:** Optionally submits invoices to the Spanish Tax Agency (AEAT) with retry logic.
-   **QR Code Generation:** Creates scannable QR codes for invoice verification.
-   **FACe Integration:** Supports submission to the public administration e-invoicing gateway.

---

## Project Structure

```
facturae-engine/
├── LICENSE            # Business Source License 1.1
├── README.md          # This file
├── Dockerfile         # Multi-stage build for production
├── docker-compose.yml # Full stack orchestration
├── go.mod             # Go module definition
└── src/               # Source code
    ├── cmd/           # Application entry points
    ├── internal/      # Private logic (aeat, signing, qr, etc.)
    └── testdata/      # Sample JSON invoices for testing
```

---

## Running the Service

### Development Mode

```bash
go run ./src/cmd/facturae-engine -socket 127.0.0.1:8080 -schemas ./src/testdata
```

### Docker

```bash
docker-compose up --build
```

---

## Testing

```bash
go test ./src/... -v
```

---

## License

This project is licensed under the **Business Source License 1.1**.

### Summary of Terms
- **Free for Individual & Internal Use:** You can use this code for your own projects or internal company tools for free.
- **Commercial License Required for:**
    - Offering this software as a **Managed Service (SaaS)**.
    - Including this software in a **Commercial Product** that you sell to third parties.
- **Open Source Conversion:** This license will automatically convert to **Apache License 2.0** on **January 1st, 2029**.

For full license terms, see the [LICENSE](LICENSE) file.

---
Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
