package main

import (
    "fmt"
    "net/http"
    "time"

    "github.com/amari03/test1/internal/data"
    "github.com/amari03/test1/internal/validator"
	"github.com/julienschmidt/httprouter"
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

func (app *application) getSessionHandler(w http.ResponseWriter, r *http.Request) {
    params := httprouter.ParamsFromContext(r.Context())
    id := params.ByName("id")

    session, err := app.models.Sessions.Get(id)
    if err != nil {
        if err == data.ErrRecordNotFound {
            app.notFoundResponse(w, r)
            return
        }
        app.serverErrorResponse(w, r, err)
        return
    }

    err = app.writeJSON(w, http.StatusOK, envelope{"session": session}, nil)
    if err != nil {
        app.serverErrorResponse(w, r, err)
    }
}

func (app *application) updateSessionHandler(w http.ResponseWriter, r *http.Request) {
    params := httprouter.ParamsFromContext(r.Context())
    id := params.ByName("id")

    session, err := app.models.Sessions.Get(id)
    if err != nil {
        app.notFoundResponse(w, r)
        return
    }

    var input struct {
        CourseID *string    `json:"course_id"`
        Start    *time.Time `json:"start_datetime"`
        End      *time.Time `json:"end_datetime"`
        Location *string    `json:"location_text"`
    }

    err = app.readJSON(w, r, &input)
    if err != nil {
        app.badRequestResponse(w, r, err)
        return
    }

    if input.CourseID != nil {
        session.CourseID = *input.CourseID
    }
    if input.Start != nil {
        session.Start = *input.Start
    }
    if input.End != nil {
        session.End = *input.End
    }
    if input.Location != nil {
        session.Location = *input.Location
    }

    v := validator.New()
    if data.ValidateSession(v, session); !v.Valid() {
        app.failedValidationResponse(w, r, v.Errors)
        return
    }

    err = app.models.Sessions.Update(session)
    if err != nil {
        app.serverErrorResponse(w, r, err)
        return
    }

    err = app.writeJSON(w, http.StatusOK, envelope{"session": session}, nil)
    if err != nil {
        app.serverErrorResponse(w, r, err)
    }
}