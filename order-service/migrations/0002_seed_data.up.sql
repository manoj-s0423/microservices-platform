-- Seed data for local development. References the seeded user
-- (22222222-2222-2222-2222-222222222222 from user-service) and seeded
-- product (aaaaaaaa-0000-0000-0000-000000000001 from product-service) so a
-- freshly-seeded stack has one example order to inspect end-to-end.

INSERT INTO orders (id, user_id, status, total_cents, currency, created_at, updated_at)
VALUES (
    'bbbbbbbb-0000-0000-0000-000000000001',
    '22222222-2222-2222-2222-222222222222',
    'CONFIRMED',
    3998,
    'USD',
    now(),
    now()
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO order_items (id, order_id, product_id, sku, quantity, unit_price_cents)
VALUES (
    'cccccccc-0000-0000-0000-000000000001',
    'bbbbbbbb-0000-0000-0000-000000000001',
    'aaaaaaaa-0000-0000-0000-000000000001',
    'SKU-WIRELESS-MOUSE',
    2,
    1999
)
ON CONFLICT (id) DO NOTHING;
