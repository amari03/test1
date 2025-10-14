package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
    Version            int32      `json:"version"`
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
        RETURNING id, created_at, version`

	args := []interface{}{
		course.Title,
		course.Category,
		course.DefaultCreditHours,
		course.Description,
		course.CreatedByUserID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&course.ID, &course.CreatedAt, &course.Version)
}


// Get a specific course by ID.
func (m CourseModel) Get(id string) (*Course, error) {
	query := `
        SELECT id, title, category, default_credit_hours, description, created_by_user_id,
               created_at, updated_at, version
        FROM courses
        WHERE id = $1`

	var course Course
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&course.ID,
		&course.Title,
		&course.Category,
		&course.DefaultCreditHours,
		&course.Description,
		&course.CreatedByUserID,
		&course.CreatedAt,
		&course.UpdatedAt,
		&course.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &course, nil
}

// Update a specific course record.
func (m CourseModel) Update(course *Course) error {
	query := `
        UPDATE courses
        SET title = $1, category = $2, default_credit_hours = $3, description = $4,
            updated_at = NOW(), version = version + 1
        WHERE id = $5 AND version = $6
        RETURNING updated_at, version`

	args := []interface{}{
		course.Title,
		course.Category,
		course.DefaultCreditHours,
		course.Description,
		course.ID,
		course.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&course.UpdatedAt, &course.Version)
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

// Delete a specific course by ID.
func (m CourseModel) Delete(id string) error {
	if id == "" {
		return ErrRecordNotFound
	}
	query := `
        DELETE FROM courses
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

// GetAll returns a slice of all courses, with filtering.
func (m CourseModel) GetAll(title string, category string, filters Filters) ([]*Course, Metadata, error) {
	query := fmt.Sprintf(`
        SELECT count(*) OVER(), id, title, category, default_credit_hours, description,
               created_by_user_id, created_at, updated_at, version
        FROM courses
        WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
        AND (LOWER(category) = LOWER($2) OR $2 = '')
        ORDER BY %s %s, id ASC
        LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{title, category, filters.limit(), filters.offset()}
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := int64(0)
	courses := []*Course{}

	for rows.Next() {
		var course Course
		err := rows.Scan(
			&totalRecords,
			&course.ID,
			&course.Title,
			&course.Category,
			&course.DefaultCreditHours,
			&course.Description,
			&course.CreatedByUserID,
			&course.CreatedAt,
			&course.UpdatedAt,
			&course.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		courses = append(courses, &course)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return courses, metadata, nil
}