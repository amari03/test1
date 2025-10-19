//HANDLER FOR USERS ENDPOINT
package main

import (
    "errors"
    "fmt"
    "net/http"
    "time"
    

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
        Role:  input.Role,
        Activated: false, //users not activiated by default
    }

    // Use the Set() method to hash the password.
    err = user.Password.Set(input.Password)
    if err != nil {
        app.serverErrorResponse(w, r, err)
        return
    }

    v := validator.New()
    if data.ValidateUser(v, user); !v.Valid() {
        app.failedValidationResponse(w, r, v.Errors)
        return
    }

    err = app.models.Users.Insert(user)
    if err != nil {
        switch {
        // If we get a duplicate email error, send a specific validation error.
        case errors.Is(err, data.ErrDuplicateEmail):
            v.AddError("email", "a user with this email address already exists")
            app.failedValidationResponse(w, r, v.Errors)
        default:
            app.serverErrorResponse(w, r, err)
        }
        return
    }
    
    // Generate a new activation token for the user.
    token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
    if err != nil {
        app.serverErrorResponse(w, r, err)
        return
    }

     // Send the welcome email with the activation token.
    app.background(func() {
        // Create a map to hold the data for the template.
        emailData := map[string]interface{}{
            "activationToken": token.Plaintext,
            "userID":          user.ID,
        }

        err = app.mailer.Send(user.Email, "user_welcome.tmpl", emailData)
        if err != nil {
            app.logger.Error(err.Error())
        }
    })


    // Per the slides, we'll want to send a welcome email here.
    // We will add that in the next phase.

    headers := make(http.Header)
    headers.Set("Location", fmt.Sprintf("/v1/users/%s", user.ID))

    // Return a 201 Created response with the user data in the body.
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
        if errors.Is(err, data.ErrRecordNotFound) {
            app.notFoundResponse(w, r)
            return
        }
        app.serverErrorResponse(w, r, err)
        return
    }

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
    
    err = app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
    if err != nil {
        app.serverErrorResponse(w, r, err)
    }
}

func (app *application) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
        params := httprouter.ParamsFromContext(r.Context())
        id := params.ByName("id")

        err := app.models.Users.Delete(id)
        if err != nil {
            switch {
            case errors.Is(err, data.ErrRecordNotFound):
                app.notFoundResponse(w, r)
            default:
                app.serverErrorResponse(w, r, err)
            }
            return
        }

        err = app.writeJSON(w, http.StatusOK, envelope{"message": "user successfully deleted"}, nil)
        if err != nil {
            app.serverErrorResponse(w, r, err)
        }
    }

func (app *application) listUsersHandler(w http.ResponseWriter, r *http.Request) {
    users, err := app.models.Users.GetAll()
    if err != nil {
        app.serverErrorResponse(w, r, err)
        return
    }

    err = app.writeJSON(w, http.StatusOK, envelope{"users": users}, nil)
    if err != nil {
        app.serverErrorResponse(w, r, err)
    }
}

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
    var input struct {
        TokenPlaintext string `json:"token"`
    }

    err := app.readJSON(w, r, &input)
    if err != nil {
        app.badRequestResponse(w, r, err)
        return
    }

    v := validator.New()
    if data.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
        app.failedValidationResponse(w, r, v.Errors)
        return
    }

    user, err := app.models.Users.GetForToken(data.ScopeActivation, input.TokenPlaintext)
    if err != nil {
        switch {
        case errors.Is(err, data.ErrRecordNotFound):
            v.AddError("token", "invalid or expired activation token")
            app.failedValidationResponse(w, r, v.Errors)
        default:
            app.serverErrorResponse(w, r, err)
        }
        return
    }

    user.Activated = true

    err = app.models.Users.Update(user)
    if err != nil {
        switch {
        case errors.Is(err, data.ErrEditConflict):
            app.editConflictResponse(w, r) // We will create this error response next
        default:
            app.serverErrorResponse(w, r, err)
        }
        return
    }

    err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
    if err != nil {
        app.serverErrorResponse(w, r, err)
        return
    }

    err = app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
    if err != nil {
        app.serverErrorResponse(w, r, err)
    }
}