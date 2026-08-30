-- Seed data for local development / integration testing only.
-- Password hash below is bcrypt("Password123!") - never use in production.

INSERT INTO users (id, email, password_hash, first_name, last_name, role, status)
VALUES
    ('11111111-1111-1111-1111-111111111111', 'admin@shopstream.dev', '$2a$10$7EqJtq98hPqEX7fNZaFWoOhi5L5D5vQb0BOKa3Q0G5eYQqW1Q5Q1u', 'Ada', 'Admin', 'ADMIN', 'ACTIVE'),
    ('22222222-2222-2222-2222-222222222222', 'jane.doe@shopstream.dev', '$2a$10$7EqJtq98hPqEX7fNZaFWoOhi5L5D5vQb0BOKa3Q0G5eYQqW1Q5Q1u', 'Jane', 'Doe', 'CUSTOMER', 'ACTIVE'),
    ('33333333-3333-3333-3333-333333333333', 'suspended.user@shopstream.dev', '$2a$10$7EqJtq98hPqEX7fNZaFWoOhi5L5D5vQb0BOKa3Q0G5eYQqW1Q5Q1u', 'Sam', 'Suspended', 'CUSTOMER', 'SUSPENDED')
ON CONFLICT (email) DO NOTHING;
