Interoperability & Crea y Crece
================================

This document outlines how the FacturaE Sidecar ensures compliance with **Ley 18/2022 (Crea y Crece)** by providing format interoperability for outdated ERPs.

.. contents::
   :local:

The Sidecar Responsibility
--------------------------

As a sidecar engine, this system acts as a **syntax translator**. Its goal is to allow an ERP to send a single set of data and receive any legally mandated format required by the receiver.

Key Obligations under Crea y Crece:
*   **Format Diversity**: Capability to produce UBL 2.1, CII, and FacturaE.
*   **EN 16931 Alignment**: All generated XMLs must follow the European semantic model.
*   **Black Box Integrity**: Ensuring the same business data results in identical signatures across formats.

Supported Formats (Roadmap)
---------------------------

+---------------+-----------------------+------------------------------------------+
| Format        | Standard              | Implementation Status                    |
+===============+=======================+==========================================+
| **FacturaE**  | Spanish National      | ✅ Production Ready                       |
+---------------+-----------------------+------------------------------------------+
| **UBL 2.1**   | ISO/IEC 19845         | 🚧 In Development (EN 16931 Profile)      |
+---------------+-----------------------+------------------------------------------+
| **CII**       | UN/CEFACT             | 🕒 Scheduled Q4 2026                     |
+---------------+-----------------------+------------------------------------------+

.. _ubl_mapping:

UBL 2.1 Mapping (EN 16931)
--------------------------

To satisfy European interoperability, the sidecar maps internal fields to UBL elements:

*   **Invoice ID**: ``cbc:ID``
*   **Issue Date**: ``cbc:IssueDate``
*   **Issuer NIF**: ``cac:AccountingSupplierParty/cac:Party/cac:PartyTaxScheme/cbc:CompanyID``
*   **Total Amount**: ``cac:LegalMonetaryTotal/cbc:PayableAmount``

Legal Mastermind: Why UBL?
--------------------------

While **FacturaE** is the standard for the Spanish Administration (B2G), **UBL** is the preferred standard for international B2B transactions and is mandated as a "must-accept" format under the Crea y Crece interoperability framework.

Proof of Interoperability
-------------------------

The mapping logic and multi-format generation are verified by:
:src:`src/internal/facturae/interop_test.go`
