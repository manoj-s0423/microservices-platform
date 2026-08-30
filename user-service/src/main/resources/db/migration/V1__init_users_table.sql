-- User Service schema, migration 1
-- Owned exclusively by user-service. No other service may read/write this
-- table directly; access is only via the user-service API.

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           VARCHAR(255) NOT NULL UNIQUE,
    password_hash   VARCHAR(255) NOT NULL,
    first_name      VARCHAR(100) NOT NULL,
    last_name       VARCHAR(100) NOT NULL,
    role            VARCHAR(20)  NOT NULL DEFAULT 'CUSTOMER',
    status          VARCHAR(20)  NOT NULL DEFAULT 'ACTIVE',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_status ON users (status);

COMMENT ON TABLE users IS 'Registered ShopStream users (customers and staff).';
COMMENT ON COLUMN users.role IS 'CUSTOMER | ADMIN | SUPPORT';
COMMENT ON COLUMN users.status IS 'ACTIVE | SUSPENDED | DELETED';
