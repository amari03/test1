package data

import (
    "database/sql"
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
}

type SessionModel struct {
    DB *sql.DB
}

func (m SessionModel) Insert(session *Session) error {
    query := `
        INSERT INTO sessions (course_id, start_datetime, end_datetime, location_text)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at`

    args := []interface{}{session.CourseID, session.Start, session.End, session.Location}
    return m.DB.QueryRow(query, args...).Scan(&session.ID, &session.CreatedAt)
}

// We'll add a basic validator
func ValidateSession(v *validator.Validator, session *Session) {
    v.Check(session.CourseID != "", "course_id", "must be provided")
    v.Check(!session.Start.IsZero(), "start_datetime", "must be provided")
    v.Check(!session.End.IsZero(), "end_datetime", "must be provided")
    v.Check(session.End.After(session.Start), "end_datetime", "must be after start_datetime")
}

// Get a specific session by ID.
func (m SessionModel) Get(id string) (*Session, error) {
    query := `
        SELECT id, course_id, start_datetime, end_datetime, location_text, created_at, updated_at
        FROM sessions
        WHERE id = $1`

    var session Session
    err := m.DB.QueryRow(query, id).Scan(
        &session.ID,
        &session.CourseID,
        &session.Start,
        &session.End,
        &session.Location,
        &session.CreatedAt,
        &session.UpdatedAt,
    )

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, ErrRecordNotFound
        }
        return nil, err
    }
    return &session, nil
}

// Update a specific session record.
func (m SessionModel) Update(session *Session) error {
    query := `
        UPDATE sessions
        SET course_id = $1, start_datetime = $2, end_datetime = $3, location_text = $4, updated_at = NOW()
        WHERE id = $5
        RETURNING updated_at`

    args := []interface{}{
        session.CourseID,
        session.Start,
        session.End,
        session.Location,
        session.ID,
    }

    return m.DB.QueryRow(query, args...).Scan(&session.UpdatedAt)
}

// Delete a specific session by ID.
func (m SessionModel) Delete(id string) error {
    query := `
        DELETE FROM sessions
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