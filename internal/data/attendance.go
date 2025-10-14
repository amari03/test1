package data

import (
    "context"
	"database/sql"
	"errors"
	"fmt"
	"time"

    "github.com/amari03/test1/internal/validator"
)

type Attendance struct {
    ID            string    `json:"id"`
    OfficerID     string    `json:"officer_id"`
    SessionID     string    `json:"session_id"`
    Status        string    `json:"status"`
    CreditedHours float64   `json:"credited_hours"`
    CreatedAt     time.Time `json:"created_at"`
    Version       int32      `json:"version"` 
}

type AttendanceModel struct {
    DB *sql.DB
}

func ValidateAttendance(v *validator.Validator, attendance *Attendance) {
    v.Check(attendance.OfficerID != "", "officer_id", "must be provided")
    v.Check(attendance.SessionID != "", "session_id", "must be provided")
    v.Check(validator.In(attendance.Status, "attended", "absent", "excused"), "status", "must be one of attended, absent, or excused")
    v.Check(attendance.CreditedHours >= 0, "credited_hours", "must be zero or greater")
}

func (m AttendanceModel) Insert(attendance *Attendance) error {
	query := `
        INSERT INTO attendance (officer_id, session_id, status, credited_hours)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, version`

	args := []interface{}{attendance.OfficerID, attendance.SessionID, attendance.Status, attendance.CreditedHours}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&attendance.ID, &attendance.CreatedAt, &attendance.Version)
}

func (m AttendanceModel) Get(id string) (*Attendance, error) {
	query := `
        SELECT id, officer_id, session_id, status, credited_hours, created_at, version
        FROM attendance
        WHERE id = $1`

	var record Attendance
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&record.ID,
		&record.OfficerID,
		&record.SessionID,
		&record.Status,
		&record.CreditedHours,
		&record.CreatedAt,
		&record.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &record, nil
}


func (m AttendanceModel) Update(attendance *Attendance) error {
	query := `
        UPDATE attendance
        SET status = $1, credited_hours = $2, version = version + 1
        WHERE id = $3 AND version = $4
        RETURNING version`

	args := []interface{}{
		attendance.Status,
		attendance.CreditedHours,
		attendance.ID,
		attendance.Version,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&attendance.Version)
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

// Delete a specific attendance record by ID.
func (m AttendanceModel) Delete(id string) error {
	if id == "" {
		return ErrRecordNotFound
	}
	query := `DELETE FROM attendance WHERE id = $1`
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

func (m AttendanceModel) GetAll(officerID string, sessionID string, filters Filters) ([]*Attendance, Metadata, error) {
	query := fmt.Sprintf(`
        SELECT count(*) OVER(), id, officer_id, session_id, status, credited_hours, created_at, version
        FROM attendance
        WHERE (officer_id::text = $1 OR $1 = '')
        AND (session_id::text = $2 OR $2 = '')
        ORDER BY %s %s, id ASC
        LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{officerID, sessionID, filters.limit(), filters.offset()}
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := int64(0)
	records := []*Attendance{}
	for rows.Next() {
		var attendance Attendance
		err := rows.Scan(
			&totalRecords,
			&attendance.ID,
			&attendance.OfficerID,
			&attendance.SessionID,
			&attendance.Status,
			&attendance.CreditedHours,
			&attendance.CreatedAt,
			&attendance.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		records = append(records, &attendance)
	}
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return records, metadata, nil
}