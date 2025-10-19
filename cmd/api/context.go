package main

import (
	"context"
	"net/http"

	"github.com/amari03/test1/internal/data"
)

// Define a custom contextKey type. This is used to prevent collisions
// with keys from other packages.
type contextKey string

// userContextKey is the key we'll use to store the User struct in the context.
const userContextKey = contextKey("user")

// contextSetUser returns a new request with the provided User struct added to the context.
func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

// contextGetUser retrieves the User struct from the request context.
// It will panic if the key is not in the context, as this indicates a developer error
// (calling it on a route that isn't protected by the authenticate middleware).
func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}