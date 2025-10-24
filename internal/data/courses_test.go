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

// setupCoursesTestDB sets up a clean database environment for course tests.
// It creates a minimal users table to satisfy the foreign key constraint.
func setupCoursesTestDB(t *testing.T) (*sql.DB, string) {
	// --- IMPORTANT ---
	// Update this DSN to point to your dedicated test database.
	dsn := "postgres://test1_test:fishsticks@localhost/test1_test?sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)

	err = db.Ping()
	require.NoError(t, err, "failed to connect to the test database")

	// Enable the pgcrypto extension for UUID generation.
	_, err = db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`)
	require.NoError(t, err)

	// Create a minimal users table to satisfy the created_by_user_id foreign key.
	createUsersTableSQL := `
    CREATE TABLE IF NOT EXISTS users (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        email TEXT UNIQUE NOT NULL
    );`
	_, err = db.Exec(createUsersTableSQL)
	require.NoError(t, err)

	// Create the courses table.
	createCoursesTableSQL := `
    CREATE TABLE IF NOT EXISTS courses (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        title TEXT NOT NULL,
        category TEXT NOT NULL,
        default_credit_hours NUMERIC NOT NULL,
        description TEXT,
        created_by_user_id UUID NOT NULL REFERENCES users(id),
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMPTZ,
        version INTEGER NOT NULL DEFAULT 1
    );`
	_, err = db.Exec(createCoursesTableSQL)
	require.NoError(t, err)

	// Insert a dummy user to associate courses with.
	var dummyUserID string
	err = db.QueryRow(`INSERT INTO users (email) VALUES ('testuser@example.com') ON CONFLICT (email) DO NOTHING RETURNING id;`).Scan(&dummyUserID)
	if err == sql.ErrNoRows {
		// If user already exists, just grab their ID
		err = db.QueryRow(`SELECT id FROM users WHERE email = 'testuser@example.com'`).Scan(&dummyUserID)
	}
	require.NoError(t, err)

	// Register a cleanup function to drop the tables after the test completes.
	t.Cleanup(func() {
		_, err := db.Exec("DROP TABLE IF EXISTS courses;")
		require.NoError(t, err)
		_, err = db.Exec("DROP TABLE IF EXISTS users;")
		require.NoError(t, err)
		db.Close()
	})

	return db, dummyUserID
}

// newTestCourse is a helper to create a valid course instance for testing.
func newTestCourse(t *testing.T, userID string) *Course {
	return &Course{
		Title:              "Defensive Driving",
		Category:           "mandatory",
		DefaultCreditHours: 8.5,
		Description:        "A course on safe driving techniques.",
		CreatedByUserID:    userID,
	}
}

func TestCourseModel_Insert(t *testing.T) {
	db, userID := setupCoursesTestDB(t)
	m := CourseModel{DB: db}

	course := newTestCourse(t, userID)

	err := m.Insert(course)
	require.NoError(t, err)

	// Check that the database populated the ID, CreatedAt, and Version fields.
	require.NotEmpty(t, course.ID)
	require.WithinDuration(t, time.Now(), course.CreatedAt, time.Second)
	require.Equal(t, int32(1), course.Version)

	// Fetch the record back to double-check.
	fetchedCourse, err := m.Get(course.ID)
	require.NoError(t, err)
	require.NotNil(t, fetchedCourse)
	require.Equal(t, course.Title, fetchedCourse.Title)
	require.Equal(t, course.CreatedByUserID, fetchedCourse.CreatedByUserID)
}

func TestCourseModel_Get(t *testing.T) {
	db, userID := setupCoursesTestDB(t)
	m := CourseModel{DB: db}

	// First, insert a record to test Get.
	course := newTestCourse(t, userID)
	err := m.Insert(course)
	require.NoError(t, err)

	// Test successful Get.
	fetchedCourse, err := m.Get(course.ID)
	require.NoError(t, err)
	require.NotNil(t, fetchedCourse)

	// Verify the fields match.
	require.Equal(t, course.ID, fetchedCourse.ID)
	require.Equal(t, course.Title, fetchedCourse.Title)
	require.Equal(t, course.Category, fetchedCourse.Category)
	require.Equal(t, course.DefaultCreditHours, fetchedCourse.DefaultCreditHours)
	require.WithinDuration(t, course.CreatedAt, fetchedCourse.CreatedAt, time.Second)
	require.Equal(t, int32(1), fetchedCourse.Version)

	// Test getting a non-existent record.
	nonExistentID := "f47ac10b-58cc-4372-a567-0e02b2c3d479" // A random UUID
	_, err = m.Get(nonExistentID)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrRecordNotFound))
}

func TestCourseModel_Update(t *testing.T) {
	db, userID := setupCoursesTestDB(t)
	m := CourseModel{DB: db}

	// Insert a record to test Update.
	course := newTestCourse(t, userID)
	err := m.Insert(course)
	require.NoError(t, err)

	// Update some fields.
	course.Title = "Advanced Defensive Driving"
	course.Category = "elective"

	err = m.Update(course)
	require.NoError(t, err)

	// Check that the version incremented and UpdatedAt is set.
	require.Equal(t, int32(2), course.Version)
	require.NotNil(t, course.UpdatedAt)
	require.WithinDuration(t, time.Now(), *course.UpdatedAt, time.Second)

	// Fetch the record again to verify the update persisted.
	fetchedCourse, err := m.Get(course.ID)
	require.NoError(t, err)
	require.Equal(t, "Advanced Defensive Driving", fetchedCourse.Title)
	require.Equal(t, "elective", fetchedCourse.Category)
	require.Equal(t, int32(2), fetchedCourse.Version)

	// Test for edit conflict (optimistic locking).
	// Try to update again with the old version number (version 1).
	course.Version = 1
	err = m.Update(course)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrEditConflict))
}

func TestCourseModel_Delete(t *testing.T) {
	db, userID := setupCoursesTestDB(t)
	m := CourseModel{DB: db}

	// Insert a record to test Delete.
	course := newTestCourse(t, userID)
	err := m.Insert(course)
	require.NoError(t, err)

	// Test successful deletion.
	err = m.Delete(course.ID)
	require.NoError(t, err)

	// Verify it's gone.
	_, err = m.Get(course.ID)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrRecordNotFound))

	// Test deleting a non-existent record.
	nonExistentID := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	err = m.Delete(nonExistentID)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrRecordNotFound))
}

func TestCourseModel_GetAll(t *testing.T) {
	db, userID := setupCoursesTestDB(t)
	m := CourseModel{DB: db}

	// Insert a few records for testing.
	course1 := Course{Title: "Firearms Safety", Category: "mandatory", DefaultCreditHours: 16, CreatedByUserID: userID}
	course2 := Course{Title: "Community Policing", Category: "elective", DefaultCreditHours: 8, CreatedByUserID: userID}
	course3 := Course{Title: "Advanced First Aid", Category: "mandatory", DefaultCreditHours: 24, CreatedByUserID: userID}
	require.NoError(t, m.Insert(&course1))
	require.NoError(t, m.Insert(&course2))
	require.NoError(t, m.Insert(&course3))

	// Define a standard filter safelist.
	safelist := []string{"id", "title", "category", "-id", "-title", "-category"}

	// Test case 1: Get all records with default pagination.
	filters := Filters{Page: 1, PageSize: 20, Sort: "id", SortSafelist: safelist}
	allCourses, metadata, err := m.GetAll("", "", filters)
	require.NoError(t, err)
	require.Len(t, allCourses, 3)
	require.Equal(t, int64(3), metadata.TotalRecords)

	// Test case 2: Filter by title.
	filteredCourses, metadata, err := m.GetAll("Firearms", "", filters)
	require.NoError(t, err)
	require.Len(t, filteredCourses, 1)
	require.Equal(t, "Firearms Safety", filteredCourses[0].Title)
	require.Equal(t, int64(1), metadata.TotalRecords)

	// Test case 3: Filter by category.
	filteredCourses, metadata, err = m.GetAll("", "mandatory", filters)
	require.NoError(t, err)
	require.Len(t, filteredCourses, 2)
	require.Equal(t, int64(2), metadata.TotalRecords)

	// Test case 4: Sorting (descending by title).
	filters.Sort = "-title"
	sortedCourses, _, err := m.GetAll("", "", filters)
	require.NoError(t, err)
	require.Len(t, sortedCourses, 3)
	require.Equal(t, "Firearms Safety", sortedCourses[0].Title)
	require.Equal(t, "Community Policing", sortedCourses[1].Title)
	require.Equal(t, "Advanced First Aid", sortedCourses[2].Title)

	// Test case 5: Pagination.
	filters.Page = 2
	filters.PageSize = 2
	filters.Sort = "title" // Sort ASC for predictable pagination
	paginatedCourses, metadata, err := m.GetAll("", "", filters)
	require.NoError(t, err)
	require.Len(t, paginatedCourses, 1)
	require.Equal(t, "Firearms Safety", paginatedCourses[0].Title) // Page 1: Advanced, Community. Page 2: Firearms
	require.Equal(t, int64(3), metadata.TotalRecords)
	require.Equal(t, 2, metadata.CurrentPage)
	require.Equal(t, 2, metadata.LastPage)
}

func TestValidateCourse(t *testing.T) {
	v := validator.New()
	course := &Course{
		Title:              "",        // Invalid
		Category:           "other",   // Invalid
		DefaultCreditHours: 0,         // Invalid
	}

	ValidateCourse(v, course)
	require.False(t, v.Valid())
	require.Contains(t, v.Errors, "title")
	require.Contains(t, v.Errors, "category")
	require.Contains(t, v.Errors, "default_credit_hours")

	// Test valid course
	v = validator.New()
	validCourse := newTestCourse(t, "dummy-user-id")
	ValidateCourse(v, validCourse)
	require.True(t, v.Valid())
}