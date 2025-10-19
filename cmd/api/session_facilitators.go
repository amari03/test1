package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/amari03/test1/internal/data"
	"github.com/amari03/test1/internal/validator"
	"github.com/julienschmidt/httprouter"
)

// createSessionFacilitatorHandler handles POST /v1/session-facilitators
func (app *application) createSessionFacilitatorHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		SessionID     string  `json:"session_id"`
		FacilitatorID string  `json:"facilitator_id"`
		Role          *string `json:"role"` // ADD THIS
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	sf := &data.SessionFacilitator{
		SessionID:     input.SessionID,
		FacilitatorID: input.FacilitatorID,
		Role:          input.Role, // ADD THIS
	}

	v := validator.New()
	if data.ValidateSessionFacilitator(v, sf); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.SessionFacilitators.Insert(sf)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/session-facilitators/%s", sf.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"session_facilitator": sf}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// deleteSessionFacilitatorHandler handles DELETE /v1/session-facilitators/:id
func (app *application) deleteSessionFacilitatorHandler(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id := params.ByName("id")

	err := app.models.SessionFacilitators.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "session facilitator link successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listSessionFacilitatorsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		SessionID     string
		FacilitatorID string
		data.Filters
	}

	v := validator.New()
	qs := r.URL.Query()

	input.SessionID = app.readString(qs, "session_id", "")
	input.FacilitatorID = app.readString(qs, "facilitator_id", "")

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	
	// Define the safelist for sorting.
	input.Filters.SortSafelist = []string{"id", "session_id", "facilitator_id", "-id", "-session_id", "-facilitator_id"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	sfs, metadata, err := app.models.SessionFacilitators.GetAll(input.SessionID, input.FacilitatorID, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"session_facilitators": sfs, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}