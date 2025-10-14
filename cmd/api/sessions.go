package main

import (
    "fmt"
    "net/http"
    "time"
    "errors"

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
        switch {
        case errors.Is(err, data.ErrRecordNotFound):
            app.notFoundResponse(w, r)
        default:
            app.serverErrorResponse(w, r, err)
        }
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
        switch {
        case errors.Is(err, data.ErrRecordNotFound):
            app.notFoundResponse(w, r)
        default:
            app.serverErrorResponse(w, r, err)
        }
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

    if input.CourseID != nil { session.CourseID = *input.CourseID }
    if input.Start != nil { session.Start = *input.Start }
    if input.End != nil { session.End = *input.End }
    if input.Location != nil { session.Location = *input.Location }

    v := validator.New()
    if data.ValidateSession(v, session); !v.Valid() {
        app.failedValidationResponse(w, r, v.Errors)
        return
    }

    err = app.models.Sessions.Update(session)
    if err != nil {
        switch {
        case errors.Is(err, data.ErrEditConflict):
            app.editConflictResponse(w, r)
        default:
            app.serverErrorResponse(w, r, err)
        }
        return
    }

    err = app.writeJSON(w, http.StatusOK, envelope{"session": session}, nil)
    if err != nil {
        app.serverErrorResponse(w, r, err)
    }
}

func (app *application) deleteSessionHandler(w http.ResponseWriter, r *http.Request) {
    params := httprouter.ParamsFromContext(r.Context())
    id := params.ByName("id")

    err := app.models.Sessions.Delete(id)
    if err != nil {
        switch {
        case errors.Is(err, data.ErrRecordNotFound):
            app.notFoundResponse(w, r)
        default:
            app.serverErrorResponse(w, r, err)
        }
        return
    }

    err = app.writeJSON(w, http.StatusOK, envelope{"message": "session successfully deleted"}, nil)
    if err != nil {
        app.serverErrorResponse(w, r, err)
    }
}

func (app *application) listSessionsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Location string
		CourseID string
		data.Filters
	}

	v := validator.New()
	qs := r.URL.Query()

	input.Location = app.readString(qs, "location", "")
	input.CourseID = app.readString(qs, "course_id", "")

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "start_datetime") // A sensible default for sessions
	input.Filters.SortSafelist = []string{"id", "start_datetime", "end_datetime", "location_text", "-id", "-start_datetime", "-end_datetime", "-location_text"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	sessions, metadata, err := app.models.Sessions.GetAll(input.Location, input.CourseID, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"sessions": sessions, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}