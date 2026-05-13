# Plan de Cumplimiento VERI*FACTU y Ley Crea y Crece

Este documento sirve como checklist y guía de seguimiento para asegurar que la implementación técnica cumple estrictamente con la **Orden HAC/1177/2024** y el **Real Decreto 1007/2023**.

## 1. Trazabilidad y Encadenamiento (Art. 7 Orden HAC/1177/2024)
- [ ] **Registro Anterior**: Cada registro XML debe incluir el NIF, Número/Serie, Fecha y Hash (primeros 64 caracteres) del registro anterior.
- [ ] **Cadena Única**: Todos los registros de un obligado tributario deben formar una única secuencia.
- [ ] **Timestamp Exacto**: Fecha y hora de generación con huso horario (margen error < 1 min).

## 2. Generación de Huella o Hash (Art. 13 Orden HAC/1177/2024)
- [ ] **Algoritmo**: SHA-256 (hexadecimal).
- [ ] **Campos Obligatorios (en orden)**:
    1. NIF Emisor
    2. NumSerieFactura (concatenado)
    3. FechaExpedicionFactura
    4. TipoFactura
    5. CuotaTotal
    6. ImporteTotal
    7. Huella anterior
    8. FechaHoraHito (Timestamp de generación)
- [ ] **Formato de Datos**: UTF-8, números con punto decimal y 2 decimales.

## 3. Firma Electrónica (Art. 14 Orden HAC/1177/2024)
- [ ] **Tipo**: XAdES Enveloped Signature (ETSI EN 319 132).
- [ ] **Certificado**: Certificado cualificado en vigor.

## 4. Código QR y URL (Art. 20 y 21 Orden HAC/1177/2024)
- [ ] **URL Base**: `https://www2.agenciatributaria.gob.es/static_files/common/internet/dep/aplicaciones/es/aeat/tike/cont/qr.html` (o la vigente en sede).
- [ ] **Parámetros URL**: `nif`, `numserie`, `fecha`, `importe`.
- [ ] **Identificación Visual**: Texto "Factura verificable en la sede electrónica de la AEAT" y logo "VERI*FACTU" si aplica.

## 5. Registro de Eventos (Art. 9 Orden HAC/1177/2024)
- [ ] **Eventos Críticos**: Inicio/Fin, anomalías, restauraciones, exportaciones.
- [ ] **Registro Resumen**: Generación cada 6 horas de inactividad o antes de apagado.

## 6. Declaración Responsable (Art. 15 Orden HAC/1177/2024)
- [ ] **Accesibilidad**: Disponible dentro del sistema informático.
- [ ] **Contenido**: Nombre software, versión, NIF productor, cumplimiento normativo expreso.

---
*Referencia legal: [Orden HAC/1177/2024, de 17 de octubre](https://www.boe.es/diario_boe/txt.php?id=BOE-A-2024-22138)*
