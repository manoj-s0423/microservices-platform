# Product Service

Owns the product catalog: CRUD operations, category filtering, pagination,
and stock levels. Sole owner of the `shopstream_products` PostgreSQL
database.

## Overview

| Field | Value |
|---|---|
| Language | Python 3.12 |
| Framework | FastAPI |
| Runtime version | Python 3.12+ |
| Build tool | pip |
| Application port | `8000` |
| Database | PostgreSQL 15+ |

## Build

```bash
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements-dev.txt
```

## Test

```bash
pytest -v --cov=app
```

Tests run against an in-memory SQLite database (see
[tests/conftest.py](tests/conftest.py)), so no PostgreSQL instance is
required to run the suite. 11 tests cover CRUD, validation, pagination,
category filtering, soft delete, and health/readiness.

## Run

```bash
uvicorn app.main:app --host 0.0.0.0 --port 8000
# or, for local auto-reload:
uvicorn app.main:app --reload --port 8000
```

## Environment variables

See [.env.example](.env.example).

| Variable | Required | Description |
|---|---|---|
| `PORT` | no (8000) | HTTP port (informational; actual bind port is passed to uvicorn) |
| `DB_HOST` / `DB_PORT` / `DB_NAME` / `DB_USER` / `DB_PASSWORD` | yes | PostgreSQL connection |
| `DB_CONNECT_TIMEOUT_SECONDS` | no (5) | TCP connect timeout to Postgres |
| `DEFAULT_PAGE_SIZE` / `MAX_PAGE_SIZE` | no | Pagination bounds |

Config is validated at import time via pydantic-settings
([app/config.py](app/config.py)) — a malformed variable (e.g. non-numeric
`DB_PORT`) fails fast with a clear validation error.

## Dependencies

FastAPI, SQLAlchemy 2.0, psycopg2-binary, Alembic, pydantic-settings,
structlog, httpx, tenacity. See [requirements.txt](requirements.txt).

## Database requirements

PostgreSQL 15+. Schema managed with Alembic migrations in
[migrations/versions](migrations/versions):

- `0001_init_products_table.py` — creates the `products` table
- `0002_seed_data.py` — seeds 3 sample products for local development

### Local development

```bash
docker run --name shopstream-products-db -e POSTGRES_DB=shopstream_products \
  -e POSTGRES_USER=shopstream -e POSTGRES_PASSWORD=devpassword \
  -p 5433:5432 -d postgres:15

alembic upgrade head
```

## API endpoints

| Method | Path | Description |
|---|---|---|
| GET | `/api/v1/products` | List products (query: `category`, `page`, `page_size`) |
| GET | `/api/v1/products/{id}` | Get a product by ID |
| POST | `/api/v1/products` | Create a product |
| PATCH | `/api/v1/products/{id}` | Partially update a product |
| DELETE | `/api/v1/products/{id}` | Soft-delete (sets `is_active=false`) |

## Health endpoint

`GET /health` — liveness only, no dependency checks.

## Readiness endpoint

`GET /ready` — runs `SELECT 1` against PostgreSQL; returns `503` with
`{"dependencies": {"database": "DOWN"}}` if unreachable.

## Service-to-service dependencies

None outbound. This service is a dependency **of** the API Gateway and the
Order Service (order-service calls product-service to validate SKUs and
prices at order-placement time).

## Failure scenarios this service is designed to reproduce

- **Database connection failure**: wrong `DB_HOST`/`DB_PASSWORD` → `/ready`
  returns `503`; requests touching the DB raise `OperationalError`.
- **Incorrect environment variable**: non-numeric `DB_PORT` → pydantic
  validation error at startup, process exits immediately.
- **Duplicate SKU / not-found / validation errors**: exercised directly by
  `tests/test_products.py`.
- **Slow API response**: add an artificial `time.sleep()` in
  `product_service.list_products` (or throttle the DB) to practice
  diagnosing latency from the gateway's `/ready` and timeout behavior.
