# API Gateway

Single entry point for external traffic into the ShopStream platform. Handles
auth token verification, request/response logging with correlation IDs, rate
limiting, and reverse-proxying to downstream microservices with
timeout+retry resilience.

## Overview

| Field | Value |
|---|---|
| Language | JavaScript (Node.js) |
| Framework | Express 4 |
| Runtime version | Node.js >= 20 (tested on 20.x LTS) |
| Build tool | npm |
| Application port | `3000` |
| Database | None (stateless) |

## Build

```bash
npm install
```

## Test

```bash
npm test              # unit + integration, with coverage
npm run test:unit
npm run test:integration
```

Integration tests mock downstream services with `nock` — no other services
need to be running to test the gateway in isolation.

## Run

```bash
npm start          # production mode
npm run dev         # nodemon, auto-reload
```

## Environment variables

See [.env.example](.env.example). Key variables:

| Variable | Required | Description |
|---|---|---|
| `PORT` | no (default 3000) | HTTP port |
| `JWT_SECRET` | **yes** | HMAC secret for verifying tokens issued by user-service. **Must be set to the exact same value as user-service's `JWT_SECRET`** — the gateway only verifies tokens, it never issues them, so a mismatch here doesn't fail startup, it fails every authenticated request at runtime with `401 invalid_token`. Auth routes fail with `500 server_misconfigured` if unset entirely — a deliberate, reproducible misconfiguration scenario. |
| `USER_SERVICE_URL` | yes | Base URL for user-service |
| `PRODUCT_SERVICE_URL` | yes | Base URL for product-service |
| `ORDER_SERVICE_URL` | yes | Base URL for order-service |
| `PAYMENT_SERVICE_URL` | yes | Base URL for payment-service (not called directly by gateway today, reserved) |
| `NOTIFICATION_SERVICE_URL` | yes | Base URL for notification-service (reserved) |
| `HTTP_TIMEOUT_MS` | no (3000) | Per-request timeout to downstream services |
| `HTTP_RETRY_ATTEMPTS` | no (2) | Retries on network error / 502/503/504 |
| `RATE_LIMIT_MAX_REQUESTS` | no (100) | Requests per window per IP |

## Dependencies

Runtime: express, axios, jsonwebtoken, helmet, cors, express-rate-limit, pino.
See [package.json](package.json) for exact versions.

## Database requirements

None. The gateway is stateless; all persistence lives in the owning services.

## API endpoints

| Method | Path | Auth | Proxies to |
|---|---|---|---|
| POST | `/api/v1/auth/login` | no | user-service `POST /api/v1/auth/login` |
| POST | `/api/v1/auth/register` | no | user-service `POST /api/v1/auth/register` |
| GET | `/api/v1/users/me` | yes | user-service `GET /api/v1/users/{id}` |
| GET | `/api/v1/users/:id` | yes | user-service `GET /api/v1/users/{id}` |
| GET | `/api/v1/products` | no | product-service `GET /api/v1/products` |
| GET | `/api/v1/products/:id` | no | product-service `GET /api/v1/products/{id}` |
| POST | `/api/v1/orders` | yes | order-service `POST /api/v1/orders` |
| GET | `/api/v1/orders` | yes | order-service `GET /api/v1/orders` |
| GET | `/api/v1/orders/:id` | yes | order-service `GET /api/v1/orders/{id}` |

## Health endpoint

`GET /health` — liveness only, no downstream calls. Always 200 while the
process is up.

## Readiness endpoint

`GET /ready` — calls `/health` on every downstream service (1.5s timeout
each) and returns `200 UP` only if all are reachable, else `503 DEGRADED`
with a per-service breakdown. Use this for the Kubernetes readiness probe;
use `/health` for the liveness probe.

## Service-to-service dependencies

- user-service (auth, profile)
- product-service (catalog)
- order-service (order placement/lookup)

## Failure scenarios this service is designed to reproduce

- **Missing/incorrect env var**: unset `JWT_SECRET` → all authenticated
  routes return `500 server_misconfigured`.
- **Downstream unavailable**: stop a downstream service → gateway returns
  `502 bad_gateway` after exhausting retries.
- **Slow downstream / timeout**: a downstream service that hangs past
  `HTTP_TIMEOUT_MS` → gateway returns `504 gateway_timeout`.
- **DNS/service discovery failure**: point a `*_SERVICE_URL` at a
  non-resolvable host → `ENOTFOUND`, surfaced as `502 bad_gateway`.
- **Rate limiting**: exceed `RATE_LIMIT_MAX_REQUESTS` in a window → `429`.
