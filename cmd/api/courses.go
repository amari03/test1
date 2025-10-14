//HANDLER FOR COURSES ENDPOINT
package main

import (
	"fmt"
    "net/http"
    "github.com/amari03/test1/internal/data"
    "github.com/amari03/test1/internal/validator"
)

// createCourseHandler will handle POST /v1/courses
func (app *application) createCourseHandler(w http.ResponseWriter, r *http.Request) {
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
        CreatedByUserID:    "1a7a5180-4303-4318-8789-1a007f339eec", // Our dummy user
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
	// TODO: Implement logic to get a specific course by ID.
	w.Write([]byte("TODO: Get course by ID"))
}