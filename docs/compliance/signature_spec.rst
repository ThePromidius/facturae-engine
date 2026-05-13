XAdES Signature Specification
=============================

To comply with **Art. 14 of Orden HAC/1177/2024**, every invoice record must carry a digital signature following the ETSI EN 319 132 standard.

.. contents::
   :local:

Required Profile
----------------

The "Magic Box" must produce a **XAdES-BES Enveloped Signature**.

.. _xades_bes_profile:

Signature Properties
~~~~~~~~~~~~~~~~~~~~

*   **Canonicalization Method**: ``http://www.w3.org/TR/2001/REC-xml-c14n-20010315``
*   **Signature Method**: ``http://www.w3.org/2001/04/xmldsig-more#rsa-sha256``
*   **Digest Method**: ``http://www.w3.org/2001/04/xmlenc#sha256``
*   **Transform**: ``http://www.w3.org/2000/09/xmldsig#enveloped-signature``

Mandatory XML Elements
----------------------

The signature block (``ds:Signature``) must be inserted as the last child of the root element and MUST include:

1.  **ds:KeyInfo**: Must contain the **X.509 Certificate** used for signing in base64 format (``ds:X509Certificate``).
2.  **xades:QualifyingProperties**: Even for a BES profile, AEAT expects the standard XAdES properties block.
3.  **xades:SigningCertificate**: A digest of the signing certificate to prevent substitution attacks.

Certificate Requirements
------------------------

*   **Type**: Qualified Certificate for Electronic Signature.
*   **Issuer**: A Trust Service Provider (TSP) listed in the **EU Trusted List (EUTL)**.
*   **Status**: Must be valid (not expired or revoked) at the moment of signing.

Compliance Proof
----------------

The signature implementation in our system is located at:
:src:`src/internal/signing/p12.go`

.. warning::
   The signature must cover the entire XML document except for the signature block itself (Enveloped). Any modification to the XML after signing will invalidate the compliance of the record.
