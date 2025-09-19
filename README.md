# savannah-technical-interview

# Enterprise Customer Module (Go)


This scaffold provides a production-minded customer module with:


- Clear package structure (internal/pkg)
- Repository using `sqlx` and Postgres
- Service layer with optimistic locking
- HTTP handlers using `chi`
- Validation via `go-playground/validator`
- Structured logging via `zap`


## Next steps
- Add migrations (e.g. using golang-migrate) to create `customers` table.
- Add authentication/authorization middleware and RBAC.
- Add unit and integration tests.
- Add observability: metrics (Prometheus), tracing (OpenTelemetry).