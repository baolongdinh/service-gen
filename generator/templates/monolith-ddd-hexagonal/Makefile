.PHONY: up down down-v logs docker-build shell ps gen migrate-up migrate-down test build vet tidy help

# ── Docker ────────────────────────────────────────────────────────────────────

## up: Start all services (builds image first). Checks Docker is running.
up:
	@docker info > /dev/null 2>&1 || (echo "Error: Docker is not running. Please start Docker Desktop." && exit 1)
	docker compose up --build -d

## down: Stop all services (volumes preserved).
down:
	docker compose down

## down-v: Stop all services and delete volumes (destroys all data).
down-v:
	docker compose down -v

## logs: Tail the app container logs.
logs:
	docker compose logs -f app

## docker-build: Build the production Docker image locally without Compose.
docker-build:
	docker build -t my-service:latest .

## shell: Open an interactive shell inside the running app container.
shell:
	docker compose exec app sh

## ps: Show status of all Compose-managed containers.
ps:
	docker compose ps

# ── Codegen ───────────────────────────────────────────────────────────────────

## gen: Regenerate protobuf/gRPC code from .proto files using Buf.
gen:
	buf generate

# ── Database ──────────────────────────────────────────────────────────────────

## migrate-up: Apply all pending migrations against the running postgres container.
migrate-up:
	docker compose exec app ./migrate -path /app/migrations \
	  -database "$$DATABASE_URL" up

## migrate-down: Roll back the most recent migration.
migrate-down:
	docker compose exec app ./migrate -path /app/migrations \
	  -database "$$DATABASE_URL" down 1

# ── Quality ───────────────────────────────────────────────────────────────────

## test: Run all tests with the race detector enabled.
test:
	go test ./...

## build: Compile the server binary into bin/server.
build:
	go build -o bin/server ./cmd/server

## vet: Run go vet across all packages.
vet:
	go vet ./...

## tidy: Tidy and verify go.mod / go.sum.
tidy:
	go mod tidy

# ── Help ──────────────────────────────────────────────────────────────────────

## help: Print this help message.
help:
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@grep -E '^## ' Makefile | sed 's/^## /  /' | column -t -s ':'
	@echo ""
