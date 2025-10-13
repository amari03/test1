CREATE TABLE attendance (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    officer_id UUID NOT NULL REFERENCES officers(id) ON DELETE CASCADE,
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    status TEXT NOT NULL CHECK (status IN ('attended', 'absent', 'excused')),
    credited_hours NUMERIC(4, 1) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (officer_id, session_id)
);

CREATE TABLE session_feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    facilitator_id UUID NOT NULL REFERENCES facilitators(id) ON DELETE CASCADE,
    officer_id UUID NOT NULL REFERENCES officers(id) ON DELETE CASCADE,
    rating NUMERIC(2, 1) NOT NULL,
    comments TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (officer_id, session_id, facilitator_id)
);