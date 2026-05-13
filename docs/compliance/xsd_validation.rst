XSD Schema Validation
======================

To ensure the "Magic Box" produces bytes that AEAT will accept, the system implements a **Pre-flight XSD Validation**. This prevents submission of structurally invalid XMLs.

.. contents::
   :local:

Requirement: Schema Compliance
------------------------------

According to **Art. 10 of Orden HAC/1177/2024**, every registration record must follow the technical structure defined in the official XSD schemas.

Implementation Logic
--------------------

The validation process follows these steps:

1.  **Schema Retrieval**: The system automatically downloads and caches the latest official XSDs from the AEAT servers.
2.  **Local Validation**: Before signing or sending, the XML is validated against the local XSD cache.
3.  **Error Handling**: If validation fails, the system blocks the emission and logs a technical anomaly (Art. 9).

Technical Components
--------------------

*   **Schema Manager**: Handles the lifecycle of XSD files (Download, Cache, Versioning).
    *   *Source*: :src:`src/internal/schema/manager.go`
*   **Validator**: Performs the actual structural check.
    *   *Source*: :src:`src/internal/schema/validator.go`

.. note::
   The current implementation requires ``xmllint`` (libxml2) to be available in the system environment (included in the official Docker image).

Supported Schemas (2026)
------------------------

The system is configured to validate against the following "Tike" schemas:

*   **SuministroLR.xsd**: Main schema for Alta/Anulación records.
*   **SuministroInformacion.xsd**: Base types and structures.
*   **EventosSIF.xsd**: Audit trail record structure.

Verification Proof
------------------

Validation integrity is verified by:
:src:`src/internal/schema/verifactu_test.go` (``TestVerifactuXMLValidation``)

.. tip::
   Use the ``--validate-only`` flag in the CLI to check an XML file without performing a full emission.
