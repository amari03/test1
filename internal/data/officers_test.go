package data

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/amari03/test1/internal/validator" // Make sure to use your actual module path
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// NOTE: The code you provided uses custom errors like ErrRecordNotFound and ErrEditConflict.
// Ensure these are defined in your data package, for example, in an `errors.go` file:
//
// var (
//     ErrRecordNotFound = errors.New("record not found")
//     ErrEditConflict   = errors.New("edit conflict")
// )

// setupTestDB connects to a test database, creates the necessary schema, and returns the connection.
// It also registers a cleanup function to tear down the schema after the test finishes.
func setupTestDB(t *testing.T) *sql.DB {
	// --- IMPORTANT ---
	// Update this DSN to point to your dedicated test database.
	dsn := "postgres://test1_test:fishsticks@localhost/test1_test?sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)

	err = db.Ping()
	require.NoError(t, err, "failed to connect to the test database")

	// Enable the pgcrypto extension for UUID generation if you use it.
	_, err = db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`)
	require.NoError(t, err)

	// Create the officers table for our tests.
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS officers (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        regulation_number TEXT,
        first_name TEXT NOT NULL,
        last_name TEXT NOT NULL,
        sex TEXT NOT NULL,
        rank_code TEXT NOT NULL,
        region_id UUID,
        formation_id UUID,
        posting_id UUID,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMPTZ,
        archived_at TIMESTAMPTZ,
        version INTEGER NOT NULL DEFAULT 1
    );`
	_, err = db.Exec(createTableSQL)
	require.NoError(t, err)

	// Register a cleanup function to drop the table after the test completes.
	t.Cleanup(func() {
		_, err := db.Exec("DROP TABLE IF EXISTS officers;")
		require.NoError(t, err)
		db.Close()
	})

	return db
}

// newTestOfficer is a helper to create a valid officer instance for testing.
func newTestOfficer(t *testing.T) *Officer {
	regNum := "12345"
	return &Officer{
		RegulationNumber: &regNum,
		FirstName:        "John",
		LastName:         "Doe",
		Sex:              "male",
		RankCode:         "CONSTABLE",
	}
}

func TestOfficerModel_Insert(t *testing.T) {
	db := setupTestDB(t)
	m := OfficerModel{DB: db}

	officer := newTestOfficer(t)

	err := m.Insert(officer)
	require.NoError(t, err)

	// Check that the database populated the ID and CreatedAt fields.
	require.NotEmpty(t, officer.ID)
	require.WithinDuration(t, time.Now(), officer.CreatedAt, time.Second)

	// Fetch the record back to double-check.
	fetchedOfficer, err := m.Get(officer.ID)
	require.NoError(t, err)
	require.NotNil(t, fetchedOfficer)
	require.Equal(t, officer.FirstName, fetchedOfficer.FirstName)
	require.Equal(t, officer.LastName, fetchedOfficer.LastName)
}

func TestOfficerModel_Get(t *testing.T) {
	db := setupTestDB(t)
	m := OfficerModel{DB: db}

	// First, insert a record to test Get.
	officer := newTestOfficer(t)
	err := m.Insert(officer)
	require.NoError(t, err)

	// Test successful Get.
	fetchedOfficer, err := m.Get(officer.ID)
	require.NoError(t, err)
	require.NotNil(t, fetchedOfficer)

	// Verify the fields match.
	require.Equal(t, officer.ID, fetchedOfficer.ID)
	require.Equal(t, *officer.RegulationNumber, *fetchedOfficer.RegulationNumber)
	require.Equal(t, officer.FirstName, fetchedOfficer.FirstName)
	require.Equal(t, officer.LastName, fetchedOfficer.LastName)
	require.Equal(t, officer.Sex, fetchedOfficer.Sex)
	require.Equal(t, officer.RankCode, fetchedOfficer.RankCode)
	require.WithinDuration(t, officer.CreatedAt, fetchedOfficer.CreatedAt, time.Second)
	require.Equal(t, int32(1), fetchedOfficer.Version)

	// Test getting a non-existent record.
	nonExistentID := "f47ac10b-58cc-4372-a567-0e02b2c3d479" // A random UUID
	_, err = m.Get(nonExistentID)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrRecordNotFound))
}

func TestOfficerModel_Update(t *testing.T) {
	db := setupTestDB(t)
	m := OfficerModel{DB: db}

	// Insert a record to test Update.
	officer := newTestOfficer(t)
	err := m.Insert(officer)
	require.NoError(t, err)

	// Update some fields.
	officer.FirstName = "Jane"
	officer.RankCode = "SERGEANT"
	newRegNum := "67890"
	officer.RegulationNumber = &newRegNum

	err = m.Update(officer)
	require.NoError(t, err)

	// Check that the version incremented and UpdatedAt is set.
	require.Equal(t, int32(2), officer.Version)
	require.NotNil(t, officer.UpdatedAt)
	require.WithinDuration(t, time.Now(), *officer.UpdatedAt, time.Second)

	// Fetch the record again to verify the update persisted.
	fetchedOfficer, err := m.Get(officer.ID)
	require.NoError(t, err)
	require.Equal(t, "Jane", fetchedOfficer.FirstName)
	require.Equal(t, "SERGEANT", fetchedOfficer.RankCode)
	require.Equal(t, "67890", *fetchedOfficer.RegulationNumber)
	require.Equal(t, int32(2), fetchedOfficer.Version)

	// Test for edit conflict (optimistic locking).
	// Try to update again with the old version number (version 1).
	officer.Version = 1
	err = m.Update(officer)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrEditConflict))
}

func TestOfficerModel_Delete(t *testing.T) {
	db := setupTestDB(t)
	m := OfficerModel{DB: db}

	// Insert a record to test Delete.
	officer := newTestOfficer(t)
	err := m.Insert(officer)
	require.NoError(t, err)

	// Test successful deletion.
	err = m.Delete(officer.ID)
	require.NoError(t, err)

	// Verify it's gone.
	_, err = m.Get(officer.ID)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrRecordNotFound))

	// Test deleting a non-existent record.
	nonExistentID := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	err = m.Delete(nonExistentID)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrRecordNotFound))
}

func TestOfficerModel_GetAll(t *testing.T) {
	db := setupTestDB(t)
	m := OfficerModel{DB: db}

	// Insert a few records for testing.
	officer1 := Officer{FirstName: "Alice", LastName: "Smith", Sex: "female", RankCode: "CONSTABLE"}
	officer2 := Officer{FirstName: "Bob", LastName: "Jones", Sex: "male", RankCode: "SERGEANT"}
	officer3 := Officer{FirstName: "Charlie", LastName: "Smith", Sex: "male", RankCode: "CONSTABLE"}
	require.NoError(t, m.Insert(&officer1))
	require.NoError(t, m.Insert(&officer2))
	require.NoError(t, m.Insert(&officer3))

	// Define a standard filter safelist.
	safelist := []string{"id", "first_name", "last_name", "rank_code", "-id", "-first_name", "-last_name", "-rank_code"}

	// Test case 1: Get all records with default pagination.
	filters := Filters{Page: 1, PageSize: 20, Sort: "id", SortSafelist: safelist}
	allOfficers, metadata, err := m.GetAll("", "", "", filters)
	require.NoError(t, err)
	require.Len(t, allOfficers, 3)
	require.Equal(t, int64(3), metadata.TotalRecords)

	// Test case 2: Filter by first_name.
	filteredOfficers, metadata, err := m.GetAll("Alice", "", "", filters)
	require.NoError(t, err)
	require.Len(t, filteredOfficers, 1)
	require.Equal(t, "Alice", filteredOfficers[0].FirstName)
	require.Equal(t, int64(1), metadata.TotalRecords)

	// Test case 3: Filter by last_name.
	filteredOfficers, metadata, err = m.GetAll("", "Smith", "", filters)
	require.NoError(t, err)
	require.Len(t, filteredOfficers, 2)
	require.Equal(t, int64(2), metadata.TotalRecords)

	// Test case 4: Filter by rank_code.
	filteredOfficers, metadata, err = m.GetAll("", "", "SERGEANT", filters)
	require.NoError(t, err)
	require.Len(t, filteredOfficers, 1)
	require.Equal(t, "Bob", filteredOfficers[0].FirstName)
	require.Equal(t, int64(1), metadata.TotalRecords)

	// Test case 5: Sorting (descending by first_name).
	filters.Sort = "-first_name"
	sortedOfficers, _, err := m.GetAll("", "", "", filters)
	require.NoError(t, err)
	require.Len(t, sortedOfficers, 3)
	require.Equal(t, "Charlie", sortedOfficers[0].FirstName) // Charlie, Bob, Alice
	require.Equal(t, "Bob", sortedOfficers[1].FirstName)
	require.Equal(t, "Alice", sortedOfficers[2].FirstName)

	// Test case 6: Pagination.
	filters.Page = 2
	filters.PageSize = 2
	filters.Sort = "first_name" // Sort ASC for predictable pagination
	paginatedOfficers, metadata, err := m.GetAll("", "", "", filters)
	require.NoError(t, err)
	require.Len(t, paginatedOfficers, 1)
	require.Equal(t, "Charlie", paginatedOfficers[0].FirstName) // Page 1: Alice, Bob. Page 2: Charlie
	require.Equal(t, int64(3), metadata.TotalRecords)
	require.Equal(t, 2, metadata.CurrentPage)
	require.Equal(t, 2, metadata.LastPage)
}

func TestValidateOfficer(t *testing.T) {
	v := validator.New()
	officer := Officer{
		FirstName: "", // Invalid
		LastName:  "", // Invalid
		Sex:       "other", // Invalid
		RankCode:  "", // Invalid
	}

	ValidateOfficer(v, &officer)

	require.False(t, v.Valid())
	require.Contains(t, v.Errors, "first_name")
	require.Contains(t, v.Errors, "last_name")
	require.Contains(t, v.Errors, "sex")
	require.Contains(t, v.Errors, "rank_code")

	// Test valid officer
	v = validator.New()
	validOfficer := newTestOfficer(t)
	ValidateOfficer(v, validOfficer)
	require.True(t, v.Valid())
}