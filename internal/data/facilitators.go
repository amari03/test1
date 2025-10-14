package data

import (
    "context"
	"database/sql"
	"errors"
	"fmt"
	"time"
    "github.com/amari03/test1/internal/validator"
)

type Facilitator struct {
    ID        string  `json:"id"`
    FirstName string  `json:"first_name"`
    LastName  string  `json:"last_name"`
    Notes     *string `json:"notes,omitempty"`
    Version   int32   `json:"version"`
}

type FacilitatorModel struct {
    DB *sql.DB
}

func ValidateFacilitator(v *validator.Validator, facilitator *Facilitator) {
    v.Check(facilitator.FirstName != "", "first_name", "must be provided")
    v.Check(facilitator.LastName != "", "last_name", "must be provided")
}

func (m FacilitatorModel) Insert(facilitator *Facilitator) error {
	query := `
        INSERT INTO facilitators (first_name, last_name, notes)
        VALUES ($1, $2, $3)
        RETURNING id, version`

	args := []interface{}{facilitator.FirstName, facilitator.LastName, facilitator.Notes}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&facilitator.ID, &facilitator.Version)
}

// Get a specific facilitator by ID.
func (m FacilitatorModel) Get(id string) (*Facilitator, error) {
	query := `
        SELECT id, first_name, last_name, notes, version
        FROM facilitators
        WHERE id = $1`

	var facilitator Facilitator
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&facilitator.ID,
		&facilitator.FirstName,
		&facilitator.LastName,
		&facilitator.Notes,
		&facilitator.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &facilitator, nil
}

// Update a specific facilitator by ID.
func (m FacilitatorModel) Update(facilitator *Facilitator) error {
	query := `
        UPDATE facilitators
        SET first_name = $1, last_name = $2, notes = $3, version = version + 1
        WHERE id = $4 AND version = $5
        RETURNING version`

	args := []interface{}{
		facilitator.FirstName,
		facilitator.LastName,
		facilitator.Notes,
		facilitator.ID,
		facilitator.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&facilitator.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

// Delete a specific facilitator by ID.
func (m FacilitatorModel) Delete(id string) error {
	if id == "" {
		return ErrRecordNotFound
	}
	query := `
        DELETE FROM facilitators
        WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// GetAll returns a slice of all facilitators.
func (m FacilitatorModel) GetAll(firstName string, lastName string, filters Filters) ([]*Facilitator, Metadata, error) {
	query := fmt.Sprintf(`
        SELECT count(*) OVER(), id, first_name, last_name, notes, version
        FROM facilitators
        WHERE (to_tsvector('simple', first_name) @@ plainto_tsquery('simple', $1) OR $1 = '')
        AND (to_tsvector('simple', last_name) @@ plainto_tsquery('simple', $2) OR $2 = '')
        ORDER BY %s %s, id ASC
        LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{firstName, lastName, filters.limit(), filters.offset()}
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := int64(0)
	facilitators := []*Facilitator{}

	for rows.Next() {
		var facilitator Facilitator
		err := rows.Scan(
			&totalRecords,
			&facilitator.ID,
			&facilitator.FirstName,
			&facilitator.LastName,
			&facilitator.Notes,
			&facilitator.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		facilitators = append(facilitators, &facilitator)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return facilitators, metadata, nil
}