# Architecture

## Dependency Graph

```
cmd/server/main.go
       │
       ▼
internal/handler/          (HTTP layer: decode request, encode response)
       │
       ▼
internal/services/         (Business logic, validation, orchestration)
       │
       ▼
internal/repositories/     (Data access interfaces)
       │
       ▼
internal/repositories/postgres/   (PostgreSQL implementation)
       │
       ▼
  PostgreSQL / Redis

─────────────────────────────────────────────────────────
pkg/                       (Horizontal utilities — imported by any layer)
  ├── logger/
  └── httputil/

internal/cache/            (Cache interface + memory/redis implementations)
internal/config/           (Environment-driven configuration)
internal/errors/           (Typed service errors with error codes)
internal/models/           (Plain data structs — no behaviour, no DB tags)
internal/middleware/       (HTTP middleware: auth, logging, recovery)
internal/must/             (Fatal-on-error helpers for startup only)
```

## Directory Responsibilities

| Directory | Responsibility | Allowed imports |
|-----------|---------------|-----------------|
| `cmd/server/` | Process entry point. Wire all dependencies, start HTTP server, handle OS signals. | Everything — this is the composition root. |
| `internal/handler/` | Decode HTTP requests, call service methods, encode HTTP responses. Map service errors to status codes. | `internal/services`, `internal/errors`, `internal/repositories` (filter types only), `internal/middleware`, `pkg/*` |
| `internal/services/` | Business rules, validation, orchestration across repositories. | `internal/repositories` (interfaces only), `internal/models`, `internal/errors`, `internal/cache`, `pkg/*` |
| `internal/repositories/` | Data access interfaces and filter/pagination types. No SQL here. | `internal/models` |
| `internal/repositories/postgres/` | PostgreSQL implementations of repository interfaces. SQL lives here. | `internal/repositories`, `internal/models`, `pkg/logger` |
| `internal/models/` | Plain Go structs that represent domain entities. No methods, no DB tags, no JSON logic. | Standard library only |
| `internal/errors/` | Typed error constructors (`NotFound`, `AlreadyExists`, `Internal`, etc.) and `ServiceError` struct. | Standard library only |
| `internal/cache/` | Cache interface + in-memory and Redis implementations. | `internal/config`, `pkg/*` |
| `internal/config/` | Load and validate configuration from environment variables. | Standard library, `github.com/caarlos0/env` |
| `internal/middleware/` | HTTP middleware: request logging, panic recovery, authentication stub. | `pkg/logger`, `github.com/go-chi/chi/v5/middleware` |
| `internal/must/` | `must.NoError(err)` — panic on error for startup wiring only. Never call inside request handlers. | Standard library only |
| `pkg/logger/` | Structured logger interface backed by `slog`. | Standard library |
| `pkg/httputil/` | Thin helpers: `JSON`, `Error`, `BadRequest`, `InternalServerError`. | Standard library |

## The Golden Rule

> **Handlers call services. Services call repositories. Repositories call the database. Never skip layers.**

- A handler must not query the database directly.
- A service must not import `internal/repositories/postgres` (the concrete implementation).
- A repository must not contain business logic or HTTP concepts.

## Layer Isolation — Verifiable

No file in `internal/services/` or `internal/handler/` may import the postgres package:

```sh
grep -r 'repositories/postgres' internal/services/ internal/handler/
# must return no output
```

Run this as a CI lint step to enforce the constraint automatically.

## For AI Assistants

Follow these steps in order when adding a new resource. Each step maps to exactly one layer — do not combine steps or skip ahead.

1. **Add the model** to `internal/models/`
   - Plain struct with exported fields.
   - No methods, no `db:` tags, no JSON logic.
   - Example: `type Widget struct { ID string; Name string; CreatedAt time.Time }`.

2. **Add the repository interface** to `internal/repositories/repositories.go`
   - Define a `WidgetRepository` interface with CRUD methods.
   - Add a `WidgetFilter` struct (embed `Pagination`) for list queries.
   - Do not write SQL here.

3. **Add the postgres implementation** to `internal/repositories/postgres/widget_repository.go`
   - Implement `WidgetRepository` using `pgx`.
   - Map `pgx.ErrNoRows` to `repositories.ErrRecordNotFound`.
   - Map unique constraint violations to `repositories.ErrDuplicateKey`.

4. **Add the service** to `internal/services/widget_service.go`
   - Define `WidgetService` interface and `widgetSvc` struct.
   - Inject `WidgetRepository` (the interface, not `*postgres.WidgetRepository`).
   - Use `mapRepoError` to translate repo errors to service errors.

5. **Add the handler** to `internal/handler/widget_handler.go`
   - Define `widgetHandler` struct, inject `WidgetService` and `logger.Logger`.
   - Call `RegisterWidgetRoutes(r chi.Router, ...)` from `router.go`.
   - Use `h.handleError(w, err)` consistently — never write raw `http.Error` calls for service errors.

6. **Wire the new repository** in `cmd/server/main.go`
   - Instantiate `postgres.NewWidgetRepository(db)`.
   - Pass it to `services.NewWidgetService(...)`.
   - Pass the service to `handler.New(...)` or `handler.RegisterWidgetRoutes(...)`.
