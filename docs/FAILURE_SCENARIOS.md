# Failure Scenario Catalog

This platform was deliberately built so each of the following production
incidents is reproducible on demand — useful for practicing
troubleshooting, alerting, and incident response once it's deployed to
Kubernetes. Each row says which service, what to change, and what you
should observe.

| # | Scenario | Where to trigger it | How | What you should observe |
|---|---|---|---|---|
| 1 | Service unavailable | any service | Stop the process/pod | Gateway (or whichever caller) returns `502`/`bad_gateway` after retries exhaust; caller's own `/ready` may go `DEGRADED` |
| 2 | Database connection failure | user-service, product-service, order-service, payment-service | Set `DB_HOST`/`DB_PASSWORD` wrong, or stop the DB container | Startup failure (user/order/payment fail fast and exit) or `/ready` → `503` after startup (product-service) |
| 3 | Incorrect environment variable | api-gateway, user-service, notification-service | Unset `JWT_SECRET` / `MONGO_URI` | Gateway: `500 server_misconfigured` on auth routes. user-service: refuses to start (`IllegalStateException`). notification-service: exits at startup. |
| 4 | Container crash | any service | `kill -9` the process | Kubernetes-level: pod restarts; check whether in-flight requests were lost (compare to graceful SIGTERM shutdown, scenario 12) |
| 5 | Out-of-memory condition | any service | Constrain container memory limit below steady-state usage (once containerized) | OOMKilled pod; JVM/Node/CLR/Go each report differently — good for comparing OOM signatures across runtimes |
| 6 | CPU throttling | any service | Constrain CPU limit aggressively (once containerized) | Elevated latency without errors — pairs well with the gateway's `/ready` timeout and order-service's `HTTP_CLIENT_TIMEOUT_SECONDS` |
| 7 | Slow API response | product-service, order-service, payment-service (simulated gateway), notification-service (simulated email) | Add artificial latency, or let real load degrade response time | Caller-side timeout fires (`504 gateway_timeout` at the gateway, `upstream_unavailable` at order-service) before the slow call ever completes |
| 8 | Failed service-to-service communication | order-service → user-service/product-service/payment-service | Stop the callee, or point its `*_SERVICE_URL` at nothing | order-service returns `502 upstream_unavailable`; a payment-charge failure specifically leaves the order `FAILED` rather than vanishing |
| 9 | DNS/service discovery problem | api-gateway, order-service | Point a `*_SERVICE_URL` at a non-resolvable hostname | `ENOTFOUND`/DNS resolution error, surfaced as `502` — distinguishable in logs from a plain connection refusal |
| 10 | Incorrect port configuration | any service | Set `PORT`/`SERVER_PORT`/`ASPNETCORE_URLS` to a value that conflicts with another process, or mismatched from the Kubernetes Service | Bind failure at startup, or a Service routing to the wrong port (traffic reaches the pod but never the app) |
| 11 | Failed deployment | any service | Deploy a build that fails its readiness probe (e.g. missing env var) | Once on Kubernetes: rollout stalls/rolls back if `readinessProbe` is wired to `/ready` |
| 12 | Image vulnerability | any service | Once Dockerized, scan with Trivy against a deliberately outdated base image | Trivy reports CVEs — practice triage and base-image bumping |
| 13 | Application startup failure | user-service, order-service, notification-service, payment-service | Break required config (see #3) or point at an unreachable DB | Process exits non-zero immediately, distinguishable in logs from a runtime failure |
| 14 | Broken health check | product-service, order-service, payment-service, notification-service | Stop the DB after the service has already started | `/ready` (or `/actuator/health`, or `GET /ready`) flips to `DOWN`/`DEGRADED` while `/health` (liveness) stays `UP` — proves the liveness/readiness split is doing its job |
| 15 | Dependency failure (external, non-ShopStream) | payment-service, notification-service | Set `Gateway__Mode=live` / `EMAIL_PROVIDER_MODE=live` with a bad URL | Payment: charge fails with `gateway_unavailable`, order left `FAILED`. Notification: `provider_unavailable` after retries. |

## Extra scenarios worth practicing once containerized

- **Double-charge safety**: retry a payment charge for the same order
  (same `Idempotency-Key`/orderId) and confirm payment-service returns the
  original result instead of charging twice — see
  `PaymentProcessingServiceTests.ChargeAsync_RepeatedIdempotencyKey_*` for
  the behavior being relied on.
- **Graceful vs. forced shutdown**: send `SIGTERM` to a service under load
  and confirm in-flight requests complete before the process exits, vs.
  `SIGKILL` which drops them immediately — every service's
  `ShutdownTimeout`/`SHUTDOWN_TIMEOUT_*` setting controls the grace window.
- **Rate limiting**: exceed `RATE_LIMIT_MAX_REQUESTS` against the API
  Gateway and confirm `429` responses with correct `RateLimit-*` headers.
