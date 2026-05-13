# FacturaE Module

**Package:** `internal/facturae`

**Files:** `types.go`, `builder.go`, `lines.go`, `validation.go`

This module implements the FacturaE XML format (Spanish electronic invoice standard). It provides the complete XML model as Go structs, a builder from the invoice Request, line/tax computation with rounding, and structural/XML-level validation.

---

## Constants

```go
const (
    schemaFe322            = "http://www.facturae.gob.es/formato/Versiones/Facturaev3_2_2.xml"
    schemaDs               = "http://www.w3.org/2000/09/xmldsig#"
    taxTypeIVA             = "01"                        // VAT tax type code
    personTypeLegal        = "J"                         // Legal entity person type
    residenceTypeE         = "R"                         // Resident in Spain
    invoiceDocumentTypeFC  = "FC"                        // Full invoice
    invoiceClassOR         = "OR"                        // Original
    modalityI              = "I"                         // Individual
    issuerType             = "EM"                        // Issuer
)
```

---

## XML Struct Tree

```
FacturaE
├── FileHeader
│   ├── SchemaVersion       string   (e.g. "3.2.2")
│   ├── Modality            string   ("I")
│   ├── InvoiceIssuerType   string   ("EM")
│   └── Batch
│       ├── BatchIdentifier         string
│       ├── InvoicesCount           int
│       ├── TotalInvoicesAmount     Amount {TotalAmount, EquivalentInEuros}
│       ├── TotalOutstandingAmount  Amount
│       ├── TotalExecutableAmount   Amount
│       └── InvoiceCurrencyCode     string
├── Parties
│   ├── SellerParty
│   │   ├── TaxIdentification
│   │   │   ├── PersonTypeCode          string ("J")
│   │   │   ├── ResidenceTypeCode       string ("R")
│   │   │   └── TaxIdentificationNumber string (CIF/NIF)
│   │   └── LegalEntity
│   │       ├── CorporateName    string
│   │       └── RegistrationData (optional)
│   │           └── Address
│   │               ├── Address     string
│   │               ├── PostCode    string
│   │               ├── Town        string
│   │               ├── Province    string (optional)
│   │               └── CountryCode string
│   └── BuyerParty
│       └── (same structure)
└── Invoices
    └── Invoice[]
        ├── InvoiceHeader
        │   ├── InvoiceNumber       string
        │   ├── InvoiceSeriesCode   string (optional)
        │   ├── InvoiceDocumentType string ("FC")
        │   └── InvoiceClass        string ("OR")
        ├── InvoiceIssueData
        │   ├── IssueDate           string (YYYY-MM-DD)
        │   ├── InvoiceCurrencyCode string
        │   ├── TaxCurrencyCode     string
        │   └── LanguageName        string ("es")
        ├── TaxesOutputs
        │   └── Tax[]
        │       ├── TaxTypeCode string ("01")
        │       ├── TaxRate     float64
        │       ├── TaxableBase Amount {TotalAmount}
        │       └── TaxAmount   Amount {TotalAmount}
        ├── InvoiceTotals
        │   ├── TotalGrossAmount            float64
        │   ├── TotalGeneralDiscounts       float64 (optional)
        │   ├── TotalGeneralSurcharges      float64 (optional)
        │   ├── TotalGrossAmountBeforeTaxes float64
        │   ├── TotalTaxOutputs             float64
        │   ├── TotalTaxesWithheld          float64 (optional)
        │   ├── InvoiceTotal                float64
        │   ├── TotalOutstandingAmount      float64
        │   └── TotalExecutableAmount       float64
        └── Items
            └── InvoiceLine[]
                ├── ItemDescription     string
                ├── Quantity            float64
                ├── UnitOfMeasure       string (optional)
                ├── UnitPriceWithoutTax float64
                ├── TotalCost           float64
                ├── GrossAmount         float64
                └── TaxesOutputs
                    └── Tax[]
                        ├── TaxTypeCode string ("01")
                        ├── TaxRate     float64
                        ├── TaxableBase Amount
                        └── TaxAmount   Amount
```

---

## Key Functions

### `Build`

```go
func Build(req invoice.Request) (*FacturaE, error)
```

Maps an `invoice.Request` to the full `FacturaE` struct:

1. Calls `invoice.DefaultMeta()` to fill missing metadata
2. Calls `buildLines()` for line/tax computation
3. Constructs `FileHeader` with batch totals
4. Constructs `Parties` (seller = emisor, buyer = receptor)
5. Constructs `Invoice` with header, issue data, aggregated taxes, computed totals, and items
6. Returns the complete struct ready for marshalling

### `buildLines`

```go
func buildLines(lineas []invoice.Linea, moneda string) ([]InvoiceLine, []TaxOutput, *InvoiceTotals, error)
```

Computes invoice lines and totals:

- **Per line:**
  - `lineGross = round2(cantidad × precio_unitario)`
  - `taxAmt = round2(lineGross × iva_tipo / 100)`
- **Tax aggregation:** Groups line taxes by `(code, rate)` into `taxMap`
- **Totals:**
  - `TotalGrossAmount = Σ lineGross`
  - `TotalTaxOutputs = Σ taxAmt`
  - `InvoiceTotal = round2(gross + totalTax)`
  - `TotalOutstandingAmount = InvoiceTotal`
  - `TotalExecutableAmount = InvoiceTotal`

### `round2`

```go
func round2(v float64) float64
```

Rounds to 2 decimal places using `math.Round(v * 100) / 100`.

---

## Validation

### `ValidateStruct`

```go
func ValidateStruct(f *FacturaE) error
```

Structural validation of the built FacturaE struct. Returns `StructuralErrors` (implements `error`).

| Check | Field |
|-------|-------|
| Non-empty | `FileHeader.SchemaVersion`, `FileHeader.Modality` |
| > 0 | `FileHeader.Batch.InvoicesCount`, `FileHeader.Batch.TotalInvoicesAmount.TotalAmount` |
| Non-empty | Seller CIF, Buyer CIF |
| ≥ 1 | `Invoices.Invoice` |
| Non-empty | Per invoice: `InvoiceNumber`, `IssueDate` |
| ≥ 1 | Per invoice: `Items.InvoiceLine` |
| ≥ `TotalGrossAmount` | Per invoice: `InvoiceTotals.InvoiceTotal` |
| > 0 | Per invoice: `InvoiceTotals.InvoiceTotal` |
| Non-empty | Per line: `ItemDescription` |
| > 0 | Per line: `Quantity` |
| ≥ 0 | Per line: `GrossAmount` |

### `ValidateXMLBytes`

```go
func ValidateXMLBytes(xmlData []byte) error
```

Quick XML sanity checks:
- Length >= 500 bytes
- Root element is `Facturae`
- `xmlns:fe` namespace declared

### Types

```go
type StructuralError struct {
    Field   string
    Message string
}

type StructuralErrors []*StructuralError
```

---

## Cross-References

- Input from: [Invoice module](invoice.md)
- Consumed by: `api/handlers.go` (build, validate, marshal)
- Schema validation: [Schema module](schema.md)
- The built struct is signed by: [Signing module](signing.md)
