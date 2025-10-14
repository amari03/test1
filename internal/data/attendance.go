package data

import (
    "database/sql"
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
}

type AttendanceModel struct {
    DB *sql.DB
}

func (m AttendanceModel) Insert(attendance *Attendance) error {
    query := `
        INSERT INTO attendance (officer_id, session_id, status, credited_hours)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at`

    args := []interface{}{attendance.OfficerID, attendance.SessionID, attendance.Status, attendance.CreditedHours}
    return m.DB.QueryRow(query, args...).Scan(&attendance.ID, &attendance.CreatedAt)
}

func ValidateAttendance(v *validator.Validator, attendance *Attendance) {
    v.Check(attendance.OfficerID != "", "officer_id", "must be provided")
    v.Check(attendance.SessionID != "", "session_id", "must be provided")
    v.Check(validator.In(attendance.Status, "attended", "absent", "excused"), "status", "must be one of attended, absent, or excused")
    v.Check(attendance.CreditedHours >= 0, "credited_hours", "must be zero or greater")
}