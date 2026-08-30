-- Seed data for local development. References the seeded order from
-- order-service (bbbbbbbb-0000-0000-0000-000000000001).

INSERT INTO payments (id, order_id, user_id, amount_cents, currency, status, transaction_id, idempotency_key, created_at, updated_at)
VALUES (
    'dddddddd-0000-0000-0000-000000000001',
    'bbbbbbbb-0000-0000-0000-000000000001',
    '22222222-2222-2222-2222-222222222222',
    3998,
    'USD',
    'Succeeded',
    'sim_txn_seed0000000000000000000000',
    'bbbbbbbb-0000-0000-0000-000000000001',
    now(),
    now()
)
ON CONFLICT (idempotency_key) DO NOTHING;
