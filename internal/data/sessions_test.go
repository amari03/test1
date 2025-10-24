package data

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/amari03/test1/internal/validator"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// setupSessionsTestDB sets up a DB environment with users and courses to satisfy session foreign key constraints.
// It returns the DB connection and the ID of a dummy course to be used in tests.
func setupSessionsTestDB(t *testing.T) (*sql.DB, string) {
	dsn := "postgres://test1_test:fishsticks@localhost/test1_test?sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	err = db.Ping()
	require.NoError(t, err, "failed to connect to the test database")

	_, err = db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`)
	require.NoError(t, err)

	// Create dependency tables first: users, then courses.
	db.Exec(`CREATE TABLE IF NOT EXISTS users (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), email TEXT UNIQUE NOT NULL);`)
	db.Exec(`CREATE TABLE IF NOT EXISTS courses (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), title TEXT NOT NULL, category TEXT NOT NULL, default_credit_hours NUMERIC NOT NULL, created_by_user_id UUID NOT NULL REFERENCES users(id), created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ, version INTEGER NOT NULL DEFAULT 1);`)

	// Create the sessions table.
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS sessions (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        course_id UUID NOT NULL REFERENCES courses(id),
        start_datetime TIMESTAMPTZ NOT NULL,
        end_datetime TIMESTAMPTZ NOT NULL,
        location_text TEXT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMPTZ,
        version INTEGER NOT NULL DEFAULT 1
    );`
	_, err = db.Exec(createTableSQL)
	require.NoError(t, err)

	// Insert dummy user and course to get a valid course_id.
	var userID string
	err = db.QueryRow(`INSERT INTO users (email) VALUES ('sessionuser@example.com') ON CONFLICT (email) DO NOTHING RETURNING id;`).Scan(&userID)
	if err == sql.ErrNoRows {
		err = db.QueryRow(`SELECT id FROM users WHERE email = 'sessionuser@example.com'`).Scan(&userID)
	}
	require.NoError(t, err)

	var courseID string
	err = db.QueryRow(`INSERT INTO courses (title, category, default_credit_hours, created_by_user_id) VALUES ('Session Test Course', 'mandatory', 8, $1) RETURNING id`, userID).Scan(&courseID)
	require.NoError(t, err)

	// Cleanup runs in reverse order of creation.
	t.Cleanup(func() {
		_, err := db.Exec("DROP TABLE IF EXISTS sessions;")
		require.NoError(t, err)
		_, err = db.Exec("DROP TABLE IF EXISTS courses;")
		require.NoError(t, err)
		_, err = db.Exec("DROP TABLE IF EXISTS users;")
		require.NoError(t, err)
		db.Close()
	})

	return db, courseID
}

// newTestSession is a helper to create a valid session instance for testing.
func newTestSession(t *testing.T, courseID string) *Session {
	return &Session{
		CourseID: courseID,
		Start:    time.Now().Add(24 * time.Hour),
		End:      time.Now().Add(32 * time.Hour),
		Location: "Training Room A",
	}
}

func TestSessionModel_Insert(t *testing.T) {
	db, courseID := setupSessionsTestDB(t)
	m := SessionModel{DB: db}

	session := newTestSession(t, courseID)

	err := m.Insert(session)
	require.NoError(t, err)

	// Check that the database populated the ID, CreatedAt, and Version.
	require.NotEmpty(t, session.ID)
	require.WithinDuration(t, time.Now(), session.CreatedAt, time.Second)
	require.Equal(t, int32(1), session.Version)

	// Fetch to double-check.
	fetched, err := m.Get(session.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	require.Equal(t, session.CourseID, fetched.CourseID)
	require.Equal(t, session.Location, fetched.Location)
}

func TestSessionModel_Get(t *testing.T) {
	db, courseID := setupSessionsTestDB(t)
	m := SessionModel{DB: db}

	// Insert a record to test Get.
	session := newTestSession(t, courseID)
	err := m.Insert(session)
	require.NoError(t, err)

	// Test successful Get.
	fetched, err := m.Get(session.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)

	// Verify fields match.
	require.Equal(t, session.ID, fetched.ID)
	require.Equal(t, session.CourseID, fetched.CourseID)
	// Use time.Unix() to compare timestamps without monotonic clock issues.
	require.Equal(t, session.Start.Unix(), fetched.Start.Unix())
	require.Equal(t, session.End.Unix(), fetched.End.Unix())
	require.Equal(t, int32(1), fetched.Version)

	// Test getting a non-existent record.
	_, err = m.Get("f47ac10b-58cc-4372-a567-0e02b2c3d479")
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrRecordNotFound))
}

func TestSessionModel_Update(t *testing.T) {
	db, courseID := setupSessionsTestDB(t)
	m := SessionModel{DB: db}

	session := newTestSession(t, courseID)
	err := m.Insert(session)
	require.NoError(t, err)

	// Update some fields.
	session.Location = "Auditorium B"
	session.End = session.End.Add(time.Hour)

	err = m.Update(session)
	require.NoError(t, err)

	// Check version and updated_at.
	require.Equal(t, int32(2), session.Version)
	require.NotNil(t, session.UpdatedAt)
	require.WithinDuration(t, time.Now(), *session.UpdatedAt, time.Second)

	// Fetch to verify persistence.
	fetched, err := m.Get(session.ID)
	require.NoError(t, err)
	require.Equal(t, "Auditorium B", fetched.Location)
	require.Equal(t, session.End.Unix(), fetched.End.Unix())
	require.Equal(t, int32(2), fetched.Version)

	// Test for edit conflict.
	session.Version = 1
	err = m.Update(session)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrEditConflict))
}

func TestSessionModel_Delete(t *testing.T) {
	db, courseID := setupSessionsTestDB(t)
	m := SessionModel{DB: db}

	session := newTestSession(t, courseID)
	err := m.Insert(session)
	require.NoError(t, err)

	// Test successful deletion.
	err = m.Delete(session.ID)
	require.NoError(t, err)

	// Verify it's gone.
	_, err = m.Get(session.ID)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrRecordNotFound))

	// Test deleting a non-existent record.
	err = m.Delete("f47ac10b-58cc-4372-a567-0e02b2c3d479")
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrRecordNotFound))
}

func TestSessionModel_GetAll(t *testing.T) {
	db, courseID1 := setupSessionsTestDB(t)
	m := SessionModel{DB: db}

	// Insert a second course to test course_id filtering.
	var courseID2 string
	err := db.QueryRow(`INSERT INTO courses (title, category, default_credit_hours, created_by_user_id) SELECT 'Second Course', 'elective', 4, id FROM users WHERE email = 'sessionuser@example.com' RETURNING id`).Scan(&courseID2)
	require.NoError(t, err)

	// Insert records for testing.
	session1 := Session{CourseID: courseID1, Start: time.Now().Add(10 * time.Hour), End: time.Now().Add(12 * time.Hour), Location: "Main Hall"}
	session2 := Session{CourseID: courseID2, Start: time.Now().Add(20 * time.Hour), End: time.Now().Add(22 * time.Hour), Location: "Room 101"}
	session3 := Session{CourseID: courseID1, Start: time.Now().Add(30 * time.Hour), End: time.Now().Add(32 * time.Hour), Location: "Main Hall"}
	require.NoError(t, m.Insert(&session1))
	require.NoError(t, m.Insert(&session2))
	require.NoError(t, m.Insert(&session3))

	safelist := []string{"id", "start_datetime", "location_text", "-id", "-start_datetime", "-location_text"}
	filters := Filters{Page: 1, PageSize: 20, Sort: "id", SortSafelist: safelist}

	// Test case 1: Get all records.
	allSessions, metadata, err := m.GetAll("", "", filters)
	require.NoError(t, err)
	require.Len(t, allSessions, 3)
	require.Equal(t, int64(3), metadata.TotalRecords)

	// Test case 2: Filter by location.
	filtered, metadata, err := m.GetAll("Main Hall", "", filters)
	require.NoError(t, err)
	require.Len(t, filtered, 2)
	require.Equal(t, int64(2), metadata.TotalRecords)

	// Test case 3: Filter by course_id.
	filtered, metadata, err = m.GetAll("", courseID2, filters)
	require.NoError(t, err)
	require.Len(t, filtered, 1)
	require.Equal(t, "Room 101", filtered[0].Location)
	require.Equal(t, int64(1), metadata.TotalRecords)

	// Test case 4: Sorting.
	filters.Sort = "-start_datetime"
	sorted, _, err := m.GetAll("", "", filters)
	require.NoError(t, err)
	require.Len(t, sorted, 3)
	require.Equal(t, session3.ID, sorted[0].ID) // session3 is the latest, so it should be first.

	// Test case 5: Pagination.
	filters.Page = 2
	filters.PageSize = 2
	filters.Sort = "start_datetime"
	paginated, metadata, err := m.GetAll("", "", filters)
	require.NoError(t, err)
	require.Len(t, paginated, 1)
	require.Equal(t, session3.ID, paginated[0].ID) // Page 1: session1, session2. Page 2: session3
	require.Equal(t, int64(3), metadata.TotalRecords)
}

func TestValidateSession(t *testing.T) {
	v := validator.New()
	session := &Session{
		CourseID: "", // Invalid
		Start:    time.Time{}, // Invalid
		End:      time.Time{}, // Invalid (not after zero time)
		Location: "", // Invalid
	}

	ValidateSession(v, session)
	require.False(t, v.Valid())
	require.Contains(t, v.Errors, "course_id")
	require.Contains(t, v.Errors, "start_datetime")
	require.Contains(t, v.Errors, "end_datetime")
	require.Contains(t, v.Errors, "location_text")

	// Test end_datetime before start_datetime
	v = validator.New()
	session.Start = time.Now()
	session.End = time.Now().Add(-1 * time.Hour) // End is before start
	ValidateSession(v, session)
	require.Contains(t, v.Errors, "end_datetime")

	// Test valid session
	v = validator.New()
	validSession := newTestSession(t, "dummy-course-id")
	ValidateSession(v, validSession)
	require.True(t, v.Valid())
}