# User Service

Owns user identity: registration, authentication (JWT issuance), and profile
lookups. Sole owner of the `shopstream_users` PostgreSQL database.

## Overview

| Field | Value |
|---|---|
| Language | Java 21 |
| Framework | Spring Boot 3.3 (Web, Data JPA, Security, Actuator) |
| Runtime version | Java 21 (LTS) |
| Build tool | Maven |
| Application port | `8081` |
| Database | PostgreSQL 15+ |

## Build

```bash
mvn clean package
```

Produces `target/user-service.jar`.

## Test

```bash
mvn test
```

Unit tests (`UserServiceTest`) mock the repository/JWT layers with Mockito.
Integration tests (`AuthControllerIntegrationTest`) boot the full Spring
context against an in-memory H2 database (profile `test`, see
[application-test.yml](src/main/resources/application-test.yml)) — no
PostgreSQL required to run `mvn test`.

## Run

```bash
export $(grep -v '^#' .env | xargs)   # or set vars via your process manager
java -jar target/user-service.jar
```

For local development against a real PostgreSQL instance, run migrations
automatically on boot (Flyway is enabled by default) — just point `DB_*` at
your database.

## Environment variables

See [.env.example](.env.example).

| Variable | Required | Description |
|---|---|---|
| `SERVER_PORT` | no (8081) | HTTP port |
| `DB_HOST` / `DB_PORT` / `DB_NAME` / `DB_USER` / `DB_PASSWORD` | yes | PostgreSQL connection |
| `DB_POOL_SIZE` | no (10) | HikariCP max pool size |
| `JWT_SECRET` | **yes** | HMAC signing key. **Must be set to the exact same value as api-gateway's `JWT_SECRET`** — user-service issues tokens, the gateway only verifies them, and a mismatch produces `401 invalid_token` at the gateway with no error here. App **refuses to start** if unset entirely (see `JwtUtil.init()`) — a deliberate "incorrect env var" failure scenario. |
| `JWT_EXPIRATION_MS` | no (3600000) | Token TTL |

## Dependencies

Spring Boot Web, Spring Data JPA, Spring Security, Spring Boot Actuator,
Flyway, PostgreSQL JDBC driver, jjwt, logstash-logback-encoder (structured
JSON logs). See [pom.xml](pom.xml).

## Database requirements

PostgreSQL 15+. Schema and seed data are managed via Flyway migrations in
[src/main/resources/db/migration](src/main/resources/db/migration):

- `V1__init_users_table.sql` — creates the `users` table
- `V2__seed_data.sql` — local dev seed users (bcrypt hash of `Password123!`)

### Local development

```bash
docker run --name shopstream-users-db -e POSTGRES_DB=shopstream_users \
  -e POSTGRES_USER=shopstream -e POSTGRES_PASSWORD=devpassword \
  -p 5432:5432 -d postgres:15
```

Migrations run automatically on application startup.

## API endpoints

| Method | Path | Description |
|---|---|---|
| POST | `/api/v1/auth/register` | Create a new customer account, returns a JWT |
| POST | `/api/v1/auth/login` | Authenticate, returns a JWT |
| GET | `/api/v1/users/{id}` | Fetch a user profile by ID |

## Health endpoint

`GET /actuator/health` — includes DB connectivity (`db` health indicator).

## Readiness endpoint

`GET /actuator/health/readiness` — Spring Boot's readiness probe group
(excludes liveness-only checks). Use for the Kubernetes readiness probe.

`GET /actuator/health/liveness` — for the liveness probe.

## Service-to-service dependencies

None inbound-required; this service has no outbound dependencies on other
microservices. It is a dependency **of** the API Gateway and (indirectly, for
identity lookups) the Order Service.

## Failure scenarios this service is designed to reproduce

- **Missing/incorrect env var**: unset `JWT_SECRET` → application fails to
  start (`IllegalStateException` at boot) — practice diagnosing a
  CrashLoopBackOff from a config error.
- **Database connection failure**: wrong `DB_HOST`/`DB_PASSWORD` → Hikari
  connection timeout at startup, readiness probe fails.
- **Broken health check**: kill the DB after startup → `/actuator/health`
  flips to `DOWN` (db indicator) while the process keeps running.
- **Duplicate registration / validation errors**: exercised directly by
  `AuthControllerIntegrationTest`.
