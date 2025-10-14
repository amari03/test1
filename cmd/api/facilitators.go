package main

import (
    "fmt"
    "net/http"

    "github.com/amari03/test1/internal/data"
    "github.com/amari03/test1/internal/validator"
)

func (app *application) createFacilitatorHandler(w http.ResponseWriter, r *http.Request) {
    var input struct {
        FirstName string  `json:"first_name"`
        LastName  string  `json:"last_name"`
        Notes     *string `json:"notes"`
    }

    err := app.readJSON(w, r, &input)
    if err != nil {
        app.badRequestResponse(w, r, err)
        return
    }

    facilitator := &data.Facilitator{
        FirstName: input.FirstName,
        LastName:  input.LastName,
        Notes:     input.Notes,
    }

    v := validator.New()
    if data.ValidateFacilitator(v, facilitator); !v.Valid() {
        app.failedValidationResponse(w, r, v.Errors)
        return
    }

    err = app.models.Facilitators.Insert(facilitator)
    if err != nil {
        app.serverErrorResponse(w, r, err)
        return
    }

    headers := make(http.Header)
    headers.Set("Location", fmt.Sprintf("/v1/facilitators/%s", facilitator.ID))

    err = app.writeJSON(w, http.StatusCreated, envelope{"facilitator": facilitator}, headers)
    if err != nil {
        app.serverErrorResponse(w, r, err)
    }
}