CREATE TABLE IF NOT EXISTS network_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    severity VARCHAR(20) NOT NULL,
    content TEXT NOT NULL,
    time TIMESTAMP NOT NULL DEFAULT NOW(),
    device_id UUID REFERENCES network_devices(id) ON DELETE CASCADE
);

