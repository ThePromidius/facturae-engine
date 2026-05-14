Standalone Validation Service
=============================

The FacturaE Sidecar provides a dedicated **Validation Endpoint** that allows ERPs and developers to verify invoice structure without triggering emission or chaining.

.. contents::
   :local:

Endpoint: POST /validate
------------------------

This endpoint accepts both JSON and XML and returns a structural audit report.

**Request (JSON):**

.. code-block:: bash

   curl -X POST http://localhost:8080/validate \
     -H "Content-Type: application/json" \
     -d '{"meta": {"formato": "facturae"}, "factura": {...}}'

**Request (XML):**

.. code-block:: bash

   curl -X POST http://localhost:8080/validate \
     -H "Content-Type: application/xml" \
     --data-binary @my_invoice.xml

Validation Report (Response)
----------------------------

The response is a JSON object summarizing the findings:

.. code-block:: json

   {
     "valid": false,
     "errors": [
       "XSD Validation: Line 15: Element 'Total': [facet 'minInclusive'] The value '-10.00' is less than the minimum value allowed '0.00'."
     ],
     "type": "facturae-xml",
     "version": "3.2.2"
   }

Centralized Validation Logic
----------------------------

All internal validation is consolidated in the following module:
:src:`src/internal/validation/service.go`

Stages of Validation
~~~~~~~~~~~~~~~~~~~~

1.  **JSON Schema**: Checks for mandatory fields (CIF, Number, Series).
2.  **XML Well-formedness**: Ensures the XML is parseable.
3.  **Strict XSD Validation**: Uses the official AEAT/FacturaE schemas (e.g., ``SuministroLR.xsd``).

.. tip::
   It is highly recommended to use the ``/validate`` endpoint in the ERP development environment to ensure compatibility before deploying to production.
