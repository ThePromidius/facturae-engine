.. _aeat_submission:

AEAT Submission Guide
=====================

This guide describes how the "Magic Box" communicates with the Spanish Tax Agency (AEAT) using the Verifactu SOAP protocol.

.. contents::
   :local:

Submission Protocol
-------------------

Verifactu uses **SOAP 1.1** over HTTPS. Every submission requires a qualified electronic certificate for mTLS authentication and XML signing.

.. _submission_workflow:

Workflow
~~~~~~~~

1.  **XML Generation**: The engine produces the ``sum:RegFactuSistemaFacturacion`` block.
2.  **Signing**: The block is signed using the :ref:`xades_bes_profile`.
3.  **SOAP Wrapping**: The signed XML is wrapped in a standard SOAP Envelope.
4.  **mTLS Handshake**: The client connects to AEAT using the P12 certificate.
5.  **Submission**: The POST request is sent to the appropriate :ref:`aeat_endpoints`.

Endpoints
---------

.. _aeat_endpoints:

+---------------+--------------------------------------------------------------------------+
| Environment   | URL                                                                      |
+===============+==========================================================================+
| **Test**      | ``https://prewww1.aeat.es/wlpl/VERIFACTU-CONT/ws/VeriFactuSOAP``         |
+---------------+--------------------------------------------------------------------------+
| **Production**| ``https://www1.aeat.es/wlpl/VERIFACTU-CONT/ws/VeriFactuSOAP``            |
+---------------+--------------------------------------------------------------------------+

Reliability & Retries
---------------------

To ensure compliance with **Art. 8 (Accessibility)**, the system implements an exponential backoff retry strategy:

*   **Max Retries**: 3 attempts.
*   **Wait Time**: Attempt² seconds (1s, 4s, 9s).
*   **Timeout**: 45 seconds per request.

Technical Proof
---------------

The submission logic is verified by:
:src:`src/internal/aeat/integration_test.go` (``TestClient_Submit_Mock``)

.. _submission_checklist:

Compliance Checklist
--------------------

*   [ ] **SOAP Wrapping**: XML is correctly placed inside ``soapenv:Body``.
*   [ ] **mTLS Auth**: Client uses the qualified certificate for the HTTPS handshake.
*   [ ] **Retry Logic**: System handles temporary AEAT downtime gracefully.
