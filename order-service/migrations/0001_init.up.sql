-- Order Service schema, migration 1
-- Owned exclusively by order-service.

CREATE TABLE orders (
    id           UUID PRIMARY KEY,
    user_id      UUID NOT NULL,
    status       VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    total_cents  INTEGER NOT NULL CHECK (total_cents >= 0),
    currency     VARCHAR(3) NOT NULL DEFAULT 'USD',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE order_items (
    id                UUID PRIMARY KEY,
    order_id          UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id        UUID NOT NULL,
    sku               VARCHAR(64) NOT NULL,
    quantity          INTEGER NOT NULL CHECK (quantity > 0),
    unit_price_cents  INTEGER NOT NULL CHECK (unit_price_cents >= 0)
);

CREATE INDEX idx_orders_user_id ON orders (user_id);
CREATE INDEX idx_orders_status ON orders (status);
CREATE INDEX idx_order_items_order_id ON order_items (order_id);

COMMENT ON TABLE orders IS 'Orders placed by users. user_id and product references are logical foreign keys to user-service/product-service - never enforced at the DB level across service boundaries.';
