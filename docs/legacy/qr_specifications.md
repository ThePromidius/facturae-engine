# Especificación del Código QR y Representación Gráfica

Referencia: **Orden HAC/1177/2024, Artículos 20 y 21.**

### 1. Parámetros de la URL
La URL del QR debe contener la siguiente información mínima:

| Parámetro | Descripción | Ref. BOE |
| :--- | :--- | :--- |
| `nif` | NIF del emisor de la factura. | Art. 21.a |
| `numserie` | Número y serie de la factura (concatenados). | Art. 21.b |
| `fecha` | Fecha de expedición (DD-MM-YYYY). | Art. 21.c |
| `importe` | Importe total de la factura. | Art. 21.d |

### 2. URL de Sede Electrónica
**URL Base**: `https://www2.agenciatributaria.gob.es/static_files/common/internet/dep/aplicaciones/es/aeat/tike/cont/qr.html`

### 3. Elementos Visuales Obligatorios en Factura
Si el sistema es **VERI*FACTU**:
- **Logo**: El logotipo oficial de VERI*FACTU (cuando esté disponible).
- **Frase**: *"Factura verificable en la sede electrónica de la AEAT"*.

### 4. Especificaciones del QR
- **Tamaño**: Entre 30x30 y 40x40 mm recomendado para legibilidad.
- **Nivel de corrección de errores**: Mínimo Nivel M (ISO/IEC 18004).

---
*Ref: Orden HAC/1177/2024, Art. 21.2. Implementado en `src/internal/qr/qr.go`*
