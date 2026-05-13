# Especificaciones del Registro de Alta (Veri*factu)

Referencia: **Orden HAC/1177/2024, Anexo I, Apartado 3.**

### Estructura del XML
El XML debe seguir el esquema XSD publicado en la Sede Electrónica de la AEAT.

| Campo | Requerimiento | Ref. BOE |
| :--- | :--- | :--- |
| `IDFactura` | Debe contener `NIF`, `NumSerieFactura` y `FechaExpedicionFactura`. | Art. 10 |
| `NombreRazon` | Nombre o razón social del emisor. | Anexo I.3 |
| `TipoFactura` | Código según tablas AEAT (F1, F2, R1, etc.). | Art. 10 |
| `ImporteTotal` | Importe total de la factura. | Art. 10 |
| `CuotaTotal` | Suma de todas las cuotas de IVA del registro. | Art. 13.1.a.5º |
| `RegistroAnterior` | Bloque obligatorio (excepto primer registro). | Art. 7.a |

### Bloque `<RegistroAnterior>`
Este bloque es la base del encadenamiento (Blockchain-like):
1. **IDEmisor**: NIF del emisor del registro previo.
2. **NumSerieFactura**: Número y serie del registro previo.
3. **FechaExpedicion**: Fecha del registro previo.
4. **Huella**: Los primeros 64 caracteres de la huella SHA-256 del registro previo.

---
*Documentación técnica para desarrolladores basada en la Orden HAC/1177/2024.*
