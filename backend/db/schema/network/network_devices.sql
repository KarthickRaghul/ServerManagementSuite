CREATE TABLE IF NOT EXISTS network_devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mac VARCHAR(17) NOT NULL,
    ip VARCHAR(45) NOT NULL,
    tag VARCHAR(100),
    device VARCHAR(100)
);

