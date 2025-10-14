package data

import (
    "database/sql"
    "github.com/amari03/test1/internal/validator"
)

type Facilitator struct {
    ID        string  `json:"id"`
    FirstName string  `json:"first_name"`
    LastName  string  `json:"last_name"`
    Notes     *string `json:"notes,omitempty"`
}

type FacilitatorModel struct {
    DB *sql.DB
}

func (m FacilitatorModel) Insert(facilitator *Facilitator) error {
    query := `
        INSERT INTO facilitators (first_name, last_name, notes)
        VALUES ($1, $2, $3)
        RETURNING id`

    args := []interface{}{facilitator.FirstName, facilitator.LastName, facilitator.Notes}
    return m.DB.QueryRow(query, args...).Scan(&facilitator.ID)
}

func ValidateFacilitator(v *validator.Validator, facilitator *Facilitator) {
    v.Check(facilitator.FirstName != "", "first_name", "must be provided")
    v.Check(facilitator.LastName != "", "last_name", "must be provided")
}

// Get a specific facilitator by ID.
func (m FacilitatorModel) Get(id string) (*Facilitator, error) {
    query := `
        SELECT id, first_name, last_name, notes
        FROM facilitators
        WHERE id = $1`

    var facilitator Facilitator
    err := m.DB.QueryRow(query, id).Scan(
        &facilitator.ID,
        &facilitator.FirstName,
        &facilitator.LastName,
        &facilitator.Notes,
    )

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, ErrRecordNotFound
        }
        return nil, err
    }
    return &facilitator, nil
}


// Update a specific facilitator record.
func (m FacilitatorModel) Update(facilitator *Facilitator) error {
    query := `
        UPDATE facilitators
        SET first_name = $1, last_name = $2, notes = $3
        WHERE id = $4`

    args := []interface{}{
        facilitator.FirstName,
        facilitator.LastName,
        facilitator.Notes,
        facilitator.ID,
    }

    _, err := m.DB.Exec(query, args...)
    return err
}

// Delete a specific facilitator by ID.
func (m FacilitatorModel) Delete(id string) error {
    query := `
        DELETE FROM facilitators
        WHERE id = $1`

    result, err := m.DB.Exec(query, id)
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