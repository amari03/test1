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

	// Fetch the existing record.
	officer, err := app.models.Officers.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound): // Handle "not found" specifically
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// This input struct with pointers is perfect for partial updates. No changes needed here.
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

	// Apply the updates from the input struct to the fetched officer record.
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

	// Re-validate the updated officer record.
	v := validator.New()
	if data.ValidateOfficer(v, officer); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Pass the updated officer record to the Update() method.
	err = app.models.Officers.Update(officer)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict): // Handle edit conflicts specifically
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Send the updated officer record back to the client.
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

	officer, err := app.models.Officers.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"officer": officer}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listOfficersHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		FirstName string
		LastName  string
		RankCode  string
		data.Filters
	}

	v := validator.New()
	qs := r.URL.Query()

	// Use our new helpers to extract the query parameters.
	input.FirstName = app.readString(qs, "first_name", "")
	input.LastName = app.readString(qs, "last_name", "")
	input.RankCode = app.readString(qs, "rank_code", "")

	// Read pagination and sorting parameters.
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")

	// Define the safelist for sorting.
	input.Filters.SortSafelist = []string{"id", "first_name", "last_name", "rank_code", "-id", "-first_name", "-last_name", "-rank_code"}

	// Validate the filter values.
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Call the model method.
	officers, metadata, err := app.models.Officers.GetAll(input.FirstName, input.LastName, input.RankCode, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send the JSON response with the data and metadata.
	err = app.writeJSON(w, http.StatusOK, envelope{"officers": officers, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

