//HANDLER FOR HEALTHCHECK ENDPOINT
package main

import (
	"net/http"
)

// healthcheckHandler writes a JSON response with status, environment, and version information.
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	// Create a map to hold the healthcheck data.
	// Using the envelope type provides a consistent response structure.
	data := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     appVersion, // This constant is defined in main.go
		},
	}

	// Use the writeJSON helper to send the response.
	// This ensures consistent JSON formatting and error handling.
	err := app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		// If writeJSON fails, log the error and send a generic
		// server error response to the client.
		app.serverErrorResponse(w, r, err)
	}
}