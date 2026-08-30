# Notification Service

Sends order/account notifications over email (SMS schema/routes exist but
are intentionally not dispatched - see below). Sole owner of the
`shopstream_notifications` MongoDB database.

## Overview

| Field | Value |
|---|---|
| Language | TypeScript |
| Framework | Express 4 |
| Runtime version | Node.js >= 20 |
| Build tool | npm / tsc |
| Application port | `8084` |
| Database | MongoDB 6+ |

## Build

```bash
npm install
npm run build   # compiles src/ -> dist/
```

## Test

```bash
npm test
```

14 tests across 3 files, run against an **in-memory MongoDB**
(`mongodb-memory-server`) — no real MongoDB instance required:
- `notificationService.test.ts` — business logic (send/get/list) against a
  fake `EmailProvider`.
- `health.test.ts` — liveness always up; readiness flips 503→200 as the DB
  connects.
- `notifications.route.test.ts` — HTTP-level "API tests" via `supertest`.

## Run

```bash
export $(grep -v '^#' .env | xargs)
npm run build && npm start
# or for local development with auto-reload:
npm run dev
```

## Environment variables

See [.env.example](.env.example).

| Variable | Required | Description |
|---|---|---|
| `PORT` | no (8084) | HTTP port |
| `MONGO_URI` | **yes** | MongoDB connection string. Process exits at startup if unset. |
| `MONGO_CONNECT_TIMEOUT_MS` | no (5000) | Server selection timeout |
| `EMAIL_PROVIDER_MODE` | no (`simulated`) | `simulated` (default, no external call) or `live` |
| `EMAIL_PROVIDER_API_URL` / `_API_KEY` | required if `Mode=live` | Transactional email API (the external, third-party dependency) |
| `EMAIL_PROVIDER_TIMEOUT_MS` / `_RETRY_ATTEMPTS` | no | Resilience knobs for the live provider |

## Dependencies

Express, Mongoose, Helmet, CORS, pino/pino-http. See
[package.json](package.json).

## Database requirements

MongoDB 6+. As a document store there's no DDL migration; schema shape is
enforced by the Mongoose model
([src/models/Notification.ts](src/models/Notification.ts)). Indexes and
seed data are provided as mongo shell scripts in [scripts](scripts):

- `init-indexes.js` — indexes on `userId+createdAt` and `status`
- `seed-data.js` — 2 sample notifications for the seeded user/order

### Local development

```bash
docker run --name shopstream-notifications-db \
  -p 27017:27017 -d mongo:6

mongosh "$MONGO_URI" scripts/init-indexes.js
mongosh "$MONGO_URI" scripts/seed-data.js
```

## API endpoints

| Method | Path | Description |
|---|---|---|
| POST | `/api/v1/notifications` | Send a notification (persists then dispatches via the configured provider) |
| GET | `/api/v1/notifications/{id}` | Get a notification by ID |
| GET | `/api/v1/notifications?userId={id}` | List a user's notifications, newest first |

`channel` accepts `EMAIL` or `SMS`; SMS notifications are persisted with
`status=FAILED, failureReason=sms_channel_not_implemented` — the schema and
routes are ready for an SMS provider, but none is wired up in this build.

## Health endpoint

`GET /health` — liveness only, no dependency checks.

## Readiness endpoint

`GET /ready` — checks `mongoose.connection.readyState`; `503` if not
connected.

## Service-to-service dependencies

None inbound-required from other ShopStream services in this build (in a
full platform, order-service/payment-service would call this service to
trigger order/payment notifications — see the "how it fits together"
diagram in the [root README](../README.md)).

**Outbound (external, non-ShopStream)**: a transactional email provider
behind `EmailProvider` (`simulated` by default; `live` mode calls a real
HTTP API with timeout + bounded retries).

## Failure scenarios this service is designed to reproduce

- **Database connection failure / incorrect environment variable**: unset
  or wrong `MONGO_URI` → process exits immediately at startup with a clear
  log line, rather than hanging in mongoose's default retry loop.
- **Broken health check**: kill MongoDB after startup → `/ready` flips to
  `503` while the process keeps running.
- **External API dependency failure**: `EMAIL_PROVIDER_MODE=live` with a
  wrong/unreachable `EMAIL_PROVIDER_API_URL` → notifications persist with
  `status=FAILED, failureReason=provider_unavailable` after retries,
  rather than silently vanishing.
- **Slow API response**: a slow email provider surfaces via the
  `EMAIL_PROVIDER_TIMEOUT_MS`-bounded retry loop in
  [src/services/emailProvider.ts](src/services/emailProvider.ts).
