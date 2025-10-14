package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

// Models struct holds all data models for the application.
type Models struct {
	Users    UserModel
	Officers OfficerModel
	Courses  CourseModel
	// You will add more models here as you build them out
	// Sessions SessionModel
	// Attendance AttendanceModel
}

// NewModels initializes and returns a Models struct.
func NewModels(db *sql.DB) Models {
	return Models{
		Users:    UserModel{DB: db},
		Officers: OfficerModel{DB: db},
		Courses:  CourseModel{DB: db},
	}
}