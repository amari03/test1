package main

import (
    "fmt"
    "net/http"
    "time"

    "github.com/amari03/test1/internal/data"
    "github.com/amari03/test1/internal/validator"
)

func (app *application) createSessionHandler(w http.ResponseWriter, r *http.Request) {
    var input struct {
        CourseID string    `json:"course_id"`
        Start    time.Time `json:"start_datetime"`
        End      time.Time `json:"end_datetime"`
        Location string    `json:"location_text"`
    }

    err := app.readJSON(w, r, &input)
    if err != nil {
        app.badRequestResponse(w, r, err)
        return
    }

    session := &data.Session{
        CourseID: input.CourseID,
        Start:    input.Start,
        End:      input.End,
        Location: input.Location,
    }

    v := validator.New()
    if data.ValidateSession(v, session); !v.Valid() {
        app.failedValidationResponse(w, r, v.Errors)
        return
    }

    err = app.models.Sessions.Insert(session)
    if err != nil {
        app.serverErrorResponse(w, r, err)
        return
    }

    headers := make(http.Header)
    headers.Set("Location", fmt.Sprintf("/v1/sessions/%s", session.ID))

    err = app.writeJSON(w, http.StatusCreated, envelope{"session": session}, headers)
    if err != nil {
        app.serverErrorResponse(w, r, err)
    }
}