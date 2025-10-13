//HANDLER FOR USERS ENDPOINT
package main

import (
	//"errors"
	"fmt"
	"net/http"

	"github.com/amari03/test1/internal/data"
	"github.com/amari03/test1/internal/validator"
)

// createUserHandler handles the creation of a new user.
func (app *application) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Email: input.Email,
		Role:  input.Role,
	}

	// TODO: Hash the password before storing it.
	// For now, we'll skip this for brevity.

	v := validator.New()
	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// TODO: Implement the userModel.Insert() method in internal/data/users.go
	// err = app.models.Users.Insert(user)
	// if err != nil {
	// 	app.serverErrorResponse(w, r, err)
	// 	return
	// }

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/users/%s", user.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"user": user}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// getUserHandler handles fetching a specific user by ID.
func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	// For this example, we assume ID is a string (UUID)
	// You might need to adjust readIDParam if you stick with integer IDs for some models
	// params := httprouter.ParamsFromContext(r.Context())
	// id := params.ByName("id")

	// For demonstration, let's pretend we have a user.
	// You would replace this with a call to your data model.
	// user, err := app.models.Users.Get(id)
	user := data.User{ID: "some-uuid", Email: "test@example.com", Role: "viewer"} // Dummy data
	err := app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}