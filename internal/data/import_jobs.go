package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/amari03/test1/internal/validator"
)

type ImportJob struct {
	ID              string     `json:"id"`
	Type            string     `json:"type"`
	Status          string     `json:"status"`
	ErrorMessage    *string    `json:"error_message,omitempty"` // ADDED
	CreatedByUserID string     `json:"created_by_user_id"`
	CreatedAt       time.Time  `json:"created_at"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"` // ADDED
	Version         int32      `json:"version"`
}

type ImportJobModel struct {
	DB *sql.DB
}

func ValidateImportJob(v *validator.Validator, job *ImportJob) {
	v.Check(job.Type != "", "type", "must be provided")
	// v.Check(validator.In(job.Type, "officers", "attendance", "courses"), "type", "must be 'officers', 'attendance', or 'courses'")
    // You can add the check above if you know the types
}

// Insert a new import job record.
func (m ImportJobModel) Insert(job *ImportJob) error {
	query := `
        INSERT INTO import_jobs (type, status, created_by_user_id)
        VALUES ($1, $2, $3)
        RETURNING id, created_at, version` // REMOVED file_path

	// Default status to "pending"
	job.Status = "pending"

	args := []interface{}{job.Type, job.Status, job.CreatedByUserID} // REMOVED file_path

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&job.ID, &job.CreatedAt, &job.Version)
}

// Get a specific import job by ID.
func (m ImportJobModel) Get(id string) (*ImportJob, error) {
	query := `
        SELECT id, type, status, error_message, created_by_user_id,
               created_at, finished_at, version
        FROM import_jobs
        WHERE id = $1` // UPDATED fields

	var job ImportJob

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&job.ID,
		&job.Type,
		&job.Status,
		&job.ErrorMessage, // UPDATED
		&job.CreatedByUserID,
		&job.CreatedAt,
		&job.FinishedAt, // UPDATED
		&job.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &job, nil
}

// GetAll returns a paginated list of import jobs, filterable by type and status.
func (m ImportJobModel) GetAll(jobType string, status string, filters Filters) ([]*ImportJob, Metadata, error) {
	query := fmt.Sprintf(`
        SELECT count(*) OVER(), id, type, status, error_message, created_by_user_id,
               created_at, finished_at, version
        FROM import_jobs
        WHERE (LOWER(type) = LOWER($1) OR $1 = '')
        AND (LOWER(status) = LOWER($2) OR $2 = '')
        ORDER BY %s %s, id ASC
        LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection()) // UPDATED fields

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{jobType, status, filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := int64(0)
	jobs := []*ImportJob{}

	for rows.Next() {
		var job ImportJob
		err := rows.Scan(
			&totalRecords,
			&job.ID,
			&job.Type,
			&job.Status,
			&job.ErrorMessage, // UPDATED
			&job.CreatedByUserID,
			&job.CreatedAt,
			&job.FinishedAt, // UPDATED
			&job.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		jobs = append(jobs, &job)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return jobs, metadata, nil
}