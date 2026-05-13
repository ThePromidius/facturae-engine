.. _compliance_checklist:

Legal Compliance Checklist
==========================

This checklist serves as the definitive verification guide for the FacturaE Sidecar. Each item links directly to the technical proof in our test suite.

.. contents::
   :local:

.. _check_inalterability:

1. Inalterability & Integrity
-----------------------------

*   [ ] **Hash Chain Continuity**: Each record must carry the hash of the previous one.
    *   *Technical Proof*: :src:`src/internal/legal/compliance_test.go` (``TestRequirement_HashChaining``)
    *   *Legal Basis*: :ref:`hac_1177_art7`

*   [ ] **Fingerprint Precision**: The SHA-256 hash must use the exact pipe-delimited formula.
    *   *Technical Proof*: :src:`src/internal/legal/compliance_test.go` (``TestRequirement_FingerprintFormula``)
    *   *Legal Basis*: :ref:`hac_1177_art13`

.. _check_authenticity:

2. Authenticity & Signing
-------------------------

*   [ ] **XAdES-BES Profile**: Signatures must include QualifyingProperties and SigningCertificate digest.
    *   *Technical Proof*: :src:`src/internal/legal/compliance_test.go` (``TestRequirement_XAdES_BES_Profile``)
    *   *Legal Basis*: :ref:`hac_1177_art14`

.. _check_audit_trail:

3. Auditability
---------------

*   [ ] **Event Logging**: System must log lifecycle events and cryptographic anomalies.
    *   *Technical Proof*: :src:`src/internal/legal/compliance_test.go` (``TestRequirement_EventLogging``)
    *   *Legal Basis*: :ref:`hac_1177_art9`

.. _check_schema:

4. Structural Integrity (XSD)
-----------------------------

*   [ ] **Verifactu Schema Validation**: Output XML must strictly adhere to ``SuministroLR.xsd``.
    *   *Technical Proof*: :src:`src/internal/schema/verifactu_test.go` (``TestVerifactuSchemaDownload``)
    *   *Legal Basis*: :ref:`hac_1177_art10`

.. tip::
   Run the command ``go test -v ./src/internal/legal/compliance_test.go`` to automatically verify this checklist.
