---

## Development Workflow

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/ThePromidius/facturae-engine-v2.git
    cd facturae-engine-v2
    ```

2.  **Set up Go environment:** Ensure Go 1.25+ is installed.

3.  **Download dependencies:**
    ```bash
    go mod download
    ```

4.  **Build the application:**
    ```bash
    go build -o facturae-engine ./cmd/server/main.go
    ```

5.  **Run in development mode (mock signing):**
    ```bash
    ./facturae-engine -socket 127.0.0.1:8080 -schemas ./schemas
    ```

6.  **Testing:**
    ```bash
    go test ./... -v
    ```
    
    Tests cover module logic, API endpoints, and integration scenarios. Use `testdata/` for sample JSON payloads.

---

## Adding a New Feature

1.  **Define types:** Add or modify structs in `internal/<module>/`.
2.  **Write tests:** Create tests in `internal/<module>/<module>_test.go` following existing patterns (e.g., table-driven tests).
3.  **Wire dependencies:** Update `cmd/server/main.go` to inject new components.
4.  **Expose via API:** Add a handler in `internal/api/handlers.go` and register it in `internal/api/server.go`.
5.  **Document:** Update `README.md` and relevant module documentation in `docs/modules/`.
6.  **Run all tests** before committing.

---

## Code Conventions

- Use idiomatic Go. Prefer standard library where possible.
- API error messages are in Spanish; internal logs are in English.
- Exported types and functions should have godoc comments.
- Table-driven tests are preferred.
- Mock implementations for external dependencies should be used liberally.
- Minimize external dependencies: the project relies only on `database/sql` drivers and `net/http`.

---

## Environment Variables

This service uses CLI flags exclusively, not environment variables, for configuration.
