CREATE TABLE officers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    regulation_number TEXT UNIQUE,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    sex TEXT NOT NULL CHECK (sex IN ('male', 'female', 'unknown')),
    rank_code TEXT NOT NULL REFERENCES ranks(code) ON DELETE RESTRICT,
    region_id TEXT REFERENCES regions(id) ON DELETE SET NULL,
    formation_id TEXT REFERENCES formations(id) ON DELETE SET NULL,
    posting_id TEXT REFERENCES postings(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    archived_at TIMESTAMPTZ
);