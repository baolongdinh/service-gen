# Docker Deployment — my-service

## Building the Production Image

```bash
docker build -t my-org/my-service:latest .
```

The build uses a two-stage Dockerfile:
- **Stage 1 (builder):** Compiles a static binary with `CGO_ENABLED=0` on `golang:1.22-alpine`.
- **Stage 2 (runtime):** Copies only the binary into a minimal `alpine:3.19` image, running as a non-root user.

## Tagging and Pushing to a Registry

```bash
# Tag for a registry (e.g., Docker Hub, ECR, GCR, GHCR)
docker tag my-org/my-service:latest ghcr.io/my-org/my-service:v1.0.0
docker tag my-org/my-service:latest ghcr.io/my-org/my-service:latest

# Push both tags
docker push ghcr.io/my-org/my-service:v1.0.0
docker push ghcr.io/my-org/my-service:latest
```

Always push a version tag in addition to `latest` so deployments are reproducible and rollbacks are trivial.

## Required Environment Variables for Production

| Variable | Example | Description |
|---|---|---|
| `APP_PORT` | `8080` | Port the HTTP server listens on |
| `APP_ENV` | `production` | Runtime environment (affects logging) |
| `APP_NAME` | `my-service` | Service identifier in logs |
| `DB_HOST` | `db.prod.internal` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_NAME` | `my-service-db` | PostgreSQL database name |
| `DB_USER` | `my_service` | PostgreSQL role |
| `DB_PASSWORD` | *(secret)* | PostgreSQL password — inject via secret manager |
| `DB_SSL_MODE` | `require` | Use `require` or `verify-full` in production |
| `DB_MAX_CONNS` | `20` | pgxpool max connections |
| `DB_MIN_CONNS` | `5` | pgxpool min idle connections |
| `REDIS_HOST` | `redis.prod.internal` | Redis host |
| `REDIS_PORT` | `6379` | Redis port |
| `REDIS_PASSWORD` | *(secret)* | Redis AUTH password — inject via secret manager |
| `REDIS_DB` | `0` | Redis logical database index |
| `LOG_LEVEL` | `info` | One of: debug, info, warn, error |
| `LOG_FORMAT` | `json` | Use `json` in production for structured log ingestion |

> **Security note:** Never store secrets in `.env` files committed to source control.
> Use a secrets manager (AWS Secrets Manager, GCP Secret Manager, HashiCorp Vault, Kubernetes secrets) and inject them as environment variables at runtime.

## Health Check

The service exposes a health endpoint used by Docker, Kubernetes readiness probes, and load balancers:

```
GET http://localhost:8080/health
```

A healthy response (`HTTP 200`):
```json
{"status": "ok", "db": "ok", "cache": "ok"}
```

A degraded response (`HTTP 503`) is returned when either PostgreSQL or Redis is unreachable.
The Docker `HEALTHCHECK` in the `Dockerfile` polls this endpoint every 30 seconds.

## Running with docker-compose (local / staging)

```bash
cp .env.example .env
# Edit .env with real credentials
make up        # builds image and starts all services
make logs      # tail the app container logs
make down      # stop all services (preserves volumes)
make down-v    # stop and delete volumes (destroys data)
```
