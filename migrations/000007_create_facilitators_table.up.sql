CREATE TABLE facilitators (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    officer_id UUID REFERENCES officers(id) ON DELETE SET NULL,
    first_name TEXT,
    last_name TEXT,
    rank_code TEXT REFERENCES ranks(code) ON DELETE SET NULL,
    posting_id TEXT REFERENCES postings(id) ON DELETE SET NULL,
    notes TEXT
);

CREATE TABLE session_facilitators (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    facilitator_id UUID NOT NULL REFERENCES facilitators(id) ON DELETE CASCADE,
    role TEXT,
    UNIQUE (session_id, facilitator_id)
);