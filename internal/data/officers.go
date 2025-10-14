package data

import (
	"database/sql"
	"time"

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
        SELECT id, regulation_number, first_name, last_name, sex, rank_code, created_at, updated_at
        FROM officers
        WHERE id = $1`

	var officer Officer
	err := m.DB.QueryRow(query, id).Scan(
		&officer.ID,
		&officer.RegulationNumber,
		&officer.FirstName,
		&officer.LastName,
		&officer.Sex,
		&officer.RankCode,
		&officer.CreatedAt,
		&officer.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}
	return &officer, nil
}

// Update a specific officer record.
func (m OfficerModel) Update(officer *Officer) error {
    query := `
        UPDATE officers
        SET regulation_number = $1, first_name = $2, last_name = $3, sex = $4, rank_code = $5, updated_at = NOW()
        WHERE id = $6
        RETURNING updated_at`

    args := []interface{}{
        officer.RegulationNumber,
        officer.FirstName,
        officer.LastName,
        officer.Sex,
        officer.RankCode,
        officer.ID,
    }

    return m.DB.QueryRow(query, args...).Scan(&officer.UpdatedAt)
}