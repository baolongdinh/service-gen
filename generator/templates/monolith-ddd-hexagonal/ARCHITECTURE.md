# Architecture: my-service (DDD + Hexagonal)

## Structure

```
internal/
├── domain/
│   └── my_service/          ← Bounded Context 1 — add new contexts as siblings
│       ├── entity.go        ← Aggregate root (MyService), value object (ID)
│       ├── repository.go    ← Outbound port: Repository interface
│       ├── events.go        ← Domain events (zero infra imports)
│       └── errors.go        ← Sentinel errors (ErrNotFound, ErrDuplicate …)
│
├── application/
│   └── my_service/          ← Bounded Context 1 use cases
│       ├── usecase.go       ← Inbound port: UseCase interface + DTOs
│       ├── service.go       ← Use case implementation (depends on domain.Repository)
│       └── service_test.go  ← Unit tests with mocked Repository
│
├── infrastructure/
│   └── postgres/
│       ├── db.go                      ← pgxpool connection wrapper + Pinger
│       └── my_service_repository.go   ← Outbound adapter: implements domain.Repository
│                                        Add new contexts as sibling files here.
│
└── interfaces/
    └── http/
        ├── router.go          ← Route registration; Config struct holds all UseCases
        ├── handlers/
        │   ├── health.go      ← GET /health — probes DB and cache
        │   └── my_service.go  ← HTTP adapter: calls app.UseCase, never infra directly
        └── middleware/
            └── middleware.go  ← RequestLogger, Auth, Recovery
```

## Hexagonal Ports & Adapters

```
 ┌────────────────────────────────────────────────────────────────┐
 │                      Bounded Context                            │
 │                                                                  │
 │  [HTTP Handler] ──► UseCase           (inbound port)            │
 │  [gRPC Server]  ──►    │                                        │
 │                        ▼                                        │
 │                   service.go  (application layer)               │
 │                        │                                        │
 │                        ▼                                        │
 │              Repository (outbound port) ◄── postgres adapter    │
 │                                         ◄── mock in tests       │
 └────────────────────────────────────────────────────────────────┘
```

- **Inbound port** (`application/my_service/usecase.go`): `UseCase` interface that HTTP/gRPC adapters call — the application layer's public API.
- **Outbound port** (`domain/my_service/repository.go`): `Repository` interface that the application layer calls; the postgres adapter implements it.
- **Domain** has zero imports from `application/` or `infrastructure/`. Verify: `grep -r "infrastructure\|application" internal/domain/` must return nothing.

## Layer Dependency Rules

```
interfaces/http  →  application  →  domain
infrastructure               →  domain
```

Arrows show allowed import direction. Violations are architecture bugs.

## Adding a New Bounded Context

1. Create `internal/domain/order/` — `entity.go`, `repository.go`, `errors.go`, `events.go`
2. Create `internal/application/order/` — `usecase.go`, `service.go`, `service_test.go`
3. Create `internal/infrastructure/postgres/order_repository.go` implementing `order.Repository`
4. Create `internal/interfaces/http/handlers/order.go` calling `order.UseCase`
5. Add `OrderUseCase order.UseCase` to `httpserver.Config` and register routes in `router.go`
6. Wire in `cmd/server/main.go`: `postgres.NewOrderRepository(db)` → `order.New(repo, log)` → config

Bounded contexts communicate via **domain events** dispatched through an event bus (wire one in `service.go` when needed) — never by importing each other's domain packages directly.

## For AI Assistants

Step-by-step guide to adding a feature:

1. **Domain** — modify or add to `internal/domain/my_service/entity.go`
   - Add a method that validates business rules and records a domain event
   - `go test ./internal/domain/...`

2. **Port** — if the use case interface changes, update `internal/application/my_service/usecase.go`

3. **Application service** — implement the method in `service.go`
   - FindByID → call domain method → Save → PopEvents
   - `go test ./internal/application/...`

4. **Infrastructure** — add queries to `internal/infrastructure/postgres/my_service_repository.go` if needed

5. **HTTP handler** — add/modify in `internal/interfaces/http/handlers/my_service.go`
   - Decode → call UseCase → map domain error → encode response

6. **Route** — register in `internal/interfaces/http/router.go`

7. **Layer check** — `grep -r "infrastructure" internal/domain/` must return nothing
