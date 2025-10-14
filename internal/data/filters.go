package data

import (
	"math"
	"strings"

	"github.com/amari03/test1/internal/validator"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
}

// ValidateFilters checks that the page and page_size parameters are valid.
// It also checks that the sort parameter is a valid field.
func ValidateFilters(v *validator.Validator, f Filters) {
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 10_000, "page", "must be a maximum of 10,000")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")

	// Check that the sort parameter is in the safelist.
	v.Check(validator.In(f.Sort, f.SortSafelist...), "sort", "invalid sort value")
}

// sortColumn checks if the client-provided sort field is in the safelist.
// If it is, it extracts the column name by stripping the '-' prefix.
func (f Filters) sortColumn() string {
	for _, safeValue := range f.SortSafelist {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}
	panic("unsafe sort parameter: " + f.Sort) // Should not happen if validation is used
}

// sortDirection returns the sort direction ("ASC" or "DESC") depending on the '-' prefix.
func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}

func (f Filters) limit() int {
	return f.PageSize
}

func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

// --- Metadata for Pagination ---

type Metadata struct {
	CurrentPage  int   `json:"current_page,omitempty"`
	PageSize     int   `json:"page_size,omitempty"`
	FirstPage    int   `json:"first_page,omitempty"`
	LastPage     int   `json:"last_page,omitempty"`
	TotalRecords int64 `json:"total_records,omitempty"`
}

// calculateMetadata calculates the metadata for a paginated response.
func calculateMetadata(totalRecords int64, page, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}
	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}