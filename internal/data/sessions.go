package data

import (
    "context"
	"database/sql"
	"errors"
	"fmt"
	"time"

    "github.com/amari03/test1/internal/validator"
)

type Session struct {
    ID          string     `json:"id"`
    CourseID    string     `json:"course_id"`
    Start       time.Time  `json:"start_datetime"`
    End         time.Time  `json:"end_datetime"`
    Location    string     `json:"location_text"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   *time.Time `json:"updated_at,omitempty"`
    Version   int32      `json:"version"`
}

type SessionModel struct {
    DB *sql.DB
}

// We'll add a basic validator
func ValidateSession(v *validator.Validator, session *Session) {
	v.Check(session.CourseID != "", "course_id", "must be provided")
	v.Check(!session.Start.IsZero(), "start_datetime", "must be provided")
	v.Check(!session.End.IsZero(), "end_datetime", "must be provided")
	v.Check(session.End.After(session.Start), "end_datetime", "must be after start_datetime")
	v.Check(session.Location != "", "location_text", "must be provided")
}

func (m SessionModel) Insert(session *Session) error {
	query := `
        INSERT INTO sessions (course_id, start_datetime, end_datetime, location_text)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, version`

	args := []interface{}{
		session.CourseID, 
		session.Start, 
		session.End, 
		session.Location,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&session.ID, &session.CreatedAt, &session.Version)
}

// Get a specific session by ID.
func (m SessionModel) Get(id string) (*Session, error) {
	query := `
        SELECT id, course_id, start_datetime, end_datetime, location_text,
               created_at, updated_at, version
        FROM sessions
        WHERE id = $1`

	var session Session
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&session.ID,
		&session.CourseID,
		&session.Start,
		&session.End,
		&session.Location,
		&session.CreatedAt,
		&session.UpdatedAt,
		&session.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &session, nil
}

// Update a specific session record.
func (m SessionModel) Update(session *Session) error {
	query := `
        UPDATE sessions
        SET course_id = $1, start_datetime = $2, end_datetime = $3, location_text = $4,
            updated_at = NOW(), version = version + 1
        WHERE id = $5 AND version = $6
        RETURNING updated_at, version`

	args := []interface{}{
		session.CourseID,
		session.Start,
		session.End,
		session.Location,
		session.ID,
		session.Version,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&session.UpdatedAt, &session.Version)
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

// Delete a specific session by ID.
func (m SessionModel) Delete(id string) error {
    
	if id == "" {
		return ErrRecordNotFound
	}
	query := `DELETE FROM sessions WHERE id = $1`
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

// GetAll returns a slice of all sessions.
func (m SessionModel) GetAll(location string, courseID string, filters Filters) ([]*Session, Metadata, error) {
	query := fmt.Sprintf(`
        SELECT count(*) OVER(), id, course_id, start_datetime, end_datetime, location_text,
               created_at, updated_at, version
        FROM sessions
        WHERE (to_tsvector('simple', COALESCE(location_text, '')) @@ plainto_tsquery('simple', $1) OR $1 = '')
        AND (course_id::text = $2 OR $2 = '')
        ORDER BY %s %s, id ASC
        LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{location, courseID, filters.limit(), filters.offset()}
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := int64(0)
	sessions := []*Session{}
	for rows.Next() {
		var session Session
		err := rows.Scan(
			&totalRecords,
			&session.ID,
			&session.CourseID,
			&session.Start,
			&session.End,
			&session.Location,
			&session.CreatedAt,
			&session.UpdatedAt,
			&session.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		sessions = append(sessions, &session)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return sessions, metadata, nil
}