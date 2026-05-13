# Especificación Técnica de la Huella (Hash)

Referencia: **Orden HAC/1177/2024, Artículo 13.**

### Algoritmo
- **SHA-256** codificado en hexadecimal.

### Cadena de Entrada (Canonicalización)
Para generar el hash, se deben concatenar los siguientes campos en el orden exacto, separados por el carácter pipe (`|`):

1. **NIF del Emisor**: (p.ej. `B12345678`)
2. **Serie y Número**: Concatenados sin espacios (p.ej. `SERIE001`)
3. **Fecha Expedición**: Formato `YYYY-MM-DD` (ISO 8601).
4. **Tipo Factura**: (p.ej. `F1`)
5. **Cuota Total**: Valor numérico con 2 decimales y punto (p.ej. `21.00`)
6. **Importe Total**: Valor numérico con 2 decimales y punto (p.ej. `121.00`)
7. **Huella Anterior**: Los 64 caracteres del hash anterior.
8. **Timestamp de Generación**: Formato ISO 8601 con zona horaria (p.ej. `2024-03-13T10:00:00+02:00`)

### Reglas de Formato
- **UTF-8**: La cadena de entrada debe estar en codificación UTF-8.
- **Números**: Usar `.` como separador decimal. No usar separador de miles.
- **Campos Vacíos**: Si un campo opcional está vacío, se incluye el separador pero sin contenido.

---
*Ref: Orden HAC/1177/2024, Art. 13.2. Implementado en `src/internal/chain/types.go`*
