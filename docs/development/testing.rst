Testing Reference
=================

This page documents every test package in the project, what each covers, and
how to run specific tests.

.. contents::
   :local:

Quick Start
-----------

.. code-block:: bash

   # All tests
   go test ./src/...

   # Verbose (shows test names and pass/fail per function)
   go test -v ./src/...

   # Single package
   go test -v ./src/internal/api/...

   # Single test function (regex pattern)
   go test -v -run TestInvoice_HappyPath_Returns200 ./src/internal/api/...

   # Coverage report
   go test -coverprofile=coverage.out ./src/...
   go tool cover -html=coverage.out

Test Packages
-------------

internal/api
~~~~~~~~~~~~

HTTP handlers, server lifecycle, QR endpoint.

Tests file: :src:`src/internal/api/server_test.go` (14 tests)

+--------------------------------------------------+----------------------------------------------+
| Test Function                                    | What it tests                                |
+==================================================+==============================================+
| ``TestInvoice_HappyPath_Returns200``             | Valid POST returns 200                       |
+--------------------------------------------------+----------------------------------------------+
| ``TestInvoice_ResponseIsXML``                    | Content-Type is ``application/xml``          |
+--------------------------------------------------+----------------------------------------------+
| ``TestInvoice_ResponseContainsFacturaeRoot``     | Response contains ``<Facturae>`` root        |
+--------------------------------------------------+----------------------------------------------+
| ``TestInvoice_ResponseContainsMockSignature``    | Contains ``MOCK SIGNATURE`` comment          |
+--------------------------------------------------+----------------------------------------------+
| ``TestInvoice_ResponseHasVerifactuHeader``       | ``X-Verifactu-Fingerprint`` is 64 hex chars  |
+--------------------------------------------------+----------------------------------------------+
| ``TestInvoice_ChainLengthIncrements``            | ``X-Chain-Length`` counts across invoices    |
+--------------------------------------------------+----------------------------------------------+
| ``TestInvoice_InvalidJSON_Returns400``           | Bad JSON                                     |
+--------------------------------------------------+----------------------------------------------+
| ``TestInvoice_MissingEmisor_Returns400``         | Missing emisor CIF                           |
+--------------------------------------------------+----------------------------------------------+
| ``TestInvoice_EmptyLineas_Returns400``           | Empty lineas array                           |
+--------------------------------------------------+----------------------------------------------+
| ``TestInvoice_WrongMethod_Returns405``           | GET on ``/invoice``                          |
+--------------------------------------------------+----------------------------------------------+
| ``TestInvoice_MultipleVATBrackets_XML``          | Multiple VAT brackets in XML output          |
+--------------------------------------------------+----------------------------------------------+
| ``TestHealth_Returns200``                        | ``GET /health``                              |
+--------------------------------------------------+----------------------------------------------+
| ``TestHealth_ResponseJSON``                      | Health response is valid JSON                |
+--------------------------------------------------+----------------------------------------------+
| ``TestChain_EmptyOnStart``                       | ``GET /chain`` returns count=0               |
+--------------------------------------------------+----------------------------------------------+
| ``TestChain_AfterInvoice_HasRecord``             | Chain has record after POST                  |
+--------------------------------------------------+----------------------------------------------+

QR endpoint tests file: :src:`src/internal/api/qr_test.go` (5 tests)

+--------------------------------------------------+----------------------------------------------+
| Test Function                                    | What it tests                                |
+--------------------------------------------------+----------------------------------------------+
| ``TestQREndpoint_ValidText``                     | ``/qr?text=...`` returns 200                 |
+--------------------------------------------------+----------------------------------------------+
| ``TestQREndpoint_ReturnsPNG``                    | Response is valid PNG                        |
+--------------------------------------------------+----------------------------------------------+
| ``TestQREndpoint_MissingText_Returns400``        | ``/qr`` without text                         |
+--------------------------------------------------+----------------------------------------------+
| ``TestInvoice_ResponseHasQRURL``                 | ``X-Verifactu-QR-URL`` header present        |
+--------------------------------------------------+----------------------------------------------+
| ``TestInvoice_QRDataURIHeader``                  | ``X-Verifactu-QR-DataURI`` starts with data  |
+--------------------------------------------------+----------------------------------------------+

API Integration tests file: :src:`src/internal/api/integration_test.go` (7 tests)

+----------------------------------------------------+----------------------------------------------+
| Test Function                                      | What it tests                                |
+----------------------------------------------------+----------------------------------------------+
| ``TestIntegration_SimpleInvoice``                  | Full pipeline: JSON → signed XML             |
+----------------------------------------------------+----------------------------------------------+
| ``TestIntegration_SimpleInvoice_TotalsCorrect``    | Totals (1500, 315, 1815) in XML              |
+----------------------------------------------------+----------------------------------------------+
| ``TestIntegration_SimpleInvoice_IsValidXML``       | Pre-signature XML is well-formed             |
+----------------------------------------------------+----------------------------------------------+
| ``TestIntegration_MultiIVA``                       | Multiple VAT rates in XML                    |
+----------------------------------------------------+----------------------------------------------+
| ``TestIntegration_MultiIVA_TotalsCorrect``         | Multi-IVA totals (3910, 266.6, 4176.6)       |
+----------------------------------------------------+----------------------------------------------+
| ``TestIntegration_ChainGrowsAcrossRequests``       | Different fingerprints, chain length=2       |
+----------------------------------------------------+----------------------------------------------+
| ``TestIntegration_ChainEndpoint_AfterTwoInvoices`` | Chain endpoint has 2 records                 |
+----------------------------------------------------+----------------------------------------------+

internal/aeat
~~~~~~~~~~~~~

AEAT SOAP client, envelope building, retry logic, XML marshalling,
SuministroLR builder (``build.go``), Verifactu format helpers.

Test files: :src:`src/internal/aeat/aeat_test.go` (30 tests),
``client_test.go`` (6 tests), ``integration_test.go`` (6 tests)

+--------------------------------------------------+----------------------------------------------+
| Test Function                                    | What it tests                                |
+==================================================+==============================================+
| ``TestBuildSuministroLR_PrimerRegistro``         | PrimerRegistro=true when chain is empty      |
+--------------------------------------------------+----------------------------------------------+
| ``TestBuildSuministroLR_RegistroAnterior``       | RegistroAnterior set from chain prev record  |
+--------------------------------------------------+----------------------------------------------+
| ``TestBuildSuministroLR_ZeroAmounts``            | Zero cuota/importe handled correctly         |
+--------------------------------------------------+----------------------------------------------+
| ``TestBuildSuministroLR_LargeAmounts``           | Large float64 values, precision preserved    |
+--------------------------------------------------+----------------------------------------------+
| ``TestBuildSuministroLR_MultipleTaxBrackets``    | 3 tax brackets → 3 Desglose entries          |
+--------------------------------------------------+----------------------------------------------+
| ``TestBuildSuministroLR_WithSignature``          | ds:Signature embedded in RegistroFactura     |
+--------------------------------------------------+----------------------------------------------+
| ``TestBuildSuministroLR_WithoutSignature``       | Nil signature → omitted from XML             |
+--------------------------------------------------+----------------------------------------------+
| ``TestBuildSuministroLR_CustomConfig``           | Custom config overrides defaults             |
+--------------------------------------------------+----------------------------------------------+
| ``TestFormatAmount``                             | Positive, negative, zero, large, small       |
+--------------------------------------------------+----------------------------------------------+
| ``TestFormatFechaDDMMYYYY``                      | Valid, already DD-MM, empty, partial         |
+--------------------------------------------------+----------------------------------------------+
| ``TestMapInvoiceType``                           | FC→F1, FA→F2, AF→R1, unknown→F1, empty→F1   |
+--------------------------------------------------+----------------------------------------------+
| ``TestXMLMarshalling``                           | SuministroLR XML contains expected nodes     |
+--------------------------------------------------+----------------------------------------------+
| ``TestClient_SuccessfulSubmission``              | Submit returns CSV, ``IsAccepted()=true``    |
+--------------------------------------------------+----------------------------------------------+
| ``TestClient_RejectedByAEAT``                    | Estado ``Incorrecto`` → rejected             |
+--------------------------------------------------+----------------------------------------------+
| ``TestClient_HTTPError_Retries``                 | 3 retries on 503, succeeds on 3rd            |
+--------------------------------------------------+----------------------------------------------+
| ``TestClient_ContextCancellation``               | Cancelled context returns error              |
+--------------------------------------------------+----------------------------------------------+
| ``TestSOAPEnvelope_ContainsCIF``                 | Captured request is valid SOAP               |
+--------------------------------------------------+----------------------------------------------+
| ``TestSubmitResult_IsAccepted``                  | State table: Correcto/AceptadoConErrores     |
+--------------------------------------------------+----------------------------------------------+
| ``TestEndpoints_BothEnvironmentsDefined``        | Test and prod endpoints are non-empty        |
+--------------------------------------------------+----------------------------------------------+
| ``TestSOAPResponseParsing``                      | XML unmarshal yields correct CSV             |
+--------------------------------------------------+----------------------------------------------+
| ``TestClient_Submit_Mock``                       | Full mock AEAT round-trip (integration)      |
+--------------------------------------------------+----------------------------------------------+
| Test Function                                    | What it tests                                |
+--------------------------------------------------+----------------------------------------------+
| ``TestXMLMarshalling``                           | SuministroLR XML contains expected nodes     |
+--------------------------------------------------+----------------------------------------------+
| ``TestClient_SuccessfulSubmission``              | Submit returns CSV, ``IsAccepted()=true``    |
+--------------------------------------------------+----------------------------------------------+
| ``TestClient_RejectedByAEAT``                    | Estado ``Incorrecto`` → rejected             |
+--------------------------------------------------+----------------------------------------------+
| ``TestClient_HTTPError_Retries``                 | 3 retries on 503, succeeds on 3rd            |
+--------------------------------------------------+----------------------------------------------+
| ``TestClient_ContextCancellation``               | Cancelled context returns error              |
+--------------------------------------------------+----------------------------------------------+
| ``TestSOAPEnvelope_ContainsCIF``                 | Captured request is valid SOAP               |
+--------------------------------------------------+----------------------------------------------+
| ``TestSubmitResult_IsAccepted``                  | State table: Correcto/AceptadoConErrores     |
+--------------------------------------------------+----------------------------------------------+
| ``TestEndpoints_BothEnvironmentsDefined``        | Test and prod endpoints are non-empty        |
+--------------------------------------------------+----------------------------------------------+
| ``TestSOAPResponseParsing``                      | XML unmarshal yields correct CSV             |
+--------------------------------------------------+----------------------------------------------+
| ``TestClient_Submit_Mock``                       | Full mock AEAT round-trip (integration)      |
+--------------------------------------------------+----------------------------------------------+

internal/signing
~~~~~~~~~~~~~~~~

Signer interface implementations (MockSigner, P12Signer), PKCS#12 loading,
dev certificate generation.

Test files: :src:`src/internal/signing/signer_test.go` (11 tests),
``pkcs12_test.go`` (3 tests)

+----------------------------------------------------+----------------------------------------------+
| Test Function                                      | What it tests                                |
+----------------------------------------------------+----------------------------------------------+
| ``TestMockSigner_AppendsMockComment``              | Mock adds "MOCK SIGNATURE" comment           |
+----------------------------------------------------+----------------------------------------------+
| ``TestMockSigner_ContainsSHA256``                  | Mock includes SHA-256 hex                    |
+----------------------------------------------------+----------------------------------------------+
| ``TestMockSigner_OutputContainsOriginalXML``       | Original XML preserved                       |
+----------------------------------------------------+----------------------------------------------+
| ``TestMockSigner_Algorithm``                       | Algorithm contains "mock"                    |
+----------------------------------------------------+----------------------------------------------+
| ``TestP12Signer_SignsAndContainsSignatureElement`` | P12 produces ``<ds:Signature>``              |
+----------------------------------------------------+----------------------------------------------+
| ``TestP12Signer_ClosingTagPreserved``              | Signed XML ends with ``</fe:Facturae>``      |
+----------------------------------------------------+----------------------------------------------+
| ``TestP12Signer_Algorithm``                        | Algorithm contains "RSA-SHA256"              |
+----------------------------------------------------+----------------------------------------------+
| ``TestNewP12SignerFromPEM_InvalidKey``             | Bad key PEM returns error                    |
+----------------------------------------------------+----------------------------------------------+
| ``TestNewP12SignerFromPEM_InvalidCert``            | Bad cert PEM returns error                   |
+----------------------------------------------------+----------------------------------------------+
| ``TestOpenSSLAvailable_ReturnsBoolean``            | Reports OpenSSL availability                 |
+----------------------------------------------------+----------------------------------------------+
| ``TestLoadFromP12File_MissingFile``                | Missing .p12 returns error                   |
+----------------------------------------------------+----------------------------------------------+
| ``TestLoadTLSCertFromP12_MissingFile``             | Missing .p12 returns error                   |
+----------------------------------------------------+----------------------------------------------+

internal/chain
~~~~~~~~~~~~~~

Verifactu fingerprint chain: canonicalisation, hashing, store operations,
multi-tenant chaining, tamper detection, concurrent append, SQL persistence.

Test file: :src:`src/internal/chain/store_test.go` (22 edge case tests)

+--------------------------------------------------+----------------------------------------------+
| Test Function                                    | What it tests                                |
+==================================================+==============================================+
| ``TestChain_EmptyChain``                         | Verify on empty chain returns OK             |
+--------------------------------------------------+----------------------------------------------+
| ``TestChain_SingleCIF``                          | Single tenant: 3 records link correctly      |
+--------------------------------------------------+----------------------------------------------+
| ``TestChain_TwoCIFs``                            | Multi-tenant: 2 CIFs, 6 records, grouped     |
+--------------------------------------------------+----------------------------------------------+
| ``TestChain_TamperDetection``                    | Mid-chain record edit → Verify fails         |
+--------------------------------------------------+----------------------------------------------+
| ``TestChain_GapDetection``                       | Missing record breaks chaining               |
+--------------------------------------------------+----------------------------------------------+
| ``TestChain_ConcurrentAppend``                   | 5 goroutines, 5 CIFs, no race conditions     |
+--------------------------------------------------+----------------------------------------------+
| ``TestChain_ExtremeValues``                      | 9 sub-cases: negative amounts, max float64   |
+--------------------------------------------------+----------------------------------------------+
| ``TestChain_FingerprintDeterminism``             | Same input always produces same fingerprint  |
+--------------------------------------------------+----------------------------------------------+
| ``TestChain_CanonicalizeConsistency``            | Timestamp format is UTC, stable              |
+--------------------------------------------------+----------------------------------------------+

internal/facturae
~~~~~~~~~~~~~~~~~

FacturaE XML builder, line computation, structural validation, UBL interop.

Test files: :src:`src/internal/facturae/builder_test.go` (21 tests),
``interop_test.go`` (1 test)

+--------------------------------------------------+----------------------------------------------+
| Test Function                                    | What it tests                                |
+--------------------------------------------------+----------------------------------------------+
| ``TestBuild_StructureIsCorrect``                 | SchemaVersion=3.2.2, correct CIFs            |
+--------------------------------------------------+----------------------------------------------+
| ``TestBuild_TotalsAreCorrect``                   | Gross=1050, Tax=220.50, Total=1270.50        |
+--------------------------------------------------+----------------------------------------------+
| ``TestBuild_MultipleVATBrackets``                | 3 tax brackets, Total=335.00                 |
+--------------------------------------------------+----------------------------------------------+
| ``TestBuild_ZeroVAT``                            | 0% tax=0, total=1000                         |
+--------------------------------------------------+----------------------------------------------+
| ``TestBuild_IssueDate``                          | IssueDate="2024-05-12"                       |
+--------------------------------------------------+----------------------------------------------+
| ``TestBuild_SellerAddress``                      | Seller has address with PostCode             |
+--------------------------------------------------+----------------------------------------------+
| ``TestBuild_BatchCounterMatchesInvoices``        | Batch.InvoicesCount matches                  |
+--------------------------------------------------+----------------------------------------------+
| ``TestValidateStruct_HappyPath``                 | Valid struct passes                          |
+--------------------------------------------------+----------------------------------------------+
| (8 more validation tests)                        | SchemaVersion, negative totals, XML checks   |
+--------------------------------------------------+----------------------------------------------+
| ``TestInteroperability_UBL_and_FacturaE``        | Same Request builds both formats (interop)   |
+--------------------------------------------------+----------------------------------------------+

internal/invoice
~~~~~~~~~~~~~~~~

JSON request validation rules (required fields, ranges).

Test file: :src:`src/internal/invoice/validation_test.go` (9 tests)

+--------------------------------------------------+----------------------------------------------+
| Test Function                                    | What it tests                                |
+--------------------------------------------------+----------------------------------------------+
| ``TestValidate_HappyPath``                       | Valid request passes                         |
+--------------------------------------------------+----------------------------------------------+
| ``TestValidate_MissingNumero``                   | Missing invoice number                       |
+--------------------------------------------------+----------------------------------------------+
| ``TestValidate_MissingEmisorCIF``                | Missing emisor CIF                           |
+--------------------------------------------------+----------------------------------------------+
| ``TestValidate_MissingEmisorDireccion``          | Missing emisor address                       |
+--------------------------------------------------+----------------------------------------------+
| ``TestValidate_EmptyLineas``                     | Empty lineas array                           |
+--------------------------------------------------+----------------------------------------------+
| ``TestValidate_NegativeCantidad``                | Negative quantity                            |
+--------------------------------------------------+----------------------------------------------+
| ``TestValidate_InvalidIVA``                      | IVA > 100                                    |
+--------------------------------------------------+----------------------------------------------+
| ``TestValidate_MultipleErrors``                  | Multiple errors collected                    |
+--------------------------------------------------+----------------------------------------------+
| ``TestDefaultMeta``                              | Defaults: version=3.2.2, currency=EUR        |
+--------------------------------------------------+----------------------------------------------+

internal/legal
~~~~~~~~~~~~~~

Verifactu legal compliance enforcement: hash chaining, XAdES-BES profile,
event logging, responsible declaration.

Test files: :src:`src/internal/legal/compliance_test.go` (4 tests),
``legal_test.go`` (1 test)

+--------------------------------------------------+----------------------------------------------+
| Test Function                                    | What it tests                                |
+--------------------------------------------------+----------------------------------------------+
| ``TestRequirement_HashChaining``                 | Art. 7: records link via prev fingerprint    |
+--------------------------------------------------+----------------------------------------------+
| ``TestRequirement_FingerprintFormula``           | Art. 13: pipe-delimited format               |
+--------------------------------------------------+----------------------------------------------+
| ``TestRequirement_XAdES_BES_Profile``            | Art. 14: mandatory QualifyingProperties      |
+--------------------------------------------------+----------------------------------------------+
| ``TestRequirement_EventLogging``                 | Art. 9: LogEvent succeeds                    |
+--------------------------------------------------+----------------------------------------------+
| ``TestGenerateDeclaracionResponsable``           | Declaration has all required phrases         |
+--------------------------------------------------+----------------------------------------------+

internal/schema
~~~~~~~~~~~~~~~

XSD schema manager: download, cache, validation via ``xmllint``.

Test files: :src:`src/internal/schema/manager_test.go` (7 tests),
``verifactu_test.go`` (2 tests)

+------------------------------------------------------+------------------------------------------------+
| Test Function                                        | What it tests                                  |
+------------------------------------------------------+------------------------------------------------+
| ``TestNewManager_CreatesDirectory``                  | Manager creates cache dir                      |
+------------------------------------------------------+------------------------------------------------+
| ``TestManager_IsCached_FalseWhenEmpty``              | Fresh manager has nothing cached               |
+------------------------------------------------------+------------------------------------------------+
| ``TestManager_SchemaPath_DownloadsAndCaches``        | Downloads and caches XSD from known URL        |
+------------------------------------------------------+------------------------------------------------+
| ``TestManager_SchemaPath_ReturnsCachedOnSecondCall`` | Second call uses cache                         |
+------------------------------------------------------+------------------------------------------------+
| ``TestManager_SchemaPath_UnknownVersion``            | Unknown version returns error                  |
+------------------------------------------------------+------------------------------------------------+
| ``TestManager_SchemaPath_ServerError``               | Server 503 returns error                       |
+------------------------------------------------------+------------------------------------------------+
| ``TestManager_ClearCache``                           | ClearCache removes all cached XSDs             |
+------------------------------------------------------+------------------------------------------------+
| ``TestVerifactuSchemaDownload``                      | Downloads verifactu-suministro schema          |
+------------------------------------------------------+------------------------------------------------+
| ``TestVerifactuXMLValidation``                       | Attempts xmllint validation (skips if missing) |
+------------------------------------------------------+------------------------------------------------+

internal/qr
~~~~~~~~~~~

QR code generation and Verifactu verification URL builder.

Test file: :src:`src/internal/qr/qr_test.go` (1 test)

+--------------------------------------------------+----------------------------------------------+
| Test Function                                    | What it tests                                |
+--------------------------------------------------+----------------------------------------------+
| ``TestVerificationURL``                          | URL params: nif, numserie, fecha, importe    |
+--------------------------------------------------+----------------------------------------------+

Manual Pipeline Test (CLI)
--------------------------

Quick way to test the full pipeline end-to-end:

.. code-block:: bash

   # 1. Build and start the engine (mock mode, no cert needed)
   go build -o .tmp/facturae-engine ./src/cmd/facturae-engine
   .tmp/facturae-engine -socket 127.0.0.1:8080 &

   # 2. Wait for startup, then POST an invoice
   sleep 2
   curl -X POST http://127.0.0.1:8080/invoice \
     -H "Content-Type: application/json" \
     -d @src/testdata/invoice_simple.json

   # 3. Check the chain endpoint
   curl http://127.0.0.1:8080/chain

   # 4. Health check
   curl http://127.0.0.1:8080/health

   # 5. Stop
   kill %1

Available test data in :src:`src/testdata/`:

*   ``invoice_simple.json`` — 2 lines at 21 % IVA (total 1815)
*   ``invoice_multi_iva.json`` — 4 lines at 4/10/21/0 % (total 4176.60)

Test Scripts
------------

Pre-built PowerShell scripts for automated manual testing (Windows):

.. code-block:: powershell

   # Validate invoice pipeline (XML, signature, headers)
   .\scripts\test-invoice.ps1

   # Validate chain integrity (fingerprint linking)
   .\scripts\test-chain.ps1

   # Real submission against AEAT PRE (requires FNMT .p12)
   .\scripts\test-aeat-pre.ps1 -P12Path "certificado.p12" -P12Pass "contraseña"

Bash equivalents for Docker/Linux/CI:

.. code-block:: bash

   # All integration scripts in sequence
   ./scripts/run-all.sh test-invoice.json /tmp/engine.sock 9094

   # Or run individual scripts:
   ./scripts/test-invoice.sh src/testdata/invoice_simple.json /tmp/engine.sock 9094
   ./scripts/test-chain.sh /tmp/engine.sock 9094
   ./scripts/test-chain-verify.sh /tmp/engine.sock 9094
   ./scripts/test-graceful-shutdown.sh src/testdata/invoice_simple.json /tmp/engine.sock 9094

All ``test-invoice``, ``test-chain``, and ``test-graceful-shutdown`` run
with mock signing — no certificate needed. They start the engine, run
assertions, and clean up.

Graceful Shutdown Test
~~~~~~~~~~~~~~~~~~~~~~

The ``test-graceful-shutdown.sh`` script verifies that the engine survives
a SIGTERM and preserves the Verifactu chain across restarts:

.. code-block:: bash

   ./scripts/test-graceful-shutdown.sh src/testdata/invoice_simple.json /tmp/engine.sock 9094

It performs these steps:

#.  Starts the engine with SQLite persistence.
#.  Sends an invoice and records the chain fingerprint.
#.  Sends SIGTERM to the engine and waits for clean shutdown.
#.  Restarts the engine with the same database.
#.  Verifies the chain still contains the first invoice's record.
#.  Sends a second invoice and verifies the chain grows (length=2).

Chain Verification
~~~~~~~~~~~~~~~~~

The engine exposes ``GET /chain/verify`` to check the integrity of the
entire Verifactu chain. It runs ``Chain.Verify()`` which walks every
record and validates fingerprint continuity:

.. code-block:: bash

   curl http://127.0.0.1:8080/chain/verify

Response (intact chain):

.. code-block:: json

   {"status": "ok", "chain_length": 5, "message": "Cadena Verifactu intacta: todos los fingerprints son consistentes"}

Response (corrupted chain):

.. code-block:: json

   {"status": "tampered", "error": "record 2 (F2024-003): previous fingerprint mismatch (chain broken)", "message": "La cadena Verifactu esta corrupta o ha sido manipulada"}

Pre-Submission Verification
~~~~~~~~~~~~~~~~~~~~~~~~~~~

Before every AEAT submission, the engine runs ``Chain.Verify()``
automatically. If the chain is corrupted, the submission is **cancelled**
and logged — the invoice will never reach AEAT with a broken chain.
This prevents sending inconsistent data.

See :src:`src/internal/api/handlers.go` — the AEAT goroutine calls
``s.chain.Verify()`` before ``s.aeat.Submit()``.

Chain Verify Script
~~~~~~~~~~~~~~~~~~~

.. code-block:: powershell

   # Verify chain integrity with 2 invoices (PowerShell / Windows)
   .\scripts\test-chain-verify.ps1

   # With SQL persistence (chain survives restart)
   .\scripts\test-chain-verify.ps1 -UseSQL

.. code-block:: bash

   # Bash equivalent for Docker/Linux (uses run-all orchestrator)
   ./scripts/test-chain-verify.sh /tmp/engine.sock 9094

Each script runs 5 tests:

#.  Empty chain reports OK.
#.  Two invoices POST successfully.
#.  ``GET /chain/verify`` reports OK with chain_length = 2.
#.  Manual fingerprint validation (PreviousFingerprint linking).
#.  With ``-UseSQL`` (PowerShell) or SQLite mode (bash): restarts the server
    and verifies persistence.

The ``run-all.sh`` orchestrator runs all bash scripts in sequence:

.. code-block:: bash

   cd scripts
   ./run-all.sh src/testdata/invoice_simple.json /tmp/engine.sock 9094

This runs test-invoice, test-chain, test-chain-verify, and test-graceful-shutdown
in sequence, with a single engine instance (restarted for the graceful-shutdown
test on a separate port).

AEAT PRE Test
~~~~~~~~~~~~~

The ``test-aeat-pre.ps1`` script sends a real invoice to AEAT's test
environment (``prewww1.aeat.es``) using your FNMT certificate:

.. code-block:: powershell

   .\scripts\test-aeat-pre.ps1 -P12Path ".\micert.p12" -P12Pass "mipass"

It will:

#.  Start the engine with your certificate and ``-aeat test``.
#.  Send an invoice (the CIF in the JSON must match your cert's NIF —
    use the ``-CIF`` parameter to override it).
#.  Wait for AEAT's async response and print the result (CSV, status).
#.  Show the full server log so you can see the SOAP exchange.

If the invoice's CIF doesn't match your certificate's NIF, AEAT will
reject it. Override it with:

.. code-block:: powershell

   .\scripts\test-aeat-pre.ps1 -P12Path "cert.p12" -P12Pass "pass" -CIF "12345678Z"

.. warning::

   **AEAT mantiene la cadena por NIF.** Si envías facturas sueltas en
   PRE, cada envío debe continuar la cadena del NIF que usas. Si pierdes
   la continuidad (p.ej., re-arrancas el engine sin persistencia SQL),
   AEAT rechazará la factura porque el ``PreviousFingerprint`` no
   coincidirá con el último registro que AEAT tiene para ese NIF.

   Para pruebas aisladas usa un NIF de prueba diferente cada vez, o
   usa persistencia SQL (``-db sqlite``) para mantener la cadena entre
   sesiones.

AEAT Mock Server
----------------

The engine's **AEAT submission** tests (package ``internal/aeat``) use an
**AEAT mock server** — they never connect to the real AEAT endpoint. The
mock:

*   Listens on a random local port (no network access).
*   Responds with a valid SOAP ``Correcto`` envelope.
*   Never sends data to ``prewww1.aeat.es`` or ``www1.aeat.es``.

No automated test in this project calls the real AEAT API.

.. warning::

   Real AEAT submission (PRE or PROD) requires a valid FNMT-issued
   certificate for mTLS. The ``-dev-p12`` flag generates a **self-signed**
   cert that will **not** work against AEAT.
   See :doc:`/operations/certificates` for details.

XSD Caching
-----------

**Sí, hay caché.** Los esquemas XSD se descargan una sola vez y se
guardan en el directorio ``./schemas/`` (configurable con ``-schemas``
o ``ENGINE_SCHEMAS``). En peticiones posteriores, el motor usa el
archivo local sin hacer ninguna llamada HTTP.

Ver :src:`src/internal/schema/manager.go` — ``SchemaPath()`` comprueba
``os.Stat()`` antes de descargar.

Para limpiar la caché:

.. code-block:: bash

   rm -rf ./schemas/

XSD Validation Against Official Schemas
---------------------------------------

Every ``POST /invoice`` request validates the generated XML against the
official FacturaE XSD schema using ``xmllint``. The engine also downloads
Verifactu XSD schemas from AEAT's own servers (``prewww2.aeat.es``).

Schemas downloaded:

*   ``facturae.gob.es`` → FacturaE 3.2.2, 3.2.1
*   ``prewww2.aeat.es`` → Verifactu SuministroLR, Informacion, Respuesta,
    EventosSIF, RespuestaAnulacion

This is the **same structural validation** that AEAT performs on receipt.
If the XML passes the published XSD, the structure matches AEAT's
expectations.

Signature Validation
-------------------

The generated XML is signed with **XAdES-BES** (XAdES Basic Electronic
Signature), an ETSI standard. The engine validates:

#.  The ``<ds:Signature>`` element is present and well-formed.
#.  The ``<ds:SignedInfo>`` contains the correct canonicalisation and
    digest algorithm (SHA-256).
#.  The ``<QualifyingProperties>`` block complies with the Verifactu
    XAdES-BES profile (per Art. 14 Orden HAC/1177/2024).

These checks are verified by:
:src:`src/internal/legal/compliance_test.go` (``TestRequirement_XAdES_BES_Profile``)
