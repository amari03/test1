//HANDLER FOR COURSES ENDPOINT
package main

import (
	"net/http"
)

// createCourseHandler will handle POST /v1/courses
func (app *application) createCourseHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement logic to create a new course.
	w.Write([]byte("TODO: Create a new course"))
}

// getCourseHandler will handle GET /v1/courses/:id
func (app *application) getCourseHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement logic to get a specific course by ID.
	w.Write([]byte("TODO: Get course by ID"))
}