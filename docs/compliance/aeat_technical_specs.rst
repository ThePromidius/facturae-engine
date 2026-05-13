AEAT Technical Mastermind (2026)
===============================

This document tracks the technical specifications from the AEAT Sede Electrónica as of May 13, 2026.

.. contents::
   :local:

Current Official Sources
------------------------

*   **XSD Schemas (v1.0):** Used for XML structure validation.
*   **Signature Policy (v0.1.5):** Defines XAdES-BES profile.
*   **FAQ for Developers (Dec 2025):** The "Working Bible" for technical edge cases.

XML Schema Definitions (XSD)
----------------------------

The "Magic Box" must validate its output against these definitive schemas.

.. list-table:: AEAT XSD Registry
   :widths: 30 70
   :header-rows: 1

   * - File Name
     - Purpose
   * - ``SuministroLR.xsd``
     - Root schema for Alta and Anulación operations.
   * - ``SuministroInformacion.xsd``
     - Common data types (NIF, Dates, Amounts).
   * - ``EventosSIF.xsd``
     - Schema for mandatory Art. 9 Event Logs.

.. _aeat_signature_policy:

Signature Policy (XAdES-BES)
----------------------------

Based on document ``EspecTecGenerFirmaElectRfact.pdf (v0.1.5)``.

**Mandatory Profile Requirements:**

1.  **Format:** XAdES-BES Enveloped.
2.  **Algorithm:** RSA-SHA256 (min 2048 bits).
3.  **Namespaces:**
    *   ``ds``: ``http://www.w3.org/2000/09/xmldsig#``
    *   ``xades``: ``http://uri.etsi.org/01903/v1.3.2#``
4.  **SignedProperties:** Must include ``SigningTime`` and ``SigningCertificate``.

.. _aeat_hash_algorithm:

Hash Fingerprint (Huella)
-------------------------

Based on ``Veri-Factu_especificaciones_huella_hash_registros.pdf``.

*   **Algorithm:** SHA-256.
*   **Encoding:** UTF-8.
*   **Delimiter:** Pipe (``|``).
*   **Precision:** Exactly 2 decimal places for amounts.

Operational Endpoints (Test Environment)
----------------------------------------

*   **Base URL:** ``https://prewww2.aeat.es/``
*   **WSDL Path:** ``/static_files/common/internet/dep/aplicaciones/es/aeat/tikeV1.0/cont/ws/...``

Integrity Safeguards
--------------------

*   **First Record:** Previous Fingerprint MUST be empty.
*   **Chain Continuity:** Any gap or mismatch in the hash chain invalidates compliance for the entire fiscal year.
