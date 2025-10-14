package data

import (
	"database/sql"
	"time"
	"context"
	"errors"
	"fmt"

	"github.com/amari03/test1/internal/validator"
)

type Officer struct {
	ID               string     `json:"id"`
	RegulationNumber *string    `json:"regulation_number,omitempty"`
	FirstName        string     `json:"first_name"`
	LastName         string     `json:"last_name"`
	Sex              string     `json:"sex"`
	RankCode         string     `json:"rank_code"`
	RegionID         *string    `json:"region_id,omitempty"`
	FormationID      *string    `json:"formation_id,omitempty"`
	PostingID        *string    `json:"posting_id,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        *time.Time `json:"updated_at,omitempty"`
	ArchivedAt       *time.Time `json:"archived_at,omitempty"`
	Version          int32      `json:"version"`
}

type OfficerModel struct {
	DB *sql.DB
}

func ValidateOfficer(v *validator.Validator, officer *Officer) {
	v.Check(officer.FirstName != "", "first_name", "must be provided")
	v.Check(len(officer.FirstName) <= 100, "first_name", "must not exceed 100 bytes")
	v.Check(officer.LastName != "", "last_name", "must be provided")
	v.Check(len(officer.LastName) <= 100, "last_name", "must not exceed 100 bytes")
	v.Check(officer.Sex != "", "sex", "must be provided")
	v.Check(validator.In(officer.Sex, "male", "female", "unknown"), "sex", "must be male, female, or unknown")
	v.Check(officer.RankCode != "", "rank_code", "must be provided")
}

// Insert a new officer record into the database.
func (m OfficerModel) Insert(officer *Officer) error {
	query := `
        INSERT INTO officers (regulation_number, first_name, last_name, sex, rank_code)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, created_at`

	args := []interface{}{officer.RegulationNumber, &officer.FirstName, &officer.LastName, &officer.Sex, &officer.RankCode}

	return m.DB.QueryRow(query, args...).Scan(&officer.ID, &officer.CreatedAt)
}

// Get a specific officer by ID.
func (m OfficerModel) Get(id string) (*Officer, error) {
	query := `
        SELECT id, regulation_number, first_name, last_name, sex, rank_code,
               created_at, updated_at, version
        FROM officers
        WHERE id = $1`

	var officer Officer
	
    // Use a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&officer.ID,
		&officer.RegulationNumber,
		&officer.FirstName,
		&officer.LastName,
		&officer.Sex,
		&officer.RankCode,
		&officer.CreatedAt,
		&officer.UpdatedAt,
		&officer.Version, // Scan the version
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound // Use our custom error
		default:
			return nil, err
		}
	}
	return &officer, nil
}

// Update a specific officer record.
func (m OfficerModel) Update(officer *Officer) error {
	query := `
        UPDATE officers
        SET regulation_number = $1, first_name = $2, last_name = $3, sex = $4, rank_code = $5,
            updated_at = NOW(), version = version + 1
        WHERE id = $6 AND version = $7
        RETURNING updated_at, version`

	args := []interface{}{
		officer.RegulationNumber,
		officer.FirstName,
		officer.LastName,
		officer.Sex,
		officer.RankCode,
		officer.ID,
		officer.Version, // Add the version for optimistic locking
	}
	
    // Use a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Use QueryRowContext and check for ErrNoRows to detect edit conflicts.
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&officer.UpdatedAt, &officer.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict // Return our custom edit conflict error
		default:
			return err
		}
	}
	return nil
}

// Delete a specific officer record from the database.
func (m OfficerModel) Delete(id string) error {
	// The id should be a valid UUID, but for now we'll assume it is.
	// We'll return an error if the ID is somehow empty.
	if id == "" {
		return errors.New("id must be provided")
	}

	query := `
        DELETE FROM officers
        WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Use ExecContext for DELETE operations. This returns a sql.Result object
	// which contains information about how many rows were affected.
	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	// Check how many rows were affected.
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// If no rows were affected, it means our WHERE clause didn't match any records.
	// We can then return our custom ErrRecordNotFound error.
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// GetAll returns a paginated and filtered list of officers.
func (m OfficerModel) GetAll(firstName string, lastName string, rankCode string, filters Filters) ([]*Officer, Metadata, error) {
	// Use a window function to get the total number of records.
	query := fmt.Sprintf(`
        SELECT count(*) OVER(), id, regulation_number, first_name, last_name, sex, rank_code,
               created_at, updated_at, version
        FROM officers
        WHERE (to_tsvector('simple', first_name) @@ plainto_tsquery('simple', $1) OR $1 = '')
        AND (to_tsvector('simple', last_name) @@ plainto_tsquery('simple', $2) OR $2 = '')
        AND (LOWER(rank_code) = LOWER($3) OR $3 = '')
        ORDER BY %s %s, id ASC
        LIMIT $4 OFFSET $5`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{firstName, lastName, rankCode, filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := int64(0)
	officers := []*Officer{}

	for rows.Next() {
		var officer Officer
		err := rows.Scan(
			&totalRecords, // Scan the total count
			&officer.ID,
			&officer.RegulationNumber,
			&officer.FirstName,
			&officer.LastName,
			&officer.Sex,
			&officer.RankCode,
			&officer.CreatedAt,
			&officer.UpdatedAt,
			&officer.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		officers = append(officers, &officer)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return officers, metadata, nil
}