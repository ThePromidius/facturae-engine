Chain Module
============

The chain module implements the **inalterable record chain** required for Verifactu compliance. It ensures that every invoice record is cryptographically linked to its predecessor.

Overview
--------

*   **Package:** ``internal/chain``
*   **Compliance:** Implements **Art. 7 of Orden HAC/1177/2024** (Traceability and Chaining).
*   **Source:** :src:`src/internal/chain/`

Data Structure
--------------

The core data structure is the ``Record``, which captures the state of an invoice at the moment of emission.

.. code-block:: go

   type Record struct {
       InvoiceNumber       string
       InvoiceSeries       string
       EmisorCIF           string
       IssueDate           time.Time
       Total               float64
       PreviousFingerprint string
       Fingerprint         string
       Timestamp           time.Time
   }

Fingerprint Computation
~~~~~~~~~~~~~~~~~~~~~~~

The fingerprint (huella) is computed by:
1.  **Canonicalizing** the record: ``emisorCIF|series|number|YYYY-MM-DD|total|previousFP``
2.  Generating a **SHA-256** hash of the resulting string.

Persistence (Stores)
--------------------

The module supports multiple storage backends through the ``Store`` interface.

SQLStore
~~~~~~~~

The recommended store for production. It supports **PostgreSQL** and **SQLite**.

*   **PostgreSQL:** Uses the ``pgx`` driver.
*   **SQLite:** Uses a pure Go driver (no CGO required).

Source: :src:`src/internal/chain/sql.go`

MemoryStore
~~~~~~~~~~~

An in-memory store for development and testing. **Warning:** Data is lost on restart, which is not compliant for production Verifactu use.

Source: :src:`src/internal/chain/memory.go`

Chain Operations
----------------

The ``Chain`` service provides high-level logic for managing the sequence of records.

.. code-block:: go

   func (c *Chain) Append(...) (Record, error)
   func (c *Chain) Verify() error

The ``Verify()`` method is crucial for auditors: it iterates through the entire chain, recomputes all hashes, and validates that every link is intact.

Legal Requirements (Mastermind Link)
------------------------------------

This module directly implements:
*   :ref:`hac_1177_art7` (Traceability and Chaining).
*   :ref:`hac_1177_art13` (Fingerprint Generation).
*   :ref:`req_hash_chaining` (Functional Requirement).
