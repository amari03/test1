//HANDLER FOR COURSES ENDPOINT
package main

import (
	"fmt"
    "net/http"
	"errors"

	"github.com/julienschmidt/httprouter"
    "github.com/amari03/test1/internal/data"
    "github.com/amari03/test1/internal/validator"
)

// createCourseHandler will handle POST /v1/courses
// createCourseHandler remains good. Note the hardcoded user ID is for development.
func (app *application) createCourseHandler(w http.ResponseWriter, r *http.Request) {

    user := app.contextGetUser(r)

    var input struct {
        Title              string  `json:"title"`
        Category           string  `json:"category"`
        DefaultCreditHours float64 `json:"default_credit_hours"`
        Description        string  `json:"description"`
    }

    err := app.readJSON(w, r, &input)
    if err != nil {
        app.badRequestResponse(w, r, err)
        return
    }

    // For now, we'll hardcode the user ID. Later, this will come from the logged-in user.
    course := &data.Course{
        Title:              input.Title,
        Category:           input.Category,
        DefaultCreditHours: input.DefaultCreditHours,
        Description:        input.Description,
        CreatedByUserID:    user.ID,
    }

    v := validator.New()
    if data.ValidateCourse(v, course); !v.Valid() {
        app.failedValidationResponse(w, r, v.Errors)
        return
    }

    err = app.models.Courses.Insert(course)
    if err != nil {
        app.serverErrorResponse(w, r, err)
        return
    }

    headers := make(http.Header)
    headers.Set("Location", fmt.Sprintf("/v1/courses/%s", course.ID))

    err = app.writeJSON(w, http.StatusCreated, envelope{"course": course}, headers)
    if err != nil {
        app.serverErrorResponse(w, r, err)
    }
}

// getCourseHandler will handle GET /v1/courses/:id
func (app *application) getCourseHandler(w http.ResponseWriter, r *http.Request) {
    params := httprouter.ParamsFromContext(r.Context())
    id := params.ByName("id")

    course, err := app.models.Courses.Get(id)
    if err != nil {
        switch {
        case errors.Is(err, data.ErrRecordNotFound):
            app.notFoundResponse(w, r)
        default:
            app.serverErrorResponse(w, r, err)
        }
        return
    }

    err = app.writeJSON(w, http.StatusOK, envelope{"course": course}, nil)
    if err != nil {
        app.serverErrorResponse(w, r, err)
    }
}


func (app *application) updateCourseHandler(w http.ResponseWriter, r *http.Request) {
    params := httprouter.ParamsFromContext(r.Context())
    id := params.ByName("id")

    course, err := app.models.Courses.Get(id)
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
        Title              *string  `json:"title"`
        Category           *string  `json:"category"`
        DefaultCreditHours *float64 `json:"default_credit_hours"`
        Description        *string  `json:"description"`
    }

    err = app.readJSON(w, r, &input)
    if err != nil {
        app.badRequestResponse(w, r, err)
        return
    }

    if input.Title != nil { course.Title = *input.Title }
    if input.Category != nil { course.Category = *input.Category }
    if input.DefaultCreditHours != nil { course.DefaultCreditHours = *input.DefaultCreditHours }
    if input.Description != nil { course.Description = *input.Description }

    v := validator.New()
    if data.ValidateCourse(v, course); !v.Valid() {
        app.failedValidationResponse(w, r, v.Errors)
        return
    }

    err = app.models.Courses.Update(course)
    if err != nil {
        switch {
        case errors.Is(err, data.ErrEditConflict):
            app.editConflictResponse(w, r)
        default:
            app.serverErrorResponse(w, r, err)
        }
        return
    }

    err = app.writeJSON(w, http.StatusOK, envelope{"course": course}, nil)
    if err != nil {
        app.serverErrorResponse(w, r, err)
    }
}

func (app *application) deleteCourseHandler(w http.ResponseWriter, r *http.Request) {
    params := httprouter.ParamsFromContext(r.Context())
    id := params.ByName("id")

    err := app.models.Courses.Delete(id)
    if err != nil {
        switch {
        case errors.Is(err, data.ErrRecordNotFound):
            app.notFoundResponse(w, r)
        default:
            app.serverErrorResponse(w, r, err)
        }
        return
    }

    err = app.writeJSON(w, http.StatusOK, envelope{"message": "course successfully deleted"}, nil)
    if err != nil {
        app.serverErrorResponse(w, r, err)
    }
}

func (app *application) listCoursesHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title    string
		Category string
		data.Filters
	}

	v := validator.New()
	qs := r.URL.Query()

	input.Title = app.readString(qs, "title", "")
	input.Category = app.readString(qs, "category", "")

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")

	// Define the safelist of fields that can be sorted on.
	input.Filters.SortSafelist = []string{"id", "title", "category", "default_credit_hours", "-id", "-title", "-category", "-default_credit_hours"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	courses, metadata, err := app.models.Courses.GetAll(input.Title, input.Category, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"courses": courses, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}