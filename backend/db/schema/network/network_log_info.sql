CREATE TABLE IF NOT EXISTS network_log_info (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_info VARCHAR(255) NOT NULL
);

