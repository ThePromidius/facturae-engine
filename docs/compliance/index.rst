Functional Requirements (Mastermind)
=====================================

This is the developer's single source of truth. It defines the "Active Workflow" required for compliance, aggregating requirements from various laws.

.. _req_hash_chaining:

Requirement: Hash Chaining
--------------------------

To ensure record inalterability, every invoice must be cryptographically linked to its predecessor.

**Current Workflow:**
1. Retrieve the Hash of the previous record ($H_{n-1}$).
2. If it's the first record, use a block of 64 zeros.
3. Concatenate with current record data.
4. Generate SHA-256.

**Legal Basis:**
*   Linked to :ref:`hac_1177_art7` (Traceability).
*   Algorithm specified in :ref:`hac_1177_art13` (Hash Generation).

**Implementation:**
*   Source: :src:`src/internal/chain/store.go`

.. _req_signing:

Requirement: Digital Signature
------------------------------

All XML records must be signed before submission to AEAT.

**Current Workflow:**
*   Format: XAdES-BES Enveloped.
*   The signature must cover the entire `Facturae` block.

**Legal Basis:**
*   Mandated by :ref:`hac_1177_art14`.

**Implementation:**
*   Source: :src:`src/internal/signing/signer.go`
