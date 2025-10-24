package data

import (
	"database/sql"
	"errors"
	"testing"
	//"time"

	"github.com/amari03/test1/internal/validator"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// setupFacilitatorsTestDB sets up a clean database environment for facilitator tests.
func setupFacilitatorsTestDB(t *testing.T) *sql.DB {
	// --- IMPORTANT ---
	// Update this DSN to point to your dedicated test database if needed.
	dsn := "postgres://test1_test:fishsticks@localhost/test1_test?sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)

	err = db.Ping()
	require.NoError(t, err, "failed to connect to the test database")

	// Enable the pgcrypto extension for UUID generation.
	_, err = db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`)
	require.NoError(t, err)

	// Create the facilitators table for our tests.
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS facilitators (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        first_name TEXT NOT NULL,
        last_name TEXT NOT NULL,
        notes TEXT,
        version INTEGER NOT NULL DEFAULT 1
    );`
	_, err = db.Exec(createTableSQL)
	require.NoError(t, err)

	// Register a cleanup function to drop the table after the test completes.
	t.Cleanup(func() {
		_, err := db.Exec("DROP TABLE IF EXISTS facilitators;")
		require.NoError(t, err)
		db.Close()
	})

	return db
}

// newTestFacilitator is a helper to create a valid facilitator instance for testing.
func newTestFacilitator(t *testing.T) *Facilitator {
	notes := "Expert in tactical training."
	return &Facilitator{
		FirstName: "Jane",
		LastName:  "Smith",
		Notes:     &notes,
	}
}

func TestFacilitatorModel_Insert(t *testing.T) {
	db := setupFacilitatorsTestDB(t)
	m := FacilitatorModel{DB: db}

	facilitator := newTestFacilitator(t)

	err := m.Insert(facilitator)
	require.NoError(t, err)

	// Check that the database populated the ID and Version fields.
	require.NotEmpty(t, facilitator.ID)
	require.Equal(t, int32(1), facilitator.Version)

	// Fetch the record back to double-check.
	fetched, err := m.Get(facilitator.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	require.Equal(t, "Jane", fetched.FirstName)
	require.Equal(t, "Smith", fetched.LastName)
	require.Equal(t, "Expert in tactical training.", *fetched.Notes)
}

func TestFacilitatorModel_Get(t *testing.T) {
	db := setupFacilitatorsTestDB(t)
	m := FacilitatorModel{DB: db}

	// First, insert a record to test Get.
	facilitator := newTestFacilitator(t)
	err := m.Insert(facilitator)
	require.NoError(t, err)

	// Test successful Get.
	fetched, err := m.Get(facilitator.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)

	// Verify the fields match.
	require.Equal(t, facilitator.ID, fetched.ID)
	require.Equal(t, facilitator.FirstName, fetched.FirstName)
	require.Equal(t, facilitator.LastName, fetched.LastName)
	require.Equal(t, *facilitator.Notes, *fetched.Notes)
	require.Equal(t, int32(1), fetched.Version)

	// Test getting a non-existent record.
	nonExistentID := "f47ac10b-58cc-4372-a567-0e02b2c3d479" // A random UUID
	_, err = m.Get(nonExistentID)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrRecordNotFound))
}

func TestFacilitatorModel_Update(t *testing.T) {
	db := setupFacilitatorsTestDB(t)
	m := FacilitatorModel{DB: db}

	// Insert a record to test Update.
	facilitator := newTestFacilitator(t)
	err := m.Insert(facilitator)
	require.NoError(t, err)

	// Update some fields.
	facilitator.FirstName = "John"
	facilitator.Notes = nil // Test updating to a NULL value

	err = m.Update(facilitator)
	require.NoError(t, err)

	// Check that the version incremented.
	require.Equal(t, int32(2), facilitator.Version)

	// Fetch the record again to verify the update persisted.
	fetched, err := m.Get(facilitator.ID)
	require.NoError(t, err)
	require.Equal(t, "John", fetched.FirstName)
	require.Nil(t, fetched.Notes)
	require.Equal(t, int32(2), fetched.Version)

	// Test for edit conflict (optimistic locking).
	// Try to update again with the old version number (version 1).
	facilitator.Version = 1
	err = m.Update(facilitator)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrEditConflict))
}

func TestFacilitatorModel_Delete(t *testing.T) {
	db := setupFacilitatorsTestDB(t)
	m := FacilitatorModel{DB: db}

	// Insert a record to test Delete.
	facilitator := newTestFacilitator(t)
	err := m.Insert(facilitator)
	require.NoError(t, err)

	// Test successful deletion.
	err = m.Delete(facilitator.ID)
	require.NoError(t, err)

	// Verify it's gone.
	_, err = m.Get(facilitator.ID)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrRecordNotFound))

	// Test deleting a non-existent record.
	nonExistentID := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	err = m.Delete(nonExistentID)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrRecordNotFound))
}

func TestFacilitatorModel_GetAll(t *testing.T) {
	db := setupFacilitatorsTestDB(t)
	m := FacilitatorModel{DB: db}

	// Insert a few records for testing.
	fac1 := Facilitator{FirstName: "Alice", LastName: "Williams"}
	fac2 := Facilitator{FirstName: "Bob", LastName: "Johnson"}
	fac3 := Facilitator{FirstName: "Charlie", LastName: "Williams"}
	require.NoError(t, m.Insert(&fac1))
	require.NoError(t, m.Insert(&fac2))
	require.NoError(t, m.Insert(&fac3))

	// Define a standard filter safelist.
	safelist := []string{"id", "first_name", "last_name", "-id", "-first_name", "-last_name"}

	// Test case 1: Get all records with default pagination.
	filters := Filters{Page: 1, PageSize: 20, Sort: "id", SortSafelist: safelist}
	allFacilitators, metadata, err := m.GetAll("", "", filters)
	require.NoError(t, err)
	require.Len(t, allFacilitators, 3)
	require.Equal(t, int64(3), metadata.TotalRecords)

	// Test case 2: Filter by first name.
	filtered, metadata, err := m.GetAll("Alice", "", filters)
	require.NoError(t, err)
	require.Len(t, filtered, 1)
	require.Equal(t, "Alice", filtered[0].FirstName)
	require.Equal(t, int64(1), metadata.TotalRecords)

	// Test case 3: Filter by last name.
	filtered, metadata, err = m.GetAll("", "Williams", filters)
	require.NoError(t, err)
	require.Len(t, filtered, 2)
	require.Equal(t, int64(2), metadata.TotalRecords)

	// Test case 4: Sorting (descending by first name).
	filters.Sort = "-first_name"
	sorted, _, err := m.GetAll("", "", filters)
	require.NoError(t, err)
	require.Len(t, sorted, 3)
	require.Equal(t, "Charlie", sorted[0].FirstName) // Charlie, Bob, Alice
	require.Equal(t, "Bob", sorted[1].FirstName)
	require.Equal(t, "Alice", sorted[2].FirstName)

	// Test case 5: Pagination.
	filters.Page = 2
	filters.PageSize = 2
	filters.Sort = "first_name" // Sort ASC for predictable pagination
	paginated, metadata, err := m.GetAll("", "", filters)
	require.NoError(t, err)
	require.Len(t, paginated, 1)
	require.Equal(t, "Charlie", paginated[0].FirstName) // Page 1: Alice, Bob. Page 2: Charlie
	require.Equal(t, int64(3), metadata.TotalRecords)
	require.Equal(t, 2, metadata.CurrentPage)
	require.Equal(t, 2, metadata.LastPage)
}

func TestValidateFacilitator(t *testing.T) {
	v := validator.New()
	facilitator := &Facilitator{
		FirstName: "", // Invalid
		LastName:  "", // Invalid
	}

	ValidateFacilitator(v, facilitator)
	require.False(t, v.Valid())
	require.Contains(t, v.Errors, "first_name")
	require.Contains(t, v.Errors, "last_name")

	// Test valid facilitator
	v = validator.New()
	validFacilitator := newTestFacilitator(t)
	ValidateFacilitator(v, validFacilitator)
	require.True(t, v.Valid())
}