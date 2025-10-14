//HANDLER FOR OFFICERS ENDPOINT
package main

import (
	"net/http"
	"fmt"
    "errors"

	"github.com/amari03/test1/internal/data"
	"github.com/amari03/test1/internal/validator"
    "github.com/julienschmidt/httprouter" 
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

func (app *application) updateOfficerHandler(w http.ResponseWriter, r *http.Request) {
    params := httprouter.ParamsFromContext(r.Context())
    id := params.ByName("id")

    officer, err := app.models.Officers.Get(id)
    if err != nil {
        app.notFoundResponse(w, r)
        return
    }

    var input struct {
        RegulationNumber *string `json:"regulation_number"`
        FirstName        *string `json:"first_name"`
        LastName         *string `json:"last_name"`
        Sex              *string `json:"sex"`
        RankCode         *string `json:"rank_code"`
    }

    err = app.readJSON(w, r, &input)
    if err != nil {
        app.badRequestResponse(w, r, err)
        return
    }

    if input.RegulationNumber != nil {
        officer.RegulationNumber = input.RegulationNumber
    }
    if input.FirstName != nil {
        officer.FirstName = *input.FirstName
    }
    if input.LastName != nil {
        officer.LastName = *input.LastName
    }
    if input.Sex != nil {
        officer.Sex = *input.Sex
    }
    if input.RankCode != nil {
        officer.RankCode = *input.RankCode
    }

    v := validator.New()
    if data.ValidateOfficer(v, officer); !v.Valid() {
        app.failedValidationResponse(w, r, v.Errors)
        return
    }

    err = app.models.Officers.Update(officer)
    if err != nil {
        app.serverErrorResponse(w, r, err)
        return
    }

    err = app.writeJSON(w, http.StatusOK, envelope{"officer": officer}, nil)
    if err != nil {
        app.serverErrorResponse(w, r, err)
    }
}

func (app *application) deleteOfficerHandler(w http.ResponseWriter, r *http.Request) {
    params := httprouter.ParamsFromContext(r.Context())
    id := params.ByName("id")

    err := app.models.Officers.Delete(id)
    if err != nil {
        switch {
        case errors.Is(err, data.ErrRecordNotFound):
            app.notFoundResponse(w, r)
        default:
            app.serverErrorResponse(w, r, err)
        }
        return
    }

    err = app.writeJSON(w, http.StatusOK, envelope{"message": "officer successfully deleted"}, nil)
    if err != nil {
        app.serverErrorResponse(w, r, err)
    }
}

// getOfficerHandler will handle GET /v1/officers/:id
func (app *application) getOfficerHandler(w http.ResponseWriter, r *http.Request) {
    params := httprouter.ParamsFromContext(r.Context())
    id := params.ByName("id")

    // You might want to add UUID validation here later, but for now this is fine.

    officer, err := app.models.Officers.Get(id)
    if err != nil {
        // A real app would check if it's a "not found" error vs a server error.
        // For now, we'll just send a not found for any error.
        app.notFoundResponse(w, r)
        return
    }

    err = app.writeJSON(w, http.StatusOK, envelope{"officer": officer}, nil)
    if err != nil {
        app.serverErrorResponse(w, r, err)
    }
}

func (app *application) listOfficersHandler(w http.ResponseWriter, r *http.Request) {
    // Get query parameters from the URL
    qs := r.URL.Query()
    firstName := qs.Get("first_name")
    lastName := qs.Get("last_name")
    rankCode := qs.Get("rank_code")

    officers, err := app.models.Officers.GetAll(firstName, lastName, rankCode)
    if err != nil {
        app.serverErrorResponse(w, r, err)
        return
    }

    err = app.writeJSON(w, http.StatusOK, envelope{"officers": officers}, nil)
    if err != nil {
        app.serverErrorResponse(w, r, err)
    }
}

