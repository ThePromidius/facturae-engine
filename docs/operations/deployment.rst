Deployment & Operations
=======================

This guide covers how to deploy and operate the FacturaE Sidecar in various environments.

Infrastructure Requirements
---------------------------

The sidecar is designed to be lightweight but requires the following for production compliance:

*   **Persistence:** A database (PostgreSQL recommended) to maintain the inalterable Hash Chain.
*   **Security:** A PKCS#12 certificate for XAdES signing.
*   **Connectivity:** Access to Spanish Tax Agency (AEAT) SOAP endpoints.

Docker Deployment
-----------------

The recommended way to deploy the sidecar is using Docker.

**Build the image:**

.. code-block:: bash

   docker build -t facturae-engine:latest .

**Running with Docker Compose:**

We provide a ``docker-compose.yml`` in the root for quick orchestration.

.. code-block:: bash

   cp .env.example .env
   # Edit .env with your specific database and certificate details
   docker-compose up -d

Production Configuration
------------------------

For production-grade Verifactu compliance, you must mount your certificate and configure the database.

.. code-block:: bash

   docker run -p 8080:8080 \
     -v /path/to/your/cert.p12:/app/cert.p12:ro \
     -v $(pwd)/schemas:/app/schemas \
     -e ENGINE_SOCKET=0.0.0.0:8080 \
     -e CERT_P12_PATH=/app/cert.p12 \
     -e CERT_P12_PASS=your_password \
     -e AEAT_ENV=prod \
     -e DB_DRIVER=postgres \
     -e DB_DSN="postgres://user:pass@host:port/db?sslmode=require" \
     facturae-engine:latest

Configuration Reference
-----------------------

The engine supports configuration via CLI flags or environment variables. **Flags take precedence.**

+----------------+----------------------+----------+------------------------------------------+
| Flag           | Environment Variable | Required | Description                              |
+================+======================+==========+==========================================+
| ``-socket``    | ``ENGINE_SOCKET`     | Yes      | Listen address (e.g. ``0.0.0.0:8080``)   |
+----------------+----------------------+----------+------------------------------------------+
| ``-schemas``   | ``ENGINE_SCHEMAS``   | Yes      | Directory for cached XSDs                |
+----------------+----------------------+----------+------------------------------------------+
| ``-p12``       | ``CERT_P12_PATH``    | No       | Path to PKCS#12 certificate              |
+----------------+----------------------+----------+------------------------------------------+
| ``-p12pass``   | ``CERT_P12_PASS``    | No       | Password for PKCS#12                     |
+----------------+----------------------+----------+------------------------------------------+
| ``-aeat``      | ``AEAT_ENV``         | No       | AEAT Env (``test`` | ``prod``)           |
+----------------+----------------------+----------+------------------------------------------+
| ``-db``        | ``DB_DRIVER``        | No       | Backend (``memory`` | ``postgres``)      |
+----------------+----------------------+----------+------------------------------------------+
| ``-dsn``       | ``DB_DSN``           | No       | Database connection string               |
+----------------+----------------------+----------+------------------------------------------+

.. tip::
   When using the ``memory`` driver, the Hash Chain is lost on restart. This is **not compliant** for production use under Verifactu.
