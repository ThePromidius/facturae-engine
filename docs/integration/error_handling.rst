AEAT Error Handling
===================

The FacturaE Sidecar includes an **Intelligent Error Mapper** that translates technical AEAT rejection codes into human-readable business advice.

.. contents::
   :local:

Error Categories
----------------

AEAT Verifactu responses are categorized into three severity levels:

1.  **Correcto (Success)**: Invoice is valid and legally registered.
2.  **Aceptado con Errores (Warning)**: Invoice is registered, but has technical flaws (e.g., incorrect hash calculation). **Must be fixed in the ERP**.
3.  **Rechazado (Critical)**: Invoice is NOT registered. Legal compliance is broken until fixed.

Common Error Codes & Advice
---------------------------

The sidecar maps the following codes to actionable advice:

+------------+---------------------------+-------------------------------------------------------------+
| Code       | AEAT Message              | Sidecar Advice                                              |
+============+===========================+=============================================================+
| ``4102``   | Error Esquema             | Structural XML error. Check your JSON request fields.       |
+------------+---------------------------+-------------------------------------------------------------+
| ``4107``   | NIF no identificado       | The issuer NIF is not registered in the AEAT Census.        |
+------------+---------------------------+-------------------------------------------------------------+
| ``3000``   | Duplicado                 | This Invoice Number/Series has already been submitted.      |
+------------+---------------------------+-------------------------------------------------------------+
| ``2000``   | Huella Incorrecta         | Chaining hash is mathematically wrong. Audit required.      |
+------------+---------------------------+-------------------------------------------------------------+

Technical Implementation
------------------------

The mapping logic is centralized in:
:src:`src/internal/aeat/errors.go`

.. tip::
   When an error occurs, the sidecar logs the **"Advice"** string to the console and includes it in the ``MapResult()`` method for API consumption.

Mastermind Sync
---------------

This error hardening ensures compliance with **Art. 8 (Legibility)** by making technical rejections understandable for non-technical auditors.
