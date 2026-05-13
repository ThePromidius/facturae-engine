Compliance Implementation Demo (rST)
=====================================

.. meta::
   :description: Demonstration of rST capabilities for legal compliance documentation.
   :keywords: Verifactu, rST, documentation, legal

This page demonstrates why **reStructuredText (rST)** is the standard for critical technical systems.

.. contents:: Table of Contents
   :depth: 2
   :local:

Legal Reference Directives
--------------------------

In a legal context, you want visual callouts that are standardized across the whole manual.

.. important:: 
   According to **Orden HAC/1177/2024**, all hash chains must be SHA-256. 
   Failure to comply results in invalidation of the invoice record.

.. tip::
   Use the ``--verify`` flag in our CLI to check the hash chain locally before submission.

Advanced Cross-Referencing
--------------------------

Unlike Markdown, rST allows "Roles". I can define a link to a specific part of the law or our code.

*   **Internal Link:** Read about our :ref:`legacy-hashing` (This links to the exact section even if the filename changes).
*   **Source Code Reference:** 

    In Sphinx, we can link directly to the source code files. This is crucial for auditors to verify that the law is correctly implemented in the logic.

    *   **Logic for Verifactu Signatures:** :src:`src/internal/signing/signer.go`
    *   **AEAT Soap Client:** :src:`src/internal/aeat/client.go`

    We can even link to specific lines if we know them: :src:`Chain Types <src/internal/chain/types.go#L10>`.

Technical Specifications (Tables)
---------------------------------

rST handles complex data structures much better than Markdown "pipe" tables.

+-----------------------+-----------------------+------------------------------------------+
| Field Name            | Requirement           | Implementation Status                    |
+=======================+=======================+==========================================+
| NIF Emisor            | Mandatory             | .. code-block:: go                       |
|                       |                       |                                          |
|                       |                       |    type Invoice struct { NIF string }    |
+-----------------------+-----------------------+------------------------------------------+
| Hash Chain            | Sequential            | :orange:`Pending Audit`                  |
+-----------------------+-----------------------+------------------------------------------+

.. role:: orange
   :class: orange-text

Mathematical Formulas
---------------------

For Verifactu, we often need to document the exact concatenation logic for the Fingerprint (Huella). 

The hash :math:`H_n` is calculated as:

.. math::

   H_n = \text{SHA256}(NIF_{emisor} + ID_{factura} + H_{n-1} + \text{Timestamp})

Footnotes and Citations
-----------------------

You can cite the BOE (Spanish Official Gazette) formally [BOE2024]_.

.. [BOE2024] Orden HAC/1177/2024, de 17 de octubre, por la que se desarrollan las especificaciones técnicas, funcionales y de contenido consultables en la sede electrónica de la AEAT.

Todo Tracking
-------------

You can keep track of what's missing directly in the docs.

.. todo:: 
   Update the XML signing logic to support XAdES-BES as per latest AEAT patch.

.. section-labels
.. _legacy-hashing:

Appendix: Legacy Hashing Notes
------------------------------
This is a target for the cross-reference above.
