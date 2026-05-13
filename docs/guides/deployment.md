# Deployment Guide

## Docker

### Building the Image

```bash
docker build -t facturae-engine:latest .
```

The included `Dockerfile` uses a multi-stage build:

1. **Builder stage:** `golang:1.22-alpine` — compiles a static binary with `CGO_ENABLED=0`
2. **Run stage:** `alpine:latest` with installed system dependencies:
   - `openssl` — for PKCS#12 certificate extraction
   - `libxml2-utils` — for XSD validation via `xmllint`
   - `tzdata` — timezone support

### Running with Docker

The engine supports configuration via environment variables, making it easy to run with Docker.

```bash
# Development mode (mock signing)
docker run -p 8080:8080 \
  -e ENGINE_SOCKET=0.0.0.0:8080 \
  -e ENGINE_SCHEMAS=/app/schemas \
  -v facturae-schemas:/app/schemas \
  facturae-engine:latest

# With PKCS#12 certificate (mount cert file and use Env Vars)
docker run -p 8080:8080 \
  -v /host/path/cert.p12:/app/cert.p12:ro \
  -v facturae-schemas:/app/schemas \
  -e CERT_P12_PATH=/app/cert.p12 \
  -e CERT_P12_PASS="your-password" \
  -e AEAT_ENV=test \
  facturae-engine:latest
```

### Docker Compose

Using Docker Compose with an `.env` file is the recommended approach:

1. Copy `.env.example` to `.env`.
2. Configure your variables.
3. Run `docker-compose up -d`.

The included `docker-compose.yml` is pre-configured to use these environment variables.

---

## Configuration (Flags & Environment Variables)

The engine can be configured using CLI flags or environment variables. Flags take precedence.

| Flag | Environment Variable | Default | Description |
|------|----------------------|---------|-------------|
| `-socket` | `ENGINE_SOCKET` | `127.0.0.1:8080` | Listen address |
| `-schemas` | `ENGINE_SCHEMAS` | `./schemas` | XSD cache directory |
| `-p12` | `CERT_P12_PATH` | | PKCS#12 file path |
| `-p12pass` | `CERT_P12_PASS` | | PKCS#12 password |
| `-key` | `CERT_KEY_PATH` | | PEM private key path |
| `-cert` | `CERT_PEM_PATH` | | PEM certificate path |
| `-aeat` | `AEAT_ENV` | | AEAT environment (`test`/`prod`) |
| `-db` | `DB_DRIVER` | `memory` | Chain store driver (`memory`/`postgres`/`sqlite`) |
| `-dsn` | `DB_DSN` | | Database connection string |

---

## Production Considerations

### Security

1. **Certificate management:**
   - Use `-p12` + `-p12pass` for certificate loading at startup
   - The certificate must come from a Spanish CA authorized by AEAT (e.g., FNMT, ACCV, Izenpe)
   - Store passwords in a secrets manager or use Docker secrets (mount `/run/secrets/`)
   - For production, avoid embedding passwords in docker-compose files

2. **Network:**
   - Run as a sidecar on `127.0.0.1` (do not expose the port externally)
   - Use a reverse proxy (nginx, Caddy) if external access is needed
   - For Unix sockets, set appropriate permissions (`umask 0077`)
   - The AEAT outbound calls require HTTPS connectivity to `*.aeat.es`

3. **TLS/mTLS:**
   - AEAT production requires mTLS with an authorized certificate
   - The `LoadTLSCertFromP12` function extracts the TLS certificate from the PKCS#12 file
   - Without TLS cert, AEAT test environment may work, but production will fail

### Persistence

1. **Chain store:** Always use a persistent backend in production:
   - **SQLite** for single-instance deployments (embedded, zero-config)
   - **PostgreSQL** for multi-instance or HA deployments (shared chain)
   - The `-db memory` mode is for development/testing only — all chain data is lost on restart

2. **XSD cache:** The `-schemas` directory should be backed by a persistent volume to avoid re-downloading XSD files on every restart

### Performance

- The engine is designed for low-latency sidecar operation
- AEAT submission runs asynchronously (goroutine) — the API returns before the submission completes
- Chain operations are O(1) for append and O(n) for `All()` — monitor chain length for memory store
- QR generation uses a simplified encoder — suitable for short URLs (Verifactu URLs are typically <200 chars)

### Monitoring

- **Health check:** `GET /health` returns signer algorithm and chain length
- **Logs:** The engine logs in Spanish (API errors) and English (system info) to stdout
- **Signal handling:** Graceful shutdown on `SIGINT`/`SIGTERM` (Unix socket cleanup)

### Failure Modes

| Scenario | Behavior |
|----------|----------|
| `xmllint` not available | XSD validation is skipped with a warning; structural validation still applies |
| AEAT submission fails | Error is logged; the API response still returns the signed XML. Retry logic handles transient failures |
| AEAT rejects invoice | Logged with rejection reason. The signed XML was already returned to the caller |
| Chain store unavailable | API returns `500 Internal Server Error` |
| PKCS#12 decryption fails | Engine exits with a fatal error on startup |
| OpenSSL not installed | Engine exits with a fatal error if `-p12` is used; use `-key`/`-cert` instead |

### Logging

All output goes to stdout. In production, redirect to a log collector:

```bash
./facturae-engine -socket /var/run/facturae.sock ... 2>&1 | logger -t facturae-engine
```

### Environment-Specific Configurations

**Test environment** (`-aeat test`):
```
Endpoint: https://prewww1.aeat.es/wlpl/VERIFACTU-CONT/ws/VeriFactuSOAP
- May work without mTLS certificate (depends on AEAT test policy)
- No legal effect
```

**Production environment** (`-aeat prod`):
```
Endpoint: https://www1.aeat.es/wlpl/VERIFACTU-CONT/ws/VeriFactuSOAP
- Requires valid AEAT-authorized certificate with mTLS
- Legal compliance: Verifactu law (Ley 18/2022, Real Decreto 1007/2023)
```
