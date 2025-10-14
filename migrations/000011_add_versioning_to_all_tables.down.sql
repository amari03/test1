ALTER TABLE facilitators DROP COLUMN version;
ALTER TABLE users DROP COLUMN version;
ALTER TABLE courses DROP COLUMN version;
ALTER TABLE sessions DROP COLUMN version;
ALTER TABLE session_facilitators DROP COLUMN version;
ALTER TABLE attendance DROP COLUMN version;
ALTER TABLE session_feedback DROP COLUMN version;
ALTER TABLE import_jobs DROP COLUMN version;