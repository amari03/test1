package main

import (
	"fmt"
	"net/http"

	"github.com/amari03/test1/internal/data"
	"github.com/amari03/test1/internal/validator"
)

// createSessionFeedbackHandler handles POST /v1/session-feedback
func (app *application) createSessionFeedbackHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		SessionID     string  `json:"session_id"`
		OfficerID     string  `json:"officer_id"`
		FacilitatorID string  `json:"facilitator_id"`
		Rating        float64 `json:"rating"` // CHANGE TO float64
		Comments      *string `json:"comments"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	feedback := &data.SessionFeedback{
		SessionID:     input.SessionID,
		OfficerID:     input.OfficerID,
		FacilitatorID: input.FacilitatorID,
		Rating:        input.Rating,
		Comments:      input.Comments,
	}

	v := validator.New()
	if data.ValidateSessionFeedback(v, feedback); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.SessionFeedback.Insert(feedback)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/session-feedback/%s", feedback.ID))

	err = app.writeJSON(w, http.StatusOK, envelope{"session_feedback": feedback}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}


func (app *application) listSessionFeedbackHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		SessionID     string
		OfficerID     string
		FacilitatorID string
		data.Filters
	}

	v := validator.New()
	qs := r.URL.Query()

	input.SessionID = app.readString(qs, "session_id", "")
	input.OfficerID = app.readString(qs, "officer_id", "")
	input.FacilitatorID = app.readString(qs, "facilitator_id", "")

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "created_at")

	// Define the safelist for sorting.
	input.Filters.SortSafelist = []string{"id", "created_at", "rating", "-id", "-created_at", "-rating"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	feedbacks, metadata, err := app.models.SessionFeedback.GetAll(input.SessionID, input.OfficerID, input.FacilitatorID, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"session_feedback": feedbacks, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}