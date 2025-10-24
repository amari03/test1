package data

import (
	"database/sql"
	//"errors"
	"testing"
	"time"

	"github.com/amari03/test1/internal/validator" 
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// setupAttendanceTestDB sets up a DB with all necessary dependencies for an attendance record.
// It returns the DB connection, a dummy officer ID, and a dummy session ID.
func setupAttendanceTestDB(t *testing.T) (*sql.DB, string, string) {
	dsn := "postgres://test1_test:fishsticks@localhost/test1_test?sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	err = db.Ping()
	require.NoError(t, err, "failed to connect to the test database")

	_, err = db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`)
	require.NoError(t, err)

	// Create dependency tables in order
	db.Exec(`CREATE TABLE IF NOT EXISTS users (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), email TEXT UNIQUE NOT NULL);`)
	db.Exec(`CREATE TABLE IF NOT EXISTS courses (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), title TEXT NOT NULL, category TEXT NOT NULL, default_credit_hours NUMERIC NOT NULL, created_by_user_id UUID NOT NULL REFERENCES users(id), version INTEGER NOT NULL DEFAULT 1);`)
	db.Exec(`CREATE TABLE IF NOT EXISTS sessions (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), course_id UUID NOT NULL REFERENCES courses(id), start_datetime TIMESTAMPTZ NOT NULL, end_datetime TIMESTAMPTZ NOT NULL, location_text TEXT NOT NULL, version INTEGER NOT NULL DEFAULT 1);`)
	db.Exec(`CREATE TABLE IF NOT EXISTS officers (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), first_name TEXT NOT NULL, last_name TEXT NOT NULL, sex TEXT NOT NULL, rank_code TEXT NOT NULL, version INTEGER NOT NULL DEFAULT 1);`)

	// Create the attendance table
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS attendance (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        officer_id UUID NOT NULL REFERENCES officers(id),
        session_id UUID NOT NULL REFERENCES sessions(id),
        status TEXT NOT NULL,
        credited_hours NUMERIC NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        version INTEGER NOT NULL DEFAULT 1,
        UNIQUE(officer_id, session_id)
    );`
	_, err = db.Exec(createTableSQL)
	require.NoError(t, err)

	// Insert dummy data to get valid IDs
	var userID, courseID, sessionID, officerID string
	err = db.QueryRow(`INSERT INTO users (email) VALUES ('attendanceuser@example.com') ON CONFLICT (email) DO NOTHING RETURNING id;`).Scan(&userID)
	if err == sql.ErrNoRows {
		err = db.QueryRow(`SELECT id FROM users WHERE email = 'attendanceuser@example.com'`).Scan(&userID)
	}
	require.NoError(t, err)
	err = db.QueryRow(`INSERT INTO courses (title, category, default_credit_hours, created_by_user_id) VALUES ('Attendance Test Course', 'mandatory', 8, $1) RETURNING id`, userID).Scan(&courseID)
	require.NoError(t, err)
	err = db.QueryRow(`INSERT INTO sessions (course_id, start_datetime, end_datetime, location_text) VALUES ($1, $2, $3, 'Room 1') RETURNING id`, courseID, time.Now(), time.Now().Add(time.Hour)).Scan(&sessionID)
	require.NoError(t, err)
	err = db.QueryRow(`INSERT INTO officers (first_name, last_name, sex, rank_code) VALUES ('John', 'Doe', 'male', 'CONSTABLE') RETURNING id`).Scan(&officerID)
	require.NoError(t, err)

	// Cleanup in reverse order
	t.Cleanup(func() {
		db.Exec("DROP TABLE IF EXISTS attendance;")
		db.Exec("DROP TABLE IF EXISTS officers;")
		db.Exec("DROP TABLE IF EXISTS sessions;")
		db.Exec("DROP TABLE IF EXISTS courses;")
		db.Exec("DROP TABLE IF EXISTS users;")
		db.Close()
	})

	return db, officerID, sessionID
}

// newTestAttendance is a helper to create a valid attendance instance.
func newTestAttendance(t *testing.T, officerID, sessionID string) *Attendance {
	return &Attendance{
		OfficerID:     officerID,
		SessionID:     sessionID,
		Status:        "attended",
		CreditedHours: 8.0,
	}
}

func TestAttendanceModel_Insert(t *testing.T) {
	db, officerID, sessionID := setupAttendanceTestDB(t)
	m := AttendanceModel{DB: db}

	attendance := newTestAttendance(t, officerID, sessionID)

	err := m.Insert(attendance)
	require.NoError(t, err)

	// Check populated fields
	require.NotEmpty(t, attendance.ID)
	require.WithinDuration(t, time.Now(), attendance.CreatedAt, time.Second)
	require.Equal(t, int32(1), attendance.Version)

	// Fetch to double-check
	fetched, err := m.Get(attendance.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	require.Equal(t, attendance.OfficerID, fetched.OfficerID)
	require.Equal(t, attendance.Status, fetched.Status)
}

func TestAttendanceModel_Get(t *testing.T) {
	db, officerID, sessionID := setupAttendanceTestDB(t)
	m := AttendanceModel{DB: db}

	attendance := newTestAttendance(t, officerID, sessionID)
	err := m.Insert(attendance)
	require.NoError(t, err)

	// Test successful Get
	fetched, err := m.Get(attendance.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)

	// Verify fields
	require.Equal(t, attendance.ID, fetched.ID)
	require.Equal(t, attendance.OfficerID, fetched.OfficerID)
	require.Equal(t, attendance.SessionID, fetched.SessionID)
	require.Equal(t, attendance.CreditedHours, fetched.CreditedHours)
	require.Equal(t, int32(1), fetched.Version)

	// Test not found
	_, err = m.Get("f47ac10b-58cc-4372-a567-0e02b2c3d479")
	require.ErrorIs(t, err, ErrRecordNotFound)
}

func TestAttendanceModel_Update(t *testing.T) {
	db, officerID, sessionID := setupAttendanceTestDB(t)
	m := AttendanceModel{DB: db}

	attendance := newTestAttendance(t, officerID, sessionID)
	err := m.Insert(attendance)
	require.NoError(t, err)

	// Update fields
	attendance.Status = "excused"
	attendance.CreditedHours = 0

	err = m.Update(attendance)
	require.NoError(t, err)
	require.Equal(t, int32(2), attendance.Version)

	// Fetch to verify persistence
	fetched, err := m.Get(attendance.ID)
	require.NoError(t, err)
	require.Equal(t, "excused", fetched.Status)
	require.Equal(t, float64(0), fetched.CreditedHours)
	require.Equal(t, int32(2), fetched.Version)

	// Test edit conflict
	attendance.Version = 1
	err = m.Update(attendance)
	require.ErrorIs(t, err, ErrEditConflict)
}

func TestAttendanceModel_Delete(t *testing.T) {
	db, officerID, sessionID := setupAttendanceTestDB(t)
	m := AttendanceModel{DB: db}

	attendance := newTestAttendance(t, officerID, sessionID)
	err := m.Insert(attendance)
	require.NoError(t, err)

	// Test successful delete
	err = m.Delete(attendance.ID)
	require.NoError(t, err)

	// Verify it's gone
	_, err = m.Get(attendance.ID)
	require.ErrorIs(t, err, ErrRecordNotFound)

	// Test deleting non-existent record
	err = m.Delete("f47ac10b-58cc-4372-a567-0e02b2c3d479")
	require.ErrorIs(t, err, ErrRecordNotFound)
}

func TestAttendanceModel_GetAll(t *testing.T) {
	db, officerID1, sessionID1 := setupAttendanceTestDB(t)
	m := AttendanceModel{DB: db}

	// Create a second officer and session for filtering tests
	var officerID2, sessionID2 string
	err := db.QueryRow(`INSERT INTO officers (first_name, last_name, sex, rank_code) VALUES ('Jane', 'Smith', 'female', 'SERGEANT') RETURNING id`).Scan(&officerID2)
	require.NoError(t, err)
	err = db.QueryRow(`INSERT INTO sessions (course_id, start_datetime, end_datetime, location_text) SELECT course_id, $1, $2, 'Room 2' FROM sessions LIMIT 1 RETURNING id`, time.Now().Add(2*time.Hour), time.Now().Add(3*time.Hour)).Scan(&sessionID2)
	require.NoError(t, err)

	// Insert test records
	att1 := Attendance{OfficerID: officerID1, SessionID: sessionID1, Status: "attended", CreditedHours: 8}
	att2 := Attendance{OfficerID: officerID2, SessionID: sessionID1, Status: "attended", CreditedHours: 8}
	att3 := Attendance{OfficerID: officerID1, SessionID: sessionID2, Status: "absent", CreditedHours: 0}
	require.NoError(t, m.Insert(&att1))
	require.NoError(t, m.Insert(&att2))
	require.NoError(t, m.Insert(&att3))

	safelist := []string{"id", "status", "-id", "-status"}
	filters := Filters{Page: 1, PageSize: 20, Sort: "id", SortSafelist: safelist}

	// Test case 1: Get all
	all, metadata, err := m.GetAll("", "", filters)
	require.NoError(t, err)
	require.Len(t, all, 3)
	require.Equal(t, int64(3), metadata.TotalRecords)

	// Test case 2: Filter by officer_id
	filtered, metadata, err := m.GetAll(officerID1, "", filters)
	require.NoError(t, err)
	require.Len(t, filtered, 2)
	require.Equal(t, int64(2), metadata.TotalRecords)

	// Test case 3: Filter by session_id
	filtered, metadata, err = m.GetAll("", sessionID1, filters)
	require.NoError(t, err)
	require.Len(t, filtered, 2)
	require.Equal(t, int64(2), metadata.TotalRecords)
}

func TestValidateAttendance(t *testing.T) {
	v := validator.New()
	att := &Attendance{
		OfficerID:     "",          // Invalid
		SessionID:     "",          // Invalid
		Status:        "maybe",     // Invalid
		CreditedHours: -1,          // Invalid
	}
	ValidateAttendance(v, att)
	require.False(t, v.Valid())
	require.Contains(t, v.Errors, "officer_id")
	require.Contains(t, v.Errors, "session_id")
	require.Contains(t, v.Errors, "status")
	require.Contains(t, v.Errors, "credited_hours")

	// Test valid
	v = validator.New()
	validAtt := newTestAttendance(t, "dummy-officer-id", "dummy-session-id")
	ValidateAttendance(v, validAtt)
	require.True(t, v.Valid())
}