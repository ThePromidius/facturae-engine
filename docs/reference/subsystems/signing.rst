Signing Module
==============

The signing module handles the digital signature of FacturaE XML documents, ensuring their authenticity and integrity as required by Spanish law.

Overview
--------

*   **Package:** ``internal/signing``
*   **Compliance:** Implements **Art. 14 of Orden HAC/1177/2024** (XAdES Enveloped Signature).
*   **Source:** :src:`src/internal/signing/`

Interface
---------

The core of the module is the ``Signer`` interface:

.. code-block:: go

   type Signer interface {
       Sign(xmlData []byte) ([]byte, error)
       Algorithm() string
   }

Implementations
---------------

P12Signer
~~~~~~~~~

The primary implementation used in production. It uses an RSA private key and an X.509 certificate to produce a **XAdES-BES** signature.

*   **Canonicalization:** XML-C14N (20010315).
*   **Signature Method:** RSA-SHA256.
*   **Digest Method:** SHA256.
*   **Transform:** Enveloped-signature.

Source: :src:`src/internal/signing/p12.go`

MockSigner
~~~~~~~~~~

Used for development and testing. It appends a SHA-256 comment to the XML instead of performing real cryptographic signing.

Source: :src:`src/internal/signing/mock.go`

Certificate Management
----------------------

The module includes utilities to extract certificates and keys from **PKCS#12** files (``.p12``/``.pfx``).

.. code-block:: go

   func LoadFromP12File(p12Path, password string) (*P12Signer, error)

.. note::
   This function requires ``openssl`` to be available in the system PATH.

Legal Requirements (Mastermind Link)
------------------------------------

This module directly implements the signature requirements specified in:
*   :ref:`hac_1177_art14` (Electronic Signature).
*   :ref:`req_signing` (Functional Requirement).
