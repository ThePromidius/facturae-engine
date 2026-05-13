# Invoice Module

**Package:** `internal/invoice`

**Files:** `request.go`, `validation.go`

The invoice module defines the public JSON contract that the engine accepts as input. It also provides field-level validation and default value injection.

---

## Types

### `Request`

Top-level structure for invoice creation requests.

```go
type Request struct {
    Meta    Meta    `json:"meta"`
    Factura Factura `json:"factura"`
    Emisor  Party   `json:"emisor"`
    Receptor Party  `json:"receptor"`
    Lineas  []Linea `json:"lineas"`
    Pago    *Pago   `json:"pago,omitempty"`
}
```

### `Meta`

Metadata about the invoice format and currency.

```go
type Meta struct {
    Version string `json:"version_formato"` // FacturaE version (e.g. "3.2.2")
    Moneda  string `json:"moneda"`           // Currency code (e.g. "EUR")
}
```

**Defaults** (applied by `DefaultMeta`):
- `Version`: `"3.2.2"`
- `Moneda`: `"EUR"`

### `Factura`

Invoice identification data.

```go
type Factura struct {
    Numero string    `json:"numero"` // Invoice number (e.g. "F2024-001")
    Serie  string    `json:"serie"`  // Invoice series code (e.g. "A")
    Fecha  time.Time `json:"fecha"`  // Issue date (RFC3339)
}
```

### `Party`

A business party (issuer or receiver).

```go
type Party struct {
    CIF       string `json:"cif"`       // Tax ID / NIF
    Nombre    string `json:"nombre"`    // Legal name
    Direccion string `json:"direccion,omitempty"`
    CP        string `json:"cp,omitempty"`
    Ciudad    string `json:"ciudad,omitempty"`
    Provincia string `json:"provincia,omitempty"`
    Pais      string `json:"pais,omitempty"`
    DIR3      *DIR3  `json:"dir3,omitempty"` // Public admin DIR3 codes
}
```

### `Linea`

An invoice line item.

```go
type Linea struct {
    Descripcion    string  `json:"desc"`            // Description
    Cantidad       float64 `json:"cantidad"`        // Quantity
    PrecioUnitario float64 `json:"precio_unitario"` // Unit price (ex-tax)
    IVATipo        float64 `json:"iva_tipo"`         // VAT rate (e.g. 21.0 for 21%)
}
```

### `Pago`

Payment information (optional).

```go
type Pago struct {
    Estado    string     `json:"estado"`              // "pagado" | "pendiente"
    FechaPago *time.Time `json:"fecha_pago,omitempty"` // Payment date
    Metodo    string     `json:"metodo,omitempty"`    // Payment method
}
```

### `DIR3`

Public administration administrative codes (optional).

```go
type DIR3 struct {
    OficinaContable   string `json:"oficina_contable"`
    OrganoGestor      string `json:"organo_gestor"`
    UnidadTramitadora string `json:"unidad_tramitadora"`
}
```

---

## Functions

### `DefaultMeta`

```go
func DefaultMeta(m Meta) Meta
```

Fills empty fields in `Meta` with defaults (`version_formato: "3.2.2"`, `moneda: "EUR"`) and returns the augmented copy.

### `Validate`

```go
func Validate(req Request) error
```

Validates a `Request` and returns a `*ValidationError` (or `nil`). The returned error aggregates all field violations.

---

## Validation Rules

| JSON Path | Rule | Error Code |
|-----------|------|------------|
| `factura.numero` | Non-empty after trim | `factura.numero: campo obligatorio` |
| `factura.fecha` | Not zero value | `factura.fecha: fecha invalida o ausente` |
| `emisor.cif` | Non-empty | `emisor.cif: campo obligatorio` |
| `emisor.nombre` | Non-empty | `emisor.nombre: campo obligatorio` |
| `emisor.direccion` | Non-empty | `emisor.direccion: campo obligatorio para el emisor` |
| `emisor.cp` | Non-empty | `emisor.cp: campo obligatorio para el emisor` |
| `emisor.ciudad` | Non-empty | `emisor.ciudad: campo obligatorio para el emisor` |
| `emisor.pais` | Non-empty | `emisor.pais: campo obligatorio para el emisor` |
| `receptor.cif` | Non-empty | `receptor.cif: campo obligatorio` |
| `receptor.nombre` | Non-empty | `receptor.nombre: campo obligatorio` |
| `lineas` | At least 1 | `lineas: debe incluir al menos una linea de factura` |
| `lineas[n].desc` | Non-empty | `lineas[n].desc: campo obligatorio` |
| `lineas[n].cantidad` | > 0 | `lineas[n].cantidad: debe ser mayor que cero` |
| `lineas[n].precio_unitario` | ≥ 0 | `lineas[n].precio_unitario: no puede ser negativo` |
| `lineas[n].iva_tipo` | 0–100 | `lineas[n].iva_tipo: valor invalido (X.XX)` |

---

## Type: `ValidationError`

```go
type ValidationError struct {
    Errors []string
}
```

Implements the `error` interface. All messages are Spanish.

---

## Cross-References

- Consumed by `api/handlers.go` (`handleInvoice`)
- Input to `facturae.Build()` (`internal/facturae/builder.go`)
- Related: [FacturaE module](facturae.md)
