UBL 2.1 Technical Mapping
=========================

This specification defines how the FacturaE Sidecar maps internal business data to **UBL 2.1 (ISO/IEC 19845)**, ensuring compliance with the European **EN 16931** semantic model required by the "Crea y Crece" law.

.. contents::
   :local:

Core Mapping Registry
---------------------

The sidecar translates the internal JSON request into the following UBL paths:

+----------------------------+------------------------------------------------------------------------------------------------+
| Sidecar Field (JSON)       | UBL 2.1 Path (XPath)                                                                           |
+============================+================================================================================================+
| ``factura.numero``         | ``/Invoice/cbc:ID``                                                                            |
+----------------------------+------------------------------------------------------------------------------------------------+
| ``factura.fecha``          | ``/Invoice/cbc:IssueDate``                                                                     |
+----------------------------+------------------------------------------------------------------------------------------------+
| ``meta.moneda``            | ``/Invoice/cbc:DocumentCurrencyCode``                                                          |
+----------------------------+------------------------------------------------------------------------------------------------+
| ``emisor.cif``             | ``/Invoice/cac:AccountingSupplierParty/cac:Party/cac:PartyTaxScheme/cbc:CompanyID``            |
+----------------------------+------------------------------------------------------------------------------------------------+
| ``receptor.cif``           | ``/Invoice/cac:AccountingCustomerParty/cac:Party/cac:PartyTaxScheme/cbc:CompanyID``            |
+----------------------------+------------------------------------------------------------------------------------------------+
| ``lineas[].desc``          | ``/Invoice/cac:InvoiceLine/cac:Item/cbc:Description``                                          |
+----------------------------+------------------------------------------------------------------------------------------------+
| ``lineas[].precio_unit``   | ``/Invoice/cac:InvoiceLine/cac:Price/cbc:PriceAmount``                                         |
+----------------------------+------------------------------------------------------------------------------------------------+

Tax Logic (EN 16931 Alignment)
------------------------------

For Spanish VAT (IVA), the system uses the following codes in UBL:

*   **Tax Scheme ID**: ``VAT``
*   **Tax Category**:
    *   ``S``: Standard Rate.
    *   ``E``: Exempt (mapped from internal flags).
    *   ``Z``: Zero Rated.

Bilingual Validation Notes
--------------------------

*   **Spanish (Legalese)**: El sistema garantiza que el modelo semántico europeo (EN 16931) se respete rigurosamente, traduciendo los campos locales a etiquetas universales UBL.
*   **English (Technical)**: The mapping ensures that European syntax requirements are met, allowing outdated ERPs to produce globally recognized invoice formats.

Technical Proof
---------------

Mapping logic is verified by:
:src:`src/internal/ubl/mapping_test.go`
