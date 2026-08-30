-- Payment Service schema, migration 1
-- Owned exclusively by payment-service. PaymentDbContext (EF Core) maps to
-- this schema directly ("schema-first"); no EF Core migrations are used so
-- that DevOps can apply this with whatever migration tool is standardized
-- across the platform (Flyway, sqitch, a Jenkins step running psql, etc.).

CREATE TABLE payments (
    id                UUID PRIMARY KEY,
    order_id          UUID NOT NULL,
    user_id           UUID NOT NULL,
    amount_cents      INTEGER NOT NULL CHECK (amount_cents > 0),
    currency          VARCHAR(3) NOT NULL DEFAULT 'USD',
    status            VARCHAR(20) NOT NULL DEFAULT 'Pending',
    transaction_id    VARCHAR(100),
    failure_reason    VARCHAR(255),
    idempotency_key   VARCHAR(100) NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX ux_payments_idempotency_key ON payments (idempotency_key);
CREATE INDEX idx_payments_order_id ON payments (order_id);
CREATE INDEX idx_payments_user_id ON payments (user_id);

COMMENT ON TABLE payments IS 'Charge attempts against orders. order_id/user_id are logical references to order-service/user-service, never enforced cross-database.';
COMMENT ON COLUMN payments.status IS 'Pending | Succeeded | Failed | Declined';
