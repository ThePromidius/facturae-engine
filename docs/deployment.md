---

## Deployment Considerations

### Docker

1. **Build the image:**
   ```bash
   docker build -t facturae-engine:latest .
   ```
   The included `Dockerfile` builds a multi-stage Go application.

2. **Running with Docker Compose (Recommended):**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   docker-compose up -d
   ```

3. **Running the container manually:**
   - **Development (Mock Signing):**
     ```bash
     docker run -p 8080:8080 \
       -e ENGINE_SOCKET=0.0.0.0:8080 \
       -v $(pwd)/src/testdata:/app/schemas \
       facturae-engine:latest
     ```
   - **Production (Real Signing + AEAT Submission):**
     Requires mounting a PKCS#12 certificate (`.p12`) and setting environment variables.
     ```bash
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
     ```

### Configuration (Flags & Environment Variables)

The engine can be configured using CLI flags or environment variables. Flags take precedence.

| Flag | Environment Variable | Required | Description |
|---|---|---|---|
| `-socket` | `ENGINE_SOCKET` | Yes | Listen address (e.g. `0.0.0.0:8080`) |
| `-schemas` | `ENGINE_SCHEMAS` | Yes | Directory for cached XSDs |
| `-p12` | `CERT_P12_PATH` | No | Path to PKCS#12 certificate |
| `-p12pass` | `CERT_P12_PASS` | No | Password for PKCS#12 |
| `-aeat` | `AEAT_ENV` | No | AEAT Environment (`test`|`prod`) |
| `-db` | `DB_DRIVER` | No | Backend (`memory`|`postgres`|`sqlite`) |
| `-dsn` | `DB_DSN` | No | Database connection string |

---

## Development Workflow

1. **Clone the repository:**
   ```bash
   git clone https://github.com/ThePromidius/facturae-engine.git
   cd facturae-engine
   ```

2. **Set up Go environment:** Ensure Go 1.25+ is installed.

3. **Download dependencies:**
   ```bash
   go mod download
   ```

4. **Build the application:**
   ```bash
   go build -o facturae-engine ./src/cmd/facturae-engine
   ```

5. **Run in development mode (mock signing):**
   ```bash
   ./facturae-engine -socket 127.0.0.1:8080 -schemas ./src/testdata
   ```

6. **Testing:**
   ```bash
   go test ./... -v
   ```
   
   Tests cover module logic, API endpoints, and integration scenarios. Use `testdata/` for sample JSON payloads.

---

## License

This project is licensed under the Business Source License 1.1 - see the `LICENSE` file in the root for details.
