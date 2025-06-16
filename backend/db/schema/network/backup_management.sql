CREATE TABLE IF NOT EXISTS backup_management (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date TIMESTAMP NOT NULL,
    status VARCHAR(20) NOT NULL,
    name VARCHAR(255) NOT NULL,
    time TIMESTAMP NOT NULL,
    size BIGINT NOT NULL,
    device_id UUID REFERENCES network_devices(id) ON DELETE CASCADE
);

