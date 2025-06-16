CREATE TABLE user_sessions (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE REFERENCES users(name) ON DELETE CASCADE,  -- ✅ Now works!
    refresh_token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
