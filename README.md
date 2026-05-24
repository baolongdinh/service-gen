# service-gen

Go service scaffolding tool — generates production-ready project structures for two architectural patterns:

- **Microservice** (Clean Architecture) — layered packages, interface-driven repositories, single deployable unit
- **Monolith DDD + Hexagonal** — Domain-Driven Design with Ports & Adapters, bounded context structure, domain isolation

---

## Cài đặt

### Build từ source

```bash
git clone https://github.com/interlabs/service-gen
cd service-gen
go build -o service-gen ./cmd
```

Hoặc install trực tiếp vào `$GOPATH/bin`:

```bash
go install github.com/interlabs/service-gen/cmd@latest
```

---

## Sử dụng

```bash
service-gen init
```

Tool sẽ hỏi các thông tin sau theo thứ tự:

```
Select template type:
  1) microservice    — Clean Architecture
  2) monolith-ddd   — DDD + Hexagonal (Ports & Adapters)
Choice [1]: 2

Go module path (e.g. github.com/acme/user-service): github.com/acme/order-service
Service name (kebab-case) [order-service]:
Organization / GitHub org name [acme]:
Buf.build org name [acme]:
Go version [1.22]:
Output directory [./order-service]:
```

Sau khi xác nhận, tool sẽ sinh toàn bộ project vào thư mục đã chọn.

---

## Các template

### 1. Microservice — Clean Architecture

Phù hợp khi service chạy độc lập, deploy riêng, chỉ có một domain chính.

**Cấu trúc sinh ra:**

```
my-service/
├── cmd/server/main.go          ← entry point, dependency wiring
├── internal/
│   ├── config/                 ← env-based config (caarlos0/env)
│   ├── models/                 ← plain Go structs
│   ├── repositories/           ← Repository interfaces (contracts)
│   │   └── postgres/           ← PostgreSQL implementations
│   ├── services/               ← business logic, implements repo interfaces
│   ├── handler/                ← HTTP handlers (chi router)
│   ├── middleware/             ← logging, auth, recovery
│   ├── cache/                  ← Cache interface + Redis/memory impls
│   ├── errors/                 ← typed service errors with HTTP mapping
│   └── must/                   ← startup-only panic helpers
├── pkg/
│   ├── logger/                 ← slog wrapper (injectable Logger interface)
│   ├── httputil/               ← JSON response helpers
│   └── validator/              ← go-playground/validator wrapper
├── migrations/                 ← golang-migrate SQL files
├── proto/                      ← Protobuf definitions (buf toolchain)
├── Dockerfile                  ← multi-stage build (alpine)
├── docker-compose.yml          ← postgres + redis + app
├── .env.example
└── Makefile
```

**Dependency direction:**

```
handler → services → repositories (interface) → postgres (impl)
```

---

### 2. Monolith DDD + Hexagonal

Phù hợp cho hệ thống lớn hơn với nhiều bounded context, cần tách biệt domain logic khỏi infrastructure.

**Cấu trúc sinh ra:**

```
my-service/
├── cmd/server/main.go          ← entry point, hexagonal wiring
├── config/                     ← env-based config
├── internal/
│   ├── domain/
│   │   └── my_service/         ← Bounded Context (thêm context mới ở đây)
│   │       ├── entity.go       ← Aggregate root + ID value object
│   │       ├── repository.go   ← Outbound port (interface)
│   │       ├── events.go       ← Domain events (zero infra imports)
│   │       └── errors.go       ← Sentinel errors
│   ├── application/
│   │   └── my_service/         ← Use cases cho bounded context
│   │       ├── usecase.go      ← Inbound port (interface) + DTOs
│   │       ├── service.go      ← Implementation
│   │       └── service_test.go ← Unit tests với mocked repository
│   ├── infrastructure/
│   │   └── postgres/
│   │       ├── db.go           ← pgxpool connection
│   │       └── my_service_repository.go ← Outbound adapter
│   └── interfaces/
│       └── http/
│           ├── router.go       ← Route registration
│           ├── handlers/       ← HTTP adapters (gọi UseCase, không gọi infra)
│           └── middleware/     ← logging, auth, recovery
├── migrations/
├── proto/
├── Dockerfile
├── docker-compose.yml
├── .env.example
└── Makefile
```

**Hexagonal ports:**

```
[HTTP Handler] → UseCase (inbound port)
                     ↓
              service.go (application)
                     ↓
          Repository (outbound port) ← postgres adapter
                                     ← mock in tests
```

**Domain isolation:** `internal/domain/` không import bất kỳ thứ gì từ `application/` hay `infrastructure/`. Kiểm tra bằng:

```bash
grep -r "infrastructure\|application" internal/domain/
# Phải không ra kết quả nào (chỉ có comment)
```

---

## Thêm Bounded Context mới (Monolith)

Ví dụ thêm context `order` vào project:

```bash
# 1. Domain
mkdir -p internal/domain/order
# Tạo: entity.go, repository.go, errors.go, events.go

# 2. Application
mkdir -p internal/application/order
# Tạo: usecase.go, service.go, service_test.go

# 3. Infrastructure adapter
# Thêm file: internal/infrastructure/postgres/order_repository.go

# 4. HTTP handler
# Thêm file: internal/interfaces/http/handlers/order.go

# 5. Wire vào router
# Thêm OrderUseCase vào httpserver.Config trong router.go

# 6. Wire vào main.go
# postgres.NewOrderRepository(db) → order.New(repo, log) → config
```

Các context giao tiếp qua **domain events** — không import package domain của nhau trực tiếp.

---

## Sau khi generate

```bash
cd <output-dir>

# Copy env config
cp .env.example .env

# Start dependencies (postgres + redis)
make up

# Download dependencies & build
go mod tidy
go build ./...

# Chạy tests
go test ./...

# Generate protobuf code (cần buf CLI)
make gen

# Build Docker image
make docker-build
```

---

## Makefile targets

| Target | Mô tả |
|--------|-------|
| `make up` | Start docker-compose (postgres + redis) |
| `make down` | Stop docker-compose |
| `make run` | Chạy server locally |
| `make test` | `go test ./...` |
| `make gen` | Generate protobuf code với buf |
| `make migrate-up` | Apply DB migrations |
| `make migrate-down` | Rollback DB migrations |
| `make docker-build` | Build Docker image |
| `make lint` | Run golangci-lint |

---

## Tech stack (generated projects)

| Concern | Library |
|---------|---------|
| HTTP router | `github.com/go-chi/chi/v5` |
| Config | `github.com/caarlos0/env/v11` |
| PostgreSQL | `github.com/jackc/pgx/v5` + pgxpool |
| Redis | `github.com/redis/go-redis/v9` |
| Migrations | `github.com/golang-migrate/migrate/v4` |
| Logger | `log/slog` (stdlib) |
| Validation | `github.com/go-playground/validator/v10` |
| Testing | `github.com/stretchr/testify` |
| Protobuf | `google.golang.org/grpc` + buf toolchain |
| Go version | 1.22+ |
| Base image | `golang:1.22-alpine` → `alpine:3.19` |

---

## Placeholder substitution

Tất cả file trong template dùng các placeholder sau, được thay thế tự động theo thông tin bạn nhập:

| Placeholder | Ví dụ sau generate |
|-------------|-------------------|
| `github.com/my-org/my-service` | `github.com/acme/order-service` |
| `my-org` | `acme` |
| `my-service` | `order-service` |
| `my_service` | `order_service` |
| `MyService` | `OrderService` |

---

## Yêu cầu

- Go 1.22+
- Docker & Docker Compose (để chạy dependencies)
- `buf` CLI (để generate protobuf, optional)
- `golangci-lint` (để lint, optional)
- `golang-migrate` CLI (để chạy migrations thủ công, optional — Makefile có sẵn)
