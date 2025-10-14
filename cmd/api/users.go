//HANDLER FOR USERS ENDPOINT
package main

import (
	//"errors"
	"fmt"
	"net/http"
	

	"github.com/amari03/test1/internal/data"
	"github.com/amari03/test1/internal/validator"
	"github.com/julienschmidt/httprouter"
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
        // For now, we're just assigning the password directly to the hash field.
        PasswordHash: input.Password,
        Role:  input.Role,
    }

    v := validator.New()
    if data.ValidateUser(v, user); !v.Valid() {
        app.failedValidationResponse(w, r, v.Errors)
        return
    }

    // This is the part we are now implementing
    err = app.models.Users.Insert(user)
    if err != nil {
        app.serverErrorResponse(w, r, err)
        return
    }

    headers := make(http.Header)
    headers.Set("Location", fmt.Sprintf("/v1/users/%s", user.ID))

    err = app.writeJSON(w, http.StatusCreated, envelope{"user": user}, headers)
    if err != nil {
        app.serverErrorResponse(w, r, err)
    }
}

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
    params := httprouter.ParamsFromContext(r.Context())
    id := params.ByName("id")

    user, err := app.models.Users.Get(id)
    if err != nil {
        if err == data.ErrRecordNotFound {
            app.notFoundResponse(w, r)
            return
        }
        app.serverErrorResponse(w, r, err)
        return
    }

    // We don't want to send the password hash back to the client.
    user.PasswordHash = ""

    err = app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
    if err != nil {
        app.serverErrorResponse(w, r, err)
    }
}

func (app *application) updateUserHandler(w http.ResponseWriter, r *http.Request) {
    params := httprouter.ParamsFromContext(r.Context())
    id := params.ByName("id")

    user, err := app.models.Users.Get(id)
    if err != nil {
        app.notFoundResponse(w, r)
        return
    }

    var input struct {
        Email *string `json:"email"`
        Role  *string `json:"role"`
    }

    err = app.readJSON(w, r, &input)
    if err != nil {
        app.badRequestResponse(w, r, err)
        return
    }

    if input.Email != nil {
        user.Email = *input.Email
    }
    if input.Role != nil {
        user.Role = *input.Role
    }

    v := validator.New()
    if data.ValidateUser(v, user); !v.Valid() {
        app.failedValidationResponse(w, r, v.Errors)
        return
    }

    err = app.models.Users.Update(user)
    if err != nil {
        app.serverErrorResponse(w, r, err)
        return
    }
    
    // We don't want to send the password hash back to the client.
    user.PasswordHash = ""

    err = app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
    if err != nil {
        app.serverErrorResponse(w, r, err)
    }
}