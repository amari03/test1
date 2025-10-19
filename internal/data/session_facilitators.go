package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
	"fmt"

	"github.com/amari03/test1/internal/validator"
)

type SessionFacilitator struct {
	ID            string  `json:"id"`
	SessionID     string  `json:"session_id"`
	FacilitatorID string  `json:"facilitator_id"`
	Role          *string `json:"role,omitempty"`
	Version       int32   `json:"version"`
	// REMOVED CreatedAt field
}

type SessionFacilitatorModel struct {
	DB *sql.DB
}

func ValidateSessionFacilitator(v *validator.Validator, sf *SessionFacilitator) {
	v.Check(sf.SessionID != "", "session_id", "must be provided")
	v.Check(sf.FacilitatorID != "", "facilitator_id", "must be provided")
}

// Insert a new session_facilitator record.
func (m SessionFacilitatorModel) Insert(sf *SessionFacilitator) error {
	query := `
        INSERT INTO session_facilitators (session_id, facilitator_id, role)
        VALUES ($1, $2, $3)
        RETURNING id, version` // UPDATED THIS LINE (removed created_at)

	args := []interface{}{sf.SessionID, sf.FacilitatorID, sf.Role}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// UPDATED THIS LINE (removed &sf.CreatedAt)
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&sf.ID, &sf.Version)
}

// Delete a specific session_facilitator record by its ID.
func (m SessionFacilitatorModel) Delete(id string) error {
	if id == "" {
		return errors.New("id must be provided")
	}

	query := `
        DELETE FROM session_facilitators
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

// GetAll returns a paginated and filtered list of session-facilitator links.
func (m SessionFacilitatorModel) GetAll(sessionID string, facilitatorID string, filters Filters) ([]*SessionFacilitator, Metadata, error) {
	query := fmt.Sprintf(`
        SELECT count(*) OVER(), id, session_id, facilitator_id, role, version
        FROM session_facilitators
        WHERE (session_id::text = $1 OR $1 = '')
        AND (facilitator_id::text = $2 OR $2 = '')
        ORDER BY %s %s, id ASC
        LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{sessionID, facilitatorID, filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := int64(0)
	sfs := []*SessionFacilitator{}

	for rows.Next() {
		var sf SessionFacilitator
		err := rows.Scan(
			&totalRecords,
			&sf.ID,
			&sf.SessionID,
			&sf.FacilitatorID,
			&sf.Role,
			&sf.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		sfs = append(sfs, &sf)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return sfs, metadata, nil
}