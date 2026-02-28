# golang-grpc-enterprise-demo

> **High-performance Golang gRPC + REST Hybrid API with PostgreSQL, JWT, Docker & OpenTelemetry — Ready for 2026 Cloud-Native Production.**

---

## Features

- **gRPC Service** — `UserService` with `GetUser`, `CreateUser`, `ListUsers` (pagination), defined in Protobuf v3
- **REST Hybrid API** — Gin-based HTTP gateway exposing the same operations as RESTful endpoints
- **PostgreSQL + GORM v2** — Auto-migration, type-safe queries, seed data
- **JWT Authentication** — Protects both gRPC (unary interceptor) and REST (Gin middleware) endpoints
- **Rate Limiting** — Per-IP token bucket (default 100 req/min)
- **OpenTelemetry Tracing** — Console exporter for every request span
- **Structured Logging** — `uber-go/zap` production logger
- **Clean Architecture** — handler → service → repository → model
- **Docker Ready** — Multi-stage Dockerfile + docker-compose one-command deployment

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Clients                              │
│              (curl / gRPCurl / Frontend)                    │
└──────────┬──────────────────────┬───────────────────────────┘
           │ HTTP :8080           │ gRPC :50051
           ▼                     ▼
┌──────────────────┐   ┌──────────────────┐
│   Gin Router     │   │   gRPC Server    │
│  Rate Limiter    │   │  OTel Handler    │
│  JWT Middleware   │   │  JWT Interceptor │
└────────┬─────────┘   └────────┬─────────┘
         │                      │
         ▼                      ▼
┌──────────────────────────────────────────┐
│            Handler Layer                 │
│   rest_handler.go  │  grpc_handler.go    │
└────────────────────┬─────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────┐
│            Service Layer                 │
│         user_service.go                  │
│   (business logic, password hashing)     │
└────────────────────┬─────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────┐
│          Repository Layer                │
│       user_repository.go                 │
│         (GORM v2 queries)                │
└────────────────────┬─────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────┐
│          PostgreSQL Database             │
│   users (id, name, email, password_hash, │
│          created_at)                     │
└──────────────────────────────────────────┘
```

### Directory Structure

```
golang-grpc-enterprise-demo/
├── cmd/server/main.go            # Entry point — wires everything
├── internal/
│   ├── config/config.go          # Viper-based env config
│   ├── model/user.go             # GORM User model
│   ├── repository/               # Data access (interface + GORM impl)
│   ├── service/                  # Business logic + unit tests
│   ├── handler/                  # gRPC handler + REST handler
│   ├── middleware/               # JWT auth + rate limiter
│   └── telemetry/tracing.go      # OpenTelemetry setup
├── proto/
│   ├── user.proto                # Protobuf v3 definition
│   └── gen/user/v1/              # Generated Go code
├── pkg/seed/seed.go              # DB seed (3 test users)
├── Dockerfile                    # Multi-stage build
├── docker-compose.yml            # app + postgres
├── Makefile
├── .env.example
└── README.md
```

---

## Quick Start

### Option 1: Docker Compose (Recommended)

```bash
# Clone and start
cd golang-grpc-enterprise-demo
docker compose up -d --build

# Wait a few seconds, then check health
curl http://localhost:8080/healthz
# → {"status":"ok"}
```

### Option 2: Local Development

**Prerequisites:** Go 1.23+, PostgreSQL running locally

```bash
# 1. Copy env file and edit as needed
cp .env.example .env

# 2. Install dependencies
go mod tidy

# 3. Run
make run
# or: go run ./cmd/server
```

---

## API Testing (curl)

### 1. Health Check (no auth required)

```bash
curl http://localhost:8080/healthz
```

### 2. Login (get JWT token)

Seed users all have password `password123`.

```bash
curl -s -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"password123"}'
```

Response:
```json
{"token":"eyJhbGciOiJIUzI1NiIs..."}
```

Save the token:
```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"password123"}' | jq -r .token)
```

### 3. Get User by ID

```bash
curl -s http://localhost:8080/api/v1/users/1 \
  -H "Authorization: Bearer $TOKEN"
```

### 4. Create User

```bash
curl -s -X POST http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Dave Chen","email":"dave@example.com","password":"secure456"}'
```

### 5. List Users (with pagination)

```bash
curl -s "http://localhost:8080/api/v1/users?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN"
```

### 6. gRPC (with grpcurl)

```bash
# List users via gRPC
grpcurl -plaintext \
  -H "authorization: Bearer $TOKEN" \
  -d '{"page":1,"page_size":10}' \
  localhost:50051 user.v1.UserService/ListUsers
```

---

## Running Tests

```bash
make test
# or
go test -v -race ./...
```

---

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `APP_ENV` | `development` | `development` / `production` |
| `GRPC_PORT` | `50051` | gRPC listen port |
| `HTTP_PORT` | `8080` | HTTP/REST listen port |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | DB username |
| `DB_PASSWORD` | `postgres` | DB password |
| `DB_NAME` | `grpc_demo` | Database name |
| `DB_SSLMODE` | `disable` | PostgreSQL SSL mode |
| `JWT_SECRET` | — | JWT signing key (**must set in production**) |
| `JWT_EXPIRY_HOURS` | `24` | Token validity in hours |
| `RATE_LIMIT_RPM` | `100` | Max requests per minute per IP |

---

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| RPC Framework | gRPC + Protobuf v3 |
| HTTP Framework | Gin |
| ORM | GORM v2 |
| Database | PostgreSQL 16 |
| Auth | JWT (golang-jwt/v5) |
| Rate Limiting | golang.org/x/time/rate |
| Telemetry | OpenTelemetry (console exporter) |
| Logging | uber-go/zap |
| Config | Viper |
| Container | Docker multi-stage + Compose |

---

## Performance Notes

- gRPC binary protocol provides **~10x** lower latency vs JSON REST for internal service calls
- Connection pooling via GORM + pgx driver
- Rate limiter uses in-memory token bucket with automatic cleanup — suitable for single-instance or behind a load balancer with sticky sessions
- OpenTelemetry traces are output to console; swap to Jaeger/OTLP exporter for production dashboards

---

## License

MIT
