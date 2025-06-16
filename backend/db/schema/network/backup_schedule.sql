CREATE TABLE IF NOT EXISTS backup_schedule (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    backup_name VARCHAR(255) NOT NULL,
    scheduled_time TIMESTAMP NOT NULL,
    device_id UUID REFERENCES network_devices(id) ON DELETE CASCADE
);

