package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
	"strings"
	"errors"

	"golang.org/x/time/rate"
	"github.com/amari03/test1/internal/data"
	"github.com/amari03/test1/internal/validator"
)

// recoverPanic is middleware that recovers from panics and returns a 500 error.
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// rateLimit is middleware for IP-based rate limiting.
func (app *application) rateLimit(next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		mu.Lock()
		if _, found := clients[ip]; !found {
			clients[ip] = &client{limiter: rate.NewLimiter(2, 4)} // 2 requests/sec, burst of 4
		}
		clients[ip].lastSeen = time.Now()

		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			app.rateLimitExceededResponse(w, r)
			return
		}
		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded"
	app.errorResponse(w, r, http.StatusTooManyRequests, message)
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add the "Vary: Authorization" header. This indicates to caches that the
		// response may vary based on the value of the Authorization header.
		w.Header().Add("Vary", "Authorization")

		// 1. Get the value of the Authorization header.
		authorizationHeader := r.Header.Get("Authorization")

		// 2. If the header is empty, treat the user as anonymous.
		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		// 3. Check if the header is in the "Bearer <token>" format.
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}
		token := headerParts[1]

		// 4. Validate the token.
		v := validator.New()
		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// 5. Retrieve the user associated with the token.
		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		// 6. Add the user information to the request context.
		r = app.contextSetUser(r, user)
		next.ServeHTTP(w, r)
	})
}

// requireAuthenticatedUser checks if the user is authenticated (i.e., not anonymous).
func (app *application) requireAuthenticatedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// Use contextGetUser to retrieve the user from the request context.
	user := app.contextGetUser(r)
	// If the user is anonymous, call the authenticationRequiredResponse helper.
    	if user.IsAnonymous() {
        	app.authenticationRequiredResponse(w, r)
        	return
    	}

    // Otherwise, they are authenticated; call the next handler.
    next.ServeHTTP(w, r)
	})
}

// requireActivatedUser checks if the user is both authenticated and activated.
func (app *application) requireActivatedUser(next http.Handler) http.Handler {
	// This handler will first check if the user is activated.
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	// If the user is not activated, send an inactiveAccountResponse.
    	if !user.Activated {
        	app.inactiveAccountResponse(w, r)
        	return
    	}

    next.ServeHTTP(w, r)
	})

	// Wrap the activation check with the authentication check. This ensures
	// we only check for activation if the user is not anonymous.
	return app.requireAuthenticatedUser(fn)
}

func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add the "Vary: Origin" header.
		w.Header().Add("Vary", "Origin")

		// Add the "Vary: Access-Control-Request-Method" header for preflight requests.
		w.Header().Add("Vary", "Access-Control-Request-Method")

		origin := r.Header.Get("Origin")

		// Check if we have a valid origin and if it's in our trusted list.
		if origin != "" && len(app.config.cors.trustedOrigins) > 0 {
			for i := range app.config.cors.trustedOrigins {
				if origin == app.config.cors.trustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)

					// Check if this is a preflight (OPTIONS) request.
					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						// Set the necessary headers for the preflight response.
						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

						// Send a 200 OK status and terminate the middleware chain.
						w.WriteHeader(http.StatusOK)
						return
					}
				}
			}
		}

		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}