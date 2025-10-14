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