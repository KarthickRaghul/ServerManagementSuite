-- Updated schema to avoid NullString
CREATE TABLE IF NOT EXISTS server_devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ip VARCHAR(45) NOT NULL UNIQUE,
    tag VARCHAR(100) NOT NULL DEFAULT '',  -- NOT NULL with default
    os VARCHAR(100) NOT NULL DEFAULT '',   -- NOT NULL with default
    access_token VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

