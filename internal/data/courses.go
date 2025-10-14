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
	UpdatedAt          *time.Time `json:"updated_at,omitempty"`
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

// Insert a new course record into the database.
func (m CourseModel) Insert(course *Course) error {
    query := `
        INSERT INTO courses (title, category, default_credit_hours, description, created_by_user_id)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, created_at`

    args := []interface{}{
        course.Title,
        course.Category,
        course.DefaultCreditHours,
        course.Description,
        course.CreatedByUserID,
    }

    return m.DB.QueryRow(query, args...).Scan(&course.ID, &course.CreatedAt)
}


// Get a specific course by ID.
func (m CourseModel) Get(id string) (*Course, error) {
    query := `
        SELECT id, title, category, default_credit_hours, description, created_by_user_id, created_at, updated_at
        FROM courses
        WHERE id = $1`

    var course Course
    err := m.DB.QueryRow(query, id).Scan(
        &course.ID,
        &course.Title,
        &course.Category,
        &course.DefaultCreditHours,
        &course.Description,
        &course.CreatedByUserID,
        &course.CreatedAt,
        &course.UpdatedAt,
    )

    if err != nil {
        // This is how you handle a "not found" error specifically
        if err == sql.ErrNoRows {
            return nil, ErrRecordNotFound
        }
        return nil, err
    }
    return &course, nil
}

// Update a specific course record.
func (m CourseModel) Update(course *Course) error {
    query := `
        UPDATE courses
        SET title = $1, category = $2, default_credit_hours = $3, description = $4, updated_at = NOW()
        WHERE id = $5
        RETURNING updated_at`

    args := []interface{}{
        course.Title,
        course.Category,
        course.DefaultCreditHours,
        course.Description,
        course.ID,
    }

    return m.DB.QueryRow(query, args...).Scan(&course.UpdatedAt)
}

// Delete a specific course by ID.
func (m CourseModel) Delete(id string) error {
    query := `
        DELETE FROM courses
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