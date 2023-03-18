package main

import (
	"net/http"
)

// Declare a handler which writes a plain-text response with information about the
// application status, operating environment and version
// Note how this is implemented as a "method" on the application struct
// This is the idiomatic way of making dependencies available to handlers as they
// can simply be fields on the application struct, which the handlers have access to
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an envelope map containing the data for the response. Notice the way
	// we've constructed this means the environment and version data will now be nested
	// under a system_info key in the JSON response
	env := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}

	err := app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.logger.Println(err)
		app.serverErrorResponse(w, r, err)
	}
}
