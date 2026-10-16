# EngFlex API — Go Control Plane

Go + Gin + GORM backend. Owns data, auth, and business logic, and provisions AI
voice sessions (persona/context/token) without ever touching realtime audio. See
the [root README](../../README.md) for the full system picture.

## Layout

```text
cmd/server/main.go          # Entrypoint: wires config + router only
config/                     # Auth, cors, database, log, server, voice
internal/
  common/                   # AppError, ApiResponse envelope, enums, limits
  database/
    migrations/             # Goose SQL — schema source of truth
    models/                 # GORM structs
    repositories/           # Repos over *gorm.DB + mocks/
  modules/<domain>/         # attempts, conversations, lessons, profiles, scenarios, videos, vocabulary
  server/middleware/        # RequireAuth, RequestID, RequestLogger, Recovery
  utils/                    # Map/MapSlice, pagination, IDs
docs/                       # Generated Swagger (served at /swagger)
```

All routes mount under `/api/v1` (`internal/server/router.go`).

## Quickstart

```bash
make migration-up        # goose up (DATABASE_URL from .env)
go run ./cmd/server      # http://localhost:8000, Swagger at /swagger
```

## Commands

| Task | Command |
|------|---------|
| All tests | `go test ./... -count=1` |
| One suite | `go test ./internal/modules/<domain>/... -count=1 -v` |
| Vet / fmt | `go vet ./...`, `gofmt -l internal/` (must be empty) |
| Mocks | `make mocks` |
| Contracts | `make contracts` |
| New migration | `make migration-create name=...` |
| Apply / rollback | `make migration-up` / `make migration-down` |
| Swagger | `make swagger` |

## Error handling

Centralized in `internal/common/errors.go` + `response.go`. Controllers return
errors via `common.Fail(c, err)`; the client always receives the `{ "message":
... }` envelope, and any error that is not an `*AppError` becomes a generic 500.

| Helper | Status |
|--------|--------|
| `common.BadRequest` | 400 |
| `common.Unauthorized` | 401 |
| `common.Forbidden` | 403 |
| `common.NotFound` | 404 |
| `common.Conflict` | 409 |
| `common.UnprocessableEntity` | 422 |
| `common.TooManyRequests` | 429 |
| `common.Internal` | 500 |
| `common.ServiceUnavailable` | 503 |

`common.FromDBError(err, resource)` maps database errors to `*AppError`:

| DB error | Maps to | Status |
|----------|---------|--------|
| `*InvalidIDError` (unparseable id) | `BadRequest` | 400 |
| `gorm.ErrRecordNotFound` | `NotFound` | 404 |
| Postgres `23505` (unique violation) | `Conflict` | 409 |
| Postgres `23503` (FK violation) | `NotFound` | 404 |
| anything else | `Internal` | 500 |
