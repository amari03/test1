CREATE TABLE courses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    category TEXT NOT NULL CHECK (category IN ('mandatory', 'elective', 'instructor')),
    default_credit_hours NUMERIC(4, 1) NOT NULL,
    description TEXT,
    created_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ
);

CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    start_datetime TIMESTAMPTZ NOT NULL,
    end_datetime TIMESTAMPTZ NOT NULL,
    region_id TEXT REFERENCES regions(id) ON DELETE SET NULL,
    formation_id TEXT REFERENCES formations(id) ON DELETE SET NULL,
    posting_id TEXT REFERENCES postings(id) ON DELETE SET NULL,
    location_text TEXT,
    facilitator_notes TEXT,
    credit_hours_override NUMERIC(4, 1),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ
);