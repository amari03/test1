package data

import (
	"context"
	"database/sql"
	"time"
	"fmt"

	"github.com/amari03/test1/internal/validator"
)

type SessionFeedback struct {
	ID            string     `json:"id"`
	SessionID     string     `json:"session_id"`
	OfficerID     string     `json:"officer_id"`
	FacilitatorID string     `json:"facilitator_id"`
	Rating        float64    `json:"rating"` // CHANGE TO float64
	Comments      *string    `json:"comments,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	Version       int32      `json:"version"` // ADD THIS
}

type SessionFeedbackModel struct {
	DB *sql.DB
}

func ValidateSessionFeedback(v *validator.Validator, sf *SessionFeedback) {
	v.Check(sf.SessionID != "", "session_id", "must be provided")
	v.Check(sf.OfficerID != "", "officer_id", "must be provided")
	v.Check(sf.FacilitatorID != "", "facilitator_id", "must be provided")
	v.Check(sf.Rating >= 1 && sf.Rating <= 5, "rating", "must be between 1 and 5") // This check works for float64 too
}

// Insert a new session_feedback record.
func (m SessionFeedbackModel) Insert(sf *SessionFeedback) error {
	query := `
        INSERT INTO session_feedback (session_id, officer_id, facilitator_id, rating, comments)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, created_at, version` // UPDATE THIS

	args := []interface{}{
		sf.SessionID,
		sf.OfficerID,
		sf.FacilitatorID,
		sf.Rating,
		sf.Comments,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&sf.ID, &sf.CreatedAt, &sf.Version) // UPDATE THIS
}

// GetAll returns a paginated and filtered list of session feedback.
func (m SessionFeedbackModel) GetAll(sessionID string, officerID string, facilitatorID string, filters Filters) ([]*SessionFeedback, Metadata, error) {
	query := fmt.Sprintf(`
        SELECT count(*) OVER(), id, session_id, officer_id, facilitator_id, rating, comments, created_at, version
        FROM session_feedback
        WHERE (session_id::text = $1 OR $1 = '')
        AND (officer_id::text = $2 OR $2 = '')
        AND (facilitator_id::text = $3 OR $3 = '')
        ORDER BY %s %s, id ASC
        LIMIT $4 OFFSET $5`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{sessionID, officerID, facilitatorID, filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := int64(0)
	feedbacks := []*SessionFeedback{}

	for rows.Next() {
		var feedback SessionFeedback
		err := rows.Scan(
			&totalRecords,
			&feedback.ID,
			&feedback.SessionID,
			&feedback.OfficerID,
			&feedback.FacilitatorID,
			&feedback.Rating,
			&feedback.Comments,
			&feedback.CreatedAt,
			&feedback.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		feedbacks = append(feedbacks, &feedback)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return feedbacks, metadata, nil
}