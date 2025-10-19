package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/amari03/test1/internal/data"
	"github.com/amari03/test1/internal/validator"
	"github.com/julienschmidt/httprouter"
)

// createImportJobHandler handles POST /v1/import-jobs
func (app *application) createImportJobHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Type string `json:"type"`
		// FilePath is removed
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Following the pattern from your courses.go handler
	job := &data.ImportJob{
		Type:            input.Type,
		CreatedByUserID: "1a7a5180-4303-4318-8789-1a007f339eec", // Dummy user
	}

	v := validator.New()
	if data.ValidateImportJob(v, job); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.ImportJobs.Insert(job)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/import-jobs/%s", job.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"import_job": job}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// getImportJobHandler handles GET /v1/import-jobs/:id
func (app *application) getImportJobHandler(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id := params.ByName("id")

	job, err := app.models.ImportJobs.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"import_job": job}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// listImportJobsHandler handles GET /v1/import-jobs
func (app *application) listImportJobsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Type   string
		Status string
		data.Filters
	}

	v := validator.New()
	qs := r.URL.Query()

	input.Type = app.readString(qs, "type", "")
	input.Status = app.readString(qs, "status", "")

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "created_at")

	// Define the safelist for sorting.
	input.Filters.SortSafelist = []string{"id", "type", "status", "created_at", "-id", "-type", "-status", "-created_at"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	jobs, metadata, err := app.models.ImportJobs.GetAll(input.Type, input.Status, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"import_jobs": jobs, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}