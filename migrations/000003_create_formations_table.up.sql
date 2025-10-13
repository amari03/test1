CREATE TABLE formations (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    -- Foreign key to the regions table
    region_id TEXT NOT NULL REFERENCES regions(id) ON DELETE RESTRICT
);