package main

import (
    "fmt"
    "net/http"
	"errors"

    "github.com/amari03/test1/internal/data"
    "github.com/amari03/test1/internal/validator"
	"github.com/julienschmidt/httprouter"
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

func (app *application) getFacilitatorHandler(w http.ResponseWriter, r *http.Request) {
    params := httprouter.ParamsFromContext(r.Context())
    id := params.ByName("id")

    facilitator, err := app.models.Facilitators.Get(id)
    if err != nil {
        switch {
        case errors.Is(err, data.ErrRecordNotFound):
            app.notFoundResponse(w, r)
        default:
            app.serverErrorResponse(w, r, err)
        }
        return
    }

    err = app.writeJSON(w, http.StatusOK, envelope{"facilitator": facilitator}, nil)
    if err != nil {
        app.serverErrorResponse(w, r, err)
    }
}

func (app *application) updateFacilitatorHandler(w http.ResponseWriter, r *http.Request) {
    params := httprouter.ParamsFromContext(r.Context())
    id := params.ByName("id")

    facilitator, err := app.models.Facilitators.Get(id)
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
        FirstName *string `json:"first_name"`
        LastName  *string `json:"last_name"`
        Notes     *string `json:"notes"`
    }

    err = app.readJSON(w, r, &input)
    if err != nil {
        app.badRequestResponse(w, r, err)
        return
    }

    if input.FirstName != nil {
        facilitator.FirstName = *input.FirstName
    }
    if input.LastName != nil {
        facilitator.LastName = *input.LastName
    }
    if input.Notes != nil {
        facilitator.Notes = input.Notes
    }

    v := validator.New()
    if data.ValidateFacilitator(v, facilitator); !v.Valid() {
        app.failedValidationResponse(w, r, v.Errors)
        return
    }

    err = app.models.Facilitators.Update(facilitator)
    if err != nil {
        switch {
        case errors.Is(err, data.ErrEditConflict):
            app.editConflictResponse(w, r)
        default:
            app.serverErrorResponse(w, r, err)
        }
        return
    }

    err = app.writeJSON(w, http.StatusOK, envelope{"facilitator": facilitator}, nil)
    if err != nil {
        app.serverErrorResponse(w, r, err)
    }
}

func (app *application) deleteFacilitatorHandler(w http.ResponseWriter, r *http.Request) {
    params := httprouter.ParamsFromContext(r.Context())
    id := params.ByName("id")

    err := app.models.Facilitators.Delete(id)
    if err != nil {
        switch {
        case errors.Is(err, data.ErrRecordNotFound):
            app.notFoundResponse(w, r)
        default:
            app.serverErrorResponse(w, r, err)
        }
        return
    }

    err = app.writeJSON(w, http.StatusOK, envelope{"message": "facilitator successfully deleted"}, nil)
    if err != nil {
        app.serverErrorResponse(w, r, err)
    }
}

func (app *application) listFacilitatorsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		FirstName string
		LastName  string
		data.Filters
	}

	v := validator.New()
	qs := r.URL.Query()

	input.FirstName = app.readString(qs, "first_name", "")
	input.LastName = app.readString(qs, "last_name", "")

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "first_name", "last_name", "-id", "-first_name", "-last_name"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	facilitators, metadata, err := app.models.Facilitators.GetAll(input.FirstName, input.LastName, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"facilitators": facilitators, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}