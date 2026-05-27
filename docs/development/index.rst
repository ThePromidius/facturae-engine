Development Guide
=================

This guide covers everything you need to set up, build, run, and extend the
FacturaE Sidecar during development.

.. contents::
   :local:

Prerequisites
-------------

*   **Go 1.25+** — module requirement, see ``go.mod``.
*   **OpenSSL** — required for PKCS#12 certificate parsing (``-p12`` flag).
*   **xmllint** (``libxml2-utils``) — optional, XSD validation falls back gracefully.
*   **Docker** — optional, used as a fallback for dev certificate generation.
*   **PostgreSQL** — optional, for SQL chain store testing.

Setup
-----

.. code-block:: bash

   git clone https://github.com/ThePromidius/facturae-engine.git
   cd facturae-engine
   go mod download

Building
--------

.. code-block:: bash

   # Standard build
   go build -o facturae-engine ./src/cmd/facturae-engine

   # Cross-compile for Linux
   GOOS=linux GOARCH=amd64 go build -o facturae-engine-linux ./src/cmd/facturae-engine

   # Optimised production build
   CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
     go build -ldflags="-w -s" -o facturae-engine ./src/cmd/facturae-engine

Running
-------

.. list-table::
   :header-rows: 1
   :widths: 35 65

   * - Mode
     - Command
   * - Mock signing (no cert, no AEAT)
     - ``go run ./src/cmd/facturae-engine``
   * - Self-signed dev cert (auto-generated)
     - ``go run ./src/cmd/facturae-engine -dev-p12 dev.p12``
   * - Real PKCS#12 cert
     - ``go run ./src/cmd/facturae-engine -p12 cert.p12 -p12pass secret``
   * - Dev cert + AEAT test environment
     - ``go run ./src/cmd/facturae-engine -dev-p12 dev.p12 -aeat test``
   * - Real cert + AEAT production
     - ``go run ./src/cmd/facturae-engine -p12 cert.p12 -p12pass secret -aeat prod``
   * - With SQL persistence
     - ``go run ./src/cmd/facturae-engine -db postgres -dsn "postgres://..."``
   * - Custom socket
     - ``go run ./src/cmd/facturae-engine -socket 0.0.0.0:8080``

The ``-dev-p12`` flag automatically generates a 2048-bit RSA self-signed
certificate if the file does not exist. It uses Go's ``crypto/x509`` natively
and falls back to Docker (``alpine/openssl``) if OpenSSL is not available on
the host.

.. seealso::

   :doc:`certificates` for details on certificate types and the complete
   AEAT submission workflow in :doc:`/integration/aeat_submission`.

Environment Variables
---------------------

Every CLI flag has a corresponding environment variable. **Flags take
precedence over env vars.**

.. list-table::
   :header-rows: 1
   :widths: 25 30 30 15

   * - Flag
     - Environment Variable
     - Default
     - Required
   * - ``-socket``
     - ``ENGINE_SOCKET``
     - ``127.0.0.1:8080``
     - No
   * - ``-schemas``
     - ``ENGINE_SCHEMAS``
     - ``./schemas``
     - No
   * - ``-p12``
     - ``CERT_P12_PATH``
     - *(empty)*
     - For real AEAT
   * - ``-p12pass``
     - ``CERT_P12_PASS``
     - *(empty)*
     - With ``-p12``
   * - ``-dev-p12``
     - ``DEV_P12_PATH``
     - *(empty)*
     - No
   * - ``-dev-p12pass``
     - ``DEV_P12_PASS``
     - ``changeit``
     - No
   * - ``-key``
     - ``CERT_KEY_PATH``
     - *(empty)*
     - Alternative to ``-p12``
   * - ``-cert``
     - ``CERT_PEM_PATH``
     - *(empty)*
     - With ``-key``
   * - ``-aeat``
     - ``AEAT_ENV``
     - *(empty)*
     - For AEAT
   * - ``-db``
     - ``DB_DRIVER``
     - ``memory``
     - No
   * - ``-dsn``
     - ``DB_DSN``
     - *(empty)*
     - With SQL

Signer resolution order
-----------------------

The engine selects a signer using the first matching rule:

#. **``-p12``** — real PKCS#12 signer. Requires OpenSSL. Enables AEAT TLS
   client cert when ``-aeat`` is set.
#. **``-dev-p12``** — self-signed dev cert. Auto-generates if missing.
   Warns that AEAT submission will fail in production (no TLS client cert).
#. **``-key`` + ``-cert``** — PEM-based real signer.
#. **None** — ``MockSigner`` (appends a SHA-256 comment instead of a real
   signature).

Testing
-------

.. code-block:: bash

   # Run all tests
   go test ./src/...

   # Verbose output
   go test -v ./src/...

   # Specific package
   go test -v ./src/internal/api/...

   # Specific test function
   go test -v -run TestInvoice_HappyPath_Returns200 ./src/internal/api/...
   
   # With coverage
   go test -coverprofile=coverage.out ./src/...

For a detailed breakdown of every test package, see :doc:`testing`.

Adding a New Feature
--------------------

#. **Add types** in the relevant ``src/internal/<module>/`` package.
#. **Write tests** following the table-driven pattern used throughout the
   codebase. Every exported function should have a test.
#. **Wire dependencies** in ``src/cmd/facturae-engine/main.go``.
#. **Expose via API** — add a handler in ``src/internal/api/handlers.go``
   and register the route in ``src/internal/api/server.go``.
#. **Document** the new module and its API.
#. **Run all tests** before committing.

Code Conventions
----------------

*   Idiomatic Go with standard library preferred.
*   API error messages in **Spanish**; internal logs in **English**.
*   Exported types and functions must have godoc comments.
*   Table-driven tests are preferred.
*   Mock implementations for external dependencies (AEAT, FACe).
*   Zero heavy frameworks — only stdlib ``net/http``, ``database/sql``, and
    minimal third-party drivers.

Docker Compose
--------------

For local integration testing with PostgreSQL and Redis:

.. code-block:: bash

   cp .env.example .env
   # Edit .env with your settings
   docker compose up --build

The compose stack starts:

*   ``facturae-engine`` — the Go sidecar
*   ``redis-queue`` — task queue (AOF enabled)
*   ``postgres-db`` — persistent chain store

.. note::

   The ``backend_net`` is configured as ``internal: true`` (no external
   internet access) for security. The engine can still reach AEAT outbound
   but containers cannot talk to each other except on the backend network.
