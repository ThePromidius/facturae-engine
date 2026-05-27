.. _certificates:

Certificate Management
======================

This guide covers digital certificates needed for FacturaE / Verifactu
signing and AEAT communication.

.. contents::
   :local:

Certificate Types
-----------------

.. list-table::
   :header-rows: 1
   :widths: 20 30 50

   * - Type
     - Use
     - AEAT compatibility
   * - **FNMT Persona Fisica**
     - Autonomos, firmar facturas como persona fisica
     - Test + Production (mTLS)
   * - **FNMT Representante**
     - Administradores, apoderados de empresas
     - Test + Production (mTLS)
   * - **FNMT Sello de Empresa**
     - Firma automatizada, proveedores SaaS
     - Test + Production (mTLS)
   * - **FNMT Pruebas**
     - Desarrollo e integracion
     - PRE environment only (no fiscal validity)
   * - **Self-signed (dev-p12)**
     - Desarrollo local del pipeline de firma
     - **Not valid** for AEAT (no TLS client cert)

Self-Signed Certificate (Development)
-------------------------------------

The easiest way to get a dev certificate is to use the ``-dev-p12`` flag:

.. code-block:: bash

   go run ./src/cmd/facturae-engine -dev-p12 dev.p12

If ``dev.p12`` does not exist, the engine **auto-generates** a 2048-bit RSA
self-signed certificate using Go's ``crypto/x509`` (with an OpenSSL or Docker
fallback for the PKCS#12 packaging step). The generated file is reused on
subsequent runs.

To generate a dev certificate without starting the server:

.. code-block:: powershell

   # PowerShell — via OpenSSL
   openssl req -x509 -newkey rsa:2048 `
     -keyout key.pem -out cert.pem `
     -days 365 -nodes `
     -subj "/CN=sidecar-dev/O=VIVE-X"

   openssl pkcs12 -export `
     -in cert.pem -inkey key.pem `
     -out dev.p12 -passout pass:changeit

   del key.pem, cert.pem

.. code-block:: bash

   # Linux/macOS — via OpenSSL
   openssl req -x509 -newkey rsa:2048 \
     -keyout key.pem -out cert.pem \
     -days 365 -nodes \
     -subj "/CN=sidecar-dev/O=VIVE-X"

   openssl pkcs12 -export \
     -in cert.pem -inkey key.pem \
     -out dev.p12 -passout pass:changeit

   rm key.pem cert.pem

.. code-block:: powershell

   # Windows — via Docker (no OpenSSL install)
   docker run --rm -v "${PWD}:C:/out" alpine/openssl `
     req -x509 -newkey rsa:2048 `
     -keyout /out/key.pem -out /out/cert.pem `
     -days 365 -nodes `
     -subj "/CN=sidecar-dev/O=VIVE-X"

   docker run --rm -v "${PWD}:C:/out" alpine/openssl `
     pkcs12 -export `
     -in /out/cert.pem -inkey /out/key.pem `
     -out /out/dev.p12 -passout pass:changeit

.. code-block:: bash

   # Linux/macOS — via Docker
   docker run --rm -v "$(pwd):/out" alpine/openssl \
     req -x509 -newkey rsa:2048 \
     -keyout /out/key.pem -out /out/cert.pem \
     -days 365 -nodes \
     -subj "/CN=sidecar-dev/O=VIVE-X"

   docker run --rm -v "$(pwd):/out" alpine/openssl \
     pkcs12 -export \
     -in /out/cert.pem -inkey /out/key.pem \
     -out /out/dev.p12 -passout pass:changeit

Verifying a .p12 File
---------------------

.. code-block:: bash

   openssl pkcs12 -info -in dev.p12 -passin pass:changeit -noout

If the output includes ``Bag Attributes`` with ``friendlyName`` and
``localKeyID``, the file contains a private key and is ready for signing.
If you only see ``-----BEGIN CERTIFICATE-----``, the private key is missing
and the file cannot be used for signing.

Obtaining a Real FNMT Certificate
---------------------------------

#. Request at ``https://www.sede.fnmt.gob.es/certificados/persona-fisica``
   (or the relevant section for Representante / Sello de Empresa).
#. Note the **Solicitud Code** given at the end.
#. **Accredit your identity** in person at an OAI office (AEAT, Social
   Security, participating town halls). Bring your DNI/NIE + the code.
#. After 24-48 hours, download the certificate from the same website.
#. **Export the .p12 from your browser:** the private key never leaves the
   browser's certificate store, so you must export it manually:

   **Chrome/Edge:** Settings → Privacy and Security → Security → Manage
   certificates → Personal tab → select cert → Export → check **"Export
   the private key"** → PKCS#12 (.PFX) format → set a password.

   **Firefox:** Preferences → Privacy & Security → Certificates → View
   Certificates → Your Certificates tab → select → Backup → set password.

FNMT Test Certificates (PRE Environment)
----------------------------------------

The FNMT offers test certificates **without identity verification**, valid
exclusively for the AEAT pre-production (PRE) environment at
``https://prewww1.aeat.es``.

#. Go to ``https://www.sede.fnmt.gob.es/certificados/pruebas``.
#. Request and download (no in-person accreditation).
#. Export the ``.p12`` from your browser (same procedure as above).

.. warning::

   Test certificates have **no fiscal validity**. AEAT's production endpoint
   (``www1.aeat.es``) will reject them. They only work in PRE.

Mounting Certificates in Docker
-------------------------------

When using Docker, mount the ``.p12`` file as a read-only volume:

.. code-block:: bash

   docker run -p 8080:8080 \
     -v /host/path/cert.p12:/app/cert.p12:ro \
     -e CERT_P12_PATH=/app/cert.p12 \
     -e CERT_P12_PASS=your_password \
     -e AEAT_ENV=prod \
     facturae-engine:latest

Or with Docker Compose, configure ``.env``:

.. code-block:: bash

   CERT_P12_PATH=/app/certs/cert.p12
   CERT_P12_PASS=your_password
   AEAT_ENV=test
