API Reference
=============

The FacturaE Sidecar exposes a RESTful API to handle invoice transformation, signing, and compliance chaining.

Base URL
--------

By default, the service listens on ``127.0.0.1:8080``.

.. code-block:: bash

   http://localhost:8080/

Endpoints
---------

Number Format
-------------

All numeric fields use **standard JSON number format**:

*   **Decimal separator**: ``.`` (period). Commas (``,``) are **not valid** JSON.
*   Examples: ``100.00``, ``21.0``, ``0.50``, ``3.1416``
*   Sent as JSON numbers, **not** quoted strings.

+-----------------------+-----------+------------------------------------+
| Field                 | Type      | Validation                         |
+=======================+===========+====================================+
| ``cantidad``          | number    | > 0                                |
+-----------------------+-----------+------------------------------------+
| ``precio_unitario``   | number    | >= 0                               |
+-----------------------+-----------+------------------------------------+
| ``iva_tipo``          | number    | 0 — 100 (percentage)               |
+-----------------------+-----------+------------------------------------+


POST /invoice
~~~~~~~~~~~~~

Converts a JSON invoice into a signed FacturaE XML record with Verifactu chaining.

**Request:**

.. code-block:: http

   POST /invoice HTTP/1.1
   Content-Type: application/json

.. code-block:: json

   {
     "meta": { "version_formato": "3.2.2", "moneda": "EUR" },
     "factura": { "numero": "F2024-001", "serie": "A", "fecha": "2024-03-15T00:00:00Z" },
     "emisor": { "cif": "B12345678", "nombre": "Proveedor S.L.", "pais": "ESP" },
     "receptor": { "cif": "A87654321", "nombre": "Cliente S.A.", "pais": "ESP" },
     "lineas": [
       { "desc": "Consultoria", "cantidad": 1, "precio_unitario": 100.00, "iva_tipo": 21.0 }
     ]
   }

**Response Headers:**

+------------------------------+-----------------------------------------------------------+
| Header                       | Description                                               |
+==============================+===========================================================+
| ``X-Verifactu-Fingerprint``  | SHA-256 fingerprint of the current chained record (hex).  |
+------------------------------+-----------------------------------------------------------+
| ``X-Verifactu-QR-URL``       | Verifactu verification URL.                               |
+------------------------------+-----------------------------------------------------------+

**Response Body:**

Returns the signed XML (FacturaE format).

GET /health
~~~~~~~~~~~

Returns the engine status and current chain state.

.. code-block:: json

   {
     "status": "ok",
     "chain_length": 42,
     "signer": "RSA-SHA256 / XAdES-BES",
     "timestamp": "2024-03-15T10:30:00Z"
   }

GET /chain
~~~~~~~~~~

Retrieves the complete Verifactu chain history. Use this for auditing and compliance verification.

GET /chain/verify
~~~~~~~~~~~~~~~~~

Validates the integrity of the entire Verifactu chain. Returns ``status: ok``
if every record's fingerprint is consistent, or ``status: tampered`` if the
chain has been broken or manipulated.

.. code-block:: json

   {"status": "ok", "chain_length": 5, "message": "Cadena Verifactu intacta: todos los fingerprints son consistentes"}

This endpoint runs ``Chain.Verify()`` — the same check that the engine
performs automatically before every AEAT submission.

Error Handling
--------------

All errors return a JSON object with an ``error`` field.

.. code-block:: json

   { "error": "validation failed: NIF must be 9 characters" }

+------+-----------------------------------------------------------------------+
| Code | Meaning                                                               |
+======+=======================================================================+
| 400  | Bad Request (Invalid JSON or validation failure).                     |
+------+-----------------------------------------------------------------------+
| 422  | Unprocessable Entity (Structural or XSD validation failure).          |
+------+-----------------------------------------------------------------------+
| 500  | Internal Server Error (Signing, DB, or Chaining failure).             |
+------+-----------------------------------------------------------------------+
