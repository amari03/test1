package data

import (
	"database/sql"
	"time"

	"github.com/amari03/test1/internal/validator"
)

type Course struct {
	ID                 string    `json:"id"`
	Title              string    `json:"title"`
	Category           string    `json:"category"`
	DefaultCreditHours float64   `json:"default_credit_hours"`
	Description        string    `json:"description,omitempty"`
	CreatedByUserID    string    `json:"created_by_user_id"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type CourseModel struct {
	DB *sql.DB
}

func ValidateCourse(v *validator.Validator, course *Course) {
	v.Check(course.Title != "", "title", "must be provided")
	v.Check(len(course.Title) <= 255, "title", "must not exceed 255 bytes")
	v.Check(course.Category != "", "category", "must be provided")
	v.Check(validator.In(course.Category, "mandatory", "elective", "instructor"), "category", "invalid category type")
	v.Check(course.DefaultCreditHours > 0, "default_credit_hours", "must be greater than zero")
}