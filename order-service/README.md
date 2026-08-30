# Order Service

Orchestrates order placement: validates the user (via user-service) and
each line item (via product-service), persists the order, charges it (via
payment-service), and records the final outcome. Sole owner of the
`shopstream_orders` PostgreSQL database.

## Overview

| Field | Value |
|---|---|
| Language | Go |
| Framework | Gin |
| Runtime version | Go 1.22+ (developed/tested on 1.25) |
| Build tool | Go modules (`go build`) |
| Application port | `8082` |
| Database | PostgreSQL 15+ |

## Build

```bash
go mod download
go build -o bin/order-service ./cmd/server
```

## Test

```bash
go test ./... -v
go vet ./...
```

10 tests across two packages:
- `internal/service` — business logic against hand-rolled fakes for the
  repository and all three downstream clients (no network, no DB).
- `internal/handlers` — HTTP-level tests via `httptest`, same fakes wired
  through a real `gin.Engine`.

## Run

```bash
export $(grep -v '^#' .env | xargs)
./bin/order-service
```

## Environment variables

See [.env.example](.env.example).

| Variable | Required | Description |
|---|---|---|
| `SERVER_PORT` | no (8082) | HTTP port |
| `DB_HOST` / `DB_PORT` / `DB_NAME` / `DB_USER` / `DB_PASSWORD` | yes | PostgreSQL connection |
| `USER_SERVICE_URL` | yes | Base URL for user-service |
| `PRODUCT_SERVICE_URL` | yes | Base URL for product-service |
| `PAYMENT_SERVICE_URL` | yes | Base URL for payment-service |
| `HTTP_CLIENT_TIMEOUT_SECONDS` | no (3) | Per-call timeout to downstream services |
| `HTTP_CLIENT_RETRY_ATTEMPTS` | no (2) | Retries on network error / 502/503/504 |

## Dependencies

`gin-gonic/gin`, `jackc/pgx/v5`, `google/uuid`, `stretchr/testify` (test
only). See [go.mod](go.mod).

## Database requirements

PostgreSQL 15+. Schema in [migrations](migrations) (plain numbered
`.up.sql`/`.down.sql` pairs, compatible with `golang-migrate` or any
similar tool your Jenkins pipeline drives):

- `0001_init.up.sql` / `.down.sql` — `orders` and `order_items` tables
- `0002_seed_data.up.sql` / `.down.sql` — one sample confirmed order,
  referencing the seeded user/product IDs from user-service/product-service

`user_id` and `product_id` are **logical** references only — order-service
never joins across service boundaries; it validates them at request time
via HTTP calls to user-service and product-service instead.

### Local development

```bash
docker run --name shopstream-orders-db -e POSTGRES_DB=shopstream_orders \
  -e POSTGRES_USER=shopstream -e POSTGRES_PASSWORD=devpassword \
  -p 5434:5432 -d postgres:15

migrate -database "$DATABASE_URL" -path migrations up
```

## API endpoints

| Method | Path | Description |
|---|---|---|
| POST | `/api/v1/orders` | Place an order (validates user + products, charges payment) |
| GET | `/api/v1/orders/{id}` | Get an order by ID |
| GET | `/api/v1/orders?userId={id}` | List a user's orders |

## Health endpoint

`GET /health` — liveness only, no dependency checks.

## Readiness endpoint

`GET /ready` — pings PostgreSQL with a 1.5s timeout; `503` if unreachable.

## Service-to-service dependencies

- **user-service** — `GET /api/v1/users/{id}` to verify the buyer exists
- **product-service** — `GET /api/v1/products/{id}` to price and validate stock for each line item
- **payment-service** — `POST /api/v1/payments` to charge the order total (sent with an `Idempotency-Key` header equal to the order ID)

All three calls go through `internal/client.ResilientClient`: bounded
timeout + exponential-backoff retries on network errors, timeouts, and
502/503/504.

## Failure scenarios this service is designed to reproduce

- **Database connection failure**: wrong `DB_HOST` → process exits at
  startup (fails the readiness/liveness probe immediately, no silent
  half-broken state).
- **Failed service-to-service communication**: stop user-service or
  product-service → order creation returns `502 upstream_unavailable`
  after exhausted retries.
- **Slow API response / timeout**: a slow product-service response beyond
  `HTTP_CLIENT_TIMEOUT_SECONDS` surfaces the same way.
- **Payment declined vs. payment-service unreachable**: both leave the
  order in the database with `status=FAILED` (visible via `GET
  /api/v1/orders/{id}`) instead of silently vanishing — a good scenario for
  practicing reconciliation/incident response.
- **DNS/service discovery problem**: point `PRODUCT_SERVICE_URL` at an
  unresolvable host to reproduce a DNS failure distinct from a plain
  connection refusal.
