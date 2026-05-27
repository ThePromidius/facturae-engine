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

+--------------------------------------------------+----------------------------------------------+
| Test Function                                    | What it tests                                |
+--------------------------------------------------+----------------------------------------------+
| ``TestIntegration_SimpleInvoice``                | Full pipeline: JSON → signed XML             |
+--------------------------------------------------+----------------------------------------------+
| ``TestIntegration_SimpleInvoice_TotalsCorrect``  | Totals (1500, 315, 1815) in XML              |
+--------------------------------------------------+----------------------------------------------+
| ``TestIntegration_SimpleInvoice_IsValidXML``     | Pre-signature XML is well-formed             |
+--------------------------------------------------+----------------------------------------------+
| ``TestIntegration_MultiIVA``                     | Multiple VAT rates in XML                    |
+--------------------------------------------------+----------------------------------------------+
| ``TestIntegration_MultiIVA_TotalsCorrect``       | Multi-IVA totals (3910, 266.6, 4176.6)      |
+--------------------------------------------------+----------------------------------------------+
| ``TestIntegration_ChainGrowsAcrossRequests``     | Different fingerprints, chain length=2       |
+--------------------------------------------------+----------------------------------------------+
| ``TestIntegration_ChainEndpoint_AfterTwoInvoices`` | Chain endpoint has 2 records              |
+--------------------------------------------------+----------------------------------------------+

internal/aeat
~~~~~~~~~~~~~

AEAT SOAP client, envelope building, retry logic, XML marshalling.

Test files: :src:`src/internal/aeat/aeat_test.go`, ``client_test.go``,
``integration_test.go`` (12 tests)

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

+--------------------------------------------------+----------------------------------------------+
| Test Function                                    | What it tests                                |
+--------------------------------------------------+----------------------------------------------+
| ``TestMockSigner_AppendsMockComment``            | Mock adds "MOCK SIGNATURE" comment           |
+--------------------------------------------------+----------------------------------------------+
| ``TestMockSigner_ContainsSHA256``                | Mock includes SHA-256 hex                    |
+--------------------------------------------------+----------------------------------------------+
| ``TestMockSigner_OutputContainsOriginalXML``     | Original XML preserved                      |
+--------------------------------------------------+----------------------------------------------+
| ``TestMockSigner_Algorithm``                     | Algorithm contains "mock"                    |
+--------------------------------------------------+----------------------------------------------+
| ``TestP12Signer_SignsAndContainsSignatureElement`` | P12 produces ``<ds:Signature>``            |
+--------------------------------------------------+----------------------------------------------+
| ``TestP12Signer_ClosingTagPreserved``            | Signed XML ends with ``</fe:Facturae>``      |
+--------------------------------------------------+----------------------------------------------+
| ``TestP12Signer_Algorithm``                      | Algorithm contains "RSA-SHA256"              |
+--------------------------------------------------+----------------------------------------------+
| ``TestNewP12SignerFromPEM_InvalidKey``           | Bad key PEM returns error                    |
+--------------------------------------------------+----------------------------------------------+
| ``TestNewP12SignerFromPEM_InvalidCert``          | Bad cert PEM returns error                   |
+--------------------------------------------------+----------------------------------------------+
| ``TestOpenSSLAvailable_ReturnsBoolean``          | Reports OpenSSL availability                 |
+--------------------------------------------------+----------------------------------------------+
| ``TestLoadFromP12File_MissingFile``              | Missing .p12 returns error                   |
+--------------------------------------------------+----------------------------------------------+
| ``TestLoadTLSCertFromP12_MissingFile``           | Missing .p12 returns error                   |
+--------------------------------------------------+----------------------------------------------+

internal/chain
~~~~~~~~~~~~~~

Verifactu fingerprint chain: canonicalisation, hashing, store operations.

Test file: :src:`src/internal/chain/chain_test.go` (2 tests)

+--------------------------------------------------+----------------------------------------------+
| Test Function                                    | What it tests                                |
+--------------------------------------------------+----------------------------------------------+
| ``TestCanonicalize``                             | Pipe-delimited format matches spec           |
+--------------------------------------------------+----------------------------------------------+
| ``TestFingerprint``                              | SHA-256 of known input matches expected      |
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

+--------------------------------------------------+----------------------------------------------+
| Test Function                                    | What it tests                                |
+--------------------------------------------------+----------------------------------------------+
| ``TestNewManager_CreatesDirectory``              | Manager creates cache dir                    |
+--------------------------------------------------+----------------------------------------------+
| ``TestManager_IsCached_FalseWhenEmpty``          | Fresh manager has nothing cached             |
+--------------------------------------------------+----------------------------------------------+
| ``TestManager_SchemaPath_DownloadsAndCaches``    | Downloads and caches XSD from known URL      |
+--------------------------------------------------+----------------------------------------------+
| ``TestManager_SchemaPath_ReturnsCachedOnSecondCall`` | Second call uses cache                  |
+--------------------------------------------------+----------------------------------------------+
| ``TestManager_SchemaPath_UnknownVersion``        | Unknown version returns error                |
+--------------------------------------------------+----------------------------------------------+
| ``TestManager_SchemaPath_ServerError``           | Server 503 returns error                     |
+--------------------------------------------------+----------------------------------------------+
| ``TestManager_ClearCache``                       | ClearCache removes all cached XSDs           |
+--------------------------------------------------+----------------------------------------------+
| ``TestVerifactuSchemaDownload``                  | Downloads verifactu-suministro schema        |
+--------------------------------------------------+----------------------------------------------+
| ``TestVerifactuXMLValidation``                   | Attempts xmllint validation (skips if missing)|
+--------------------------------------------------+----------------------------------------------+

internal/qr
~~~~~~~~~~~

QR code generation and Verifactu verification URL builder.

Test file: :src:`src/internal/qr/qr_test.go` (1 test)

+--------------------------------------------------+----------------------------------------------+
| Test Function                                    | What it tests                                |
+--------------------------------------------------+----------------------------------------------+
| ``TestVerificationURL``                          | URL params: nif, numserie, fecha, importe    |
+--------------------------------------------------+----------------------------------------------+

Test Data
---------

Sample invoice JSONs for manual API testing:

.. code-block:: bash

   curl -X POST http://127.0.0.1:8080/invoice \
     -H "Content-Type: application/json" \
     -d @src/testdata/invoice_simple.json

Available fixtures in :src:`src/testdata/`:

*   ``invoice_simple.json`` — 2 lines at 21 % IVA (total 1815)
*   ``invoice_multi_iva.json`` — 4 lines at 4/10/21/0 % (total 4176.60)
