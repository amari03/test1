CREATE TABLE users (
id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
email TEXT UNIQUE NOT NULL,
password_hash TEXT NOT NULL,
role TEXT NOT NULL CHECK (role IN ('admin', 'contributor', 'viewer')),
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
last_login_at TIMESTAMPTZ
);