CREATE TABLE IF NOT EXISTS network_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    tag VARCHAR(100),
    log_level VARCHAR(20),
    message TEXT,
    device_id UUID REFERENCES network_devices(id) ON DELETE CASCADE
);

