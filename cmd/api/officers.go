//HANDLER FOR OFFICERS ENDPOINT
package main

import (
	"net/http"
	"fmt"

	"github.com/amari03/test1/internal/data"
	"github.com/amari03/test1/internal/validator"
)

func (app *application) createOfficerHandler(w http.ResponseWriter, r *http.Request) {
    var input struct {
        RegulationNumber *string `json:"regulation_number"`
        FirstName        string  `json:"first_name"`
        LastName         string  `json:"last_name"`
        Sex              string  `json:"sex"`
        RankCode         string  `json:"rank_code"`
    }

    err := app.readJSON(w, r, &input)
    if err != nil {
        app.badRequestResponse(w, r, err)
        return
    }

    officer := &data.Officer{
        RegulationNumber: input.RegulationNumber,
        FirstName:        input.FirstName,
        LastName:         input.LastName,
        Sex:              input.Sex,
        RankCode:         input.RankCode,
    }

    v := validator.New()
    if data.ValidateOfficer(v, officer); !v.Valid() {
        app.failedValidationResponse(w, r, v.Errors)
        return
    }

    err = app.models.Officers.Insert(officer)
    if err != nil {
        app.serverErrorResponse(w, r, err)
        return
    }

    headers := make(http.Header)
    headers.Set("Location", fmt.Sprintf("/v1/officers/%s", officer.ID))

    err = app.writeJSON(w, http.StatusCreated, envelope{"officer": officer}, headers)
    if err != nil {
        app.serverErrorResponse(w, r, err)
    }
}

// getOfficerHandler will handle GET /v1/officers/:id
func (app *application) getOfficerHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement logic to get a specific officer by ID.
	w.Write([]byte("TODO: Get officer by ID"))
}

// listOfficersHandler will handle GET /v1/officers
func (app *application) listOfficersHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement logic to list all officers with filtering and pagination.
	w.Write([]byte("TODO: List all officers"))
}