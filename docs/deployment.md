---

## Deployment Considerations

### Docker

1. **Build the image:**
   ```bash
   docker build -t facturae-engine:latest .
   ```
   The included `Dockerfile` builds a multi-stage Go application.

2. **Running the container:**
   - **Development (Mock Signing):**
     ```bash
     docker run -p 8080:8080 -v $(pwd)/schemas:/app/schemas facturae-engine:latest \
       ./facturae-engine -socket 0.0.0.0:8080 -schemas /app/schemas
     ```
   - **Production (Real Signing + AEAT Submission):**
     Requires mounting a PKCS#12 certificate (`.p12`) and its password.
     ```bash
     docker run -p 8080:8080 \
       -v /path/to/your/cert.p12:/app/cert.p12:ro \
       -v $(pwd)/schemas:/app/schemas \
       -v /path/to/secrets:/run/secrets \
       facturae-engine:latest \
       ./facturae-engine \
         -socket 0.0.0.0:8080 \
         -schemas /app/schemas \
         -p12 /app/cert.p12 \
         -p12pass $(cat /run/secrets/cert_password) \
         -aeat prod \
         -db postgres \
         -dsn "postgres://user:pass@host:port/db?sslmode=require"
     ```

### Configuration Flags

| Flag | Required | Description |
|---|---|---|
| `-socket` | Yes | Listen address. `0.0.0.0:8080` for Docker, `127.0.0.1:8080` for sidecar, or Unix socket path. |
| `-schemas` | Yes | Directory for cached XSDs. Persistent volume recommended. |
| `-p12` | Conditionally | Path to PKCS#12 cert (or `-key` + `-cert` for PEM). |
| `-p12pass` | If `-p12` | Password for the PKCS#12 file. |
| `-aeat` | No | Enable AEAT submission (`test` or `prod`). |
| `-db` | No | Chain store backend (`memory`, `postgres`, `sqlite`). Default: `memory`. |
| `-dsn` | If `-db` is not `memory` | Database connection string. |

---

## Development Workflow

1. **Clone the repository:**
   ```bash
   git clone https://github.com/ThePromidius/facturae-engine-v2.git
   cd facturae-engine-v2
   ```

2. **Set up Go environment:** Ensure Go 1.25+ is installed.

3. **Download dependencies:**
   ```bash
   go mod download
   ```

4. **Build the application:**
   ```bash
   go build -o facturae-engine ./cmd/server/main.go
   ```

5. **Run in development mode (mock signing):**
   ```bash
   ./facturae-engine -socket 127.0.0.1:8080 -schemas ./schemas
   ```

6. **Testing:**
   ```bash
   go test ./... -v
   ```
   
   Tests cover module logic, API endpoints, and integration scenarios. Use `testdata/` for sample JSON payloads.

---

## License

This project is licensed under the Business Source License 1.1 - see the `LICENSE` file in the root for details.
