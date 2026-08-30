# Payment Service

Charges orders against the platform's one external, third-party
dependency: a card-processing gateway. Idempotent on order ID so retried
charge requests (from order-service's own resilience policy) never
double-charge. Sole owner of the `shopstream_payments` PostgreSQL database.

## Overview

| Field | Value |
|---|---|
| Language | C# |
| Framework | ASP.NET Core 8 (Web API) |
| Runtime version | .NET 8 SDK |
| Build tool | dotnet CLI / MSBuild |
| Application port | `8083` |
| Database | PostgreSQL 15+ |

## Build

```bash
dotnet restore
dotnet build
```

## Test

```bash
dotnet test
```

10 xUnit tests across two files:
- `PaymentProcessingServiceTests.cs` — business logic against an EF Core
  in-memory provider and a mocked `IPaymentGatewayClient` (approved,
  declined, gateway-unavailable, idempotent replay, invalid input).
- `PaymentsControllerTests.cs` — controller-level "API tests": request DTO
  in, HTTP status code + response DTO out.

No PostgreSQL instance or real external gateway is required to run
`dotnet test`.

## Run

```bash
export $(grep -v '^#' .env | xargs)
dotnet run --project src/PaymentService.Api
```

## Environment variables

See [.env.example](.env.example).

| Variable | Required | Description |
|---|---|---|
| `ASPNETCORE_URLS` | no | Bind address/port (default maps to 8083 via launch config) |
| `DB_HOST` / `DB_PORT` / `DB_NAME` / `DB_USER` / `DB_PASSWORD` | yes | PostgreSQL connection |
| `Gateway__Mode` | no (`simulated`) | `simulated` (no external call, deterministic approve/decline) or `live` (real HTTP calls to `Gateway__BaseUrl`) |
| `Gateway__BaseUrl` | required if `Mode=live` | External gateway base URL |
| `Gateway__TimeoutSeconds` | no (5) | Per-call timeout to the gateway |
| `Gateway__RetryAttempts` | no (2) | Polly retry attempts on transient HTTP errors |
| `Gateway__DeclineThresholdCents` | no (1,000,000) | Simulated-mode amount above which a charge is auto-declined |

.NET's configuration binder maps `Gateway__Mode` (env var, double
underscore) to `Gateway:Mode` (`appsettings.json` section) automatically.

## Dependencies

ASP.NET Core, EF Core + Npgsql, AspNetCore.HealthChecks.NpgSql,
Microsoft.Extensions.Http.Polly, Serilog.AspNetCore. See
[PaymentService.Api.csproj](src/PaymentService.Api/PaymentService.Api.csproj).

## Database requirements

PostgreSQL 15+. Schema-first: `PaymentDbContext` maps to a schema created
by plain SQL scripts in [migrations](migrations) rather than EF Core
Migrations, so DevOps can apply it with whatever tool is standardized
across the platform:

- `001_init.sql` — creates the `payments` table
- `002_seed_data.sql` — one seeded successful payment for the seeded order

### Local development

```bash
docker run --name shopstream-payments-db -e POSTGRES_DB=shopstream_payments \
  -e POSTGRES_USER=shopstream -e POSTGRES_PASSWORD=devpassword \
  -p 5435:5432 -d postgres:15

psql "$CONNECTION_STRING" -f migrations/001_init.sql
psql "$CONNECTION_STRING" -f migrations/002_seed_data.sql
```

## API endpoints

| Method | Path | Description |
|---|---|---|
| POST | `/api/v1/payments` | Charge an order (idempotent via `Idempotency-Key` header, falls back to `orderId`) |
| GET | `/api/v1/payments/{id}` | Get a payment by ID |

## Health endpoint

`GET /health` — liveness only (`HealthCheckOptions.Predicate` excludes all
checks), no dependency calls.

## Readiness endpoint

`GET /ready` — runs the `"ready"`-tagged PostgreSQL health check.

## Service-to-service / external dependencies

- **Inbound**: called by order-service.
- **Outbound (external, non-ShopStream)**: the card-processing gateway,
  behind `IPaymentGatewayClient`. In `simulated` mode (default for local
  dev/CI) no real network call is made. In `live` mode, calls are wrapped
  in a Polly retry policy on transient HTTP errors (5xx/408).

## Failure scenarios this service is designed to reproduce

- **Database connection failure**: wrong `DB_HOST` → `/ready` returns
  `503`; the `/api/v1/payments` endpoints throw on first DB access.
- **External API dependency failure**: set `Gateway__Mode=live` with a
  wrong/unreachable `Gateway__BaseUrl` → charges fail with
  `gateway_unavailable`, order stays in `PENDING`/`FAILED` (see
  order-service).
- **Declined payment vs. gateway outage**: distinguishable in the response
  (`DECLINED` with a `reason` vs. `502`/`FAILED`) — good for practicing
  alerting rules that must tell "customer's card was declined" apart from
  "our payment path is broken."
- **Double-charge prevention**: the idempotency test proves a retried
  charge for the same order never calls the gateway twice — remove the
  idempotency check locally to see the alternative (unsafe) behavior.
- **Incorrect environment variable**: an unset/blank `DB_PASSWORD` fails
  the DB connection with an auth error, distinguishable in logs from a
  network-level "host unreachable" failure.
