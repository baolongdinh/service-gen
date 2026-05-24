# ─── Stage 1: Build ───────────────────────────────────────────────────────────
# go.mod and go.sum are copied before source so that `go mod download` is cached
# as its own layer. Subsequent builds only re-download dependencies when the
# module files change, not on every source edit.
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy dependency manifests first to exploit Docker layer caching.
COPY go.mod go.sum ./
RUN go mod download

# Now copy the rest of the source and compile.
# CGO_ENABLED=0 produces a fully static binary with no glibc dependency,
# which is required for the minimal alpine runtime image below.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server ./cmd/server

# ─── Stage 2: Runtime ─────────────────────────────────────────────────────────
# Start from a minimal base — no compiler, no shell tools shipped to production.
FROM alpine:3.19

# ca-certificates: required for TLS connections to external services.
# tzdata: required if the app formats timestamps with user-facing timezones.
RUN apk --no-cache add ca-certificates tzdata

# Non-root user: limits blast radius if the container process is ever exploited.
# Running as root inside a container grants unnecessary host-level capabilities.
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/server ./server

USER appuser

EXPOSE 8080

# Docker HEALTHCHECK — used by `docker compose` health conditions and
# container orchestrators (ECS, Swarm). Kubernetes uses its own probes.
HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["./server"]
