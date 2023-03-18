package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com.go-learning.greenlight/internal/data"
	"github.com.go-learning.greenlight/internal/validator"
)

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an anonymous struct to hold the information that we expect to be in the
	// HTTP request body (note that the field names and types in the struct are a subset
	// of the Movie struct that we created earlier). This struct will be our target
	// decode destination
	// When the target decode destination is a struct, the struct fields must be exported
	// (start with a capital letter) so they're visible to the encoding/json package
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	// Initialize a new json.Decoder instance which reads from the request body, and
	// then use the Decode() method to decode the body contents into the input struct
	// Importantly, notice that when we call Decode() we pass a *pointer* to the input
	// struct as the target decode destination. If there was an error during decoding,
	// we also use our generic errorResponse() helper to send the client a 400 Bad
	// Request containing the error message
	// There is no need to close r.Body after it has been read. This will be done
	// automatically by Go’s http.Server
	// err := json.NewDecoder(r.Body).Decode(&input)

	// Use the new readJSON() helper to decode the request body into the input struct.
	// If this returns an error, we send the client the error message along with a 400
	// Bad Request status code, just like before
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Copy the values from the input struct to a new Movie struct.
	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	// Initialize a new Validator.
	v := validator.New()

	// Call the ValidateMovie() function and return a response containing the errors if // any of the checks fail.

	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	fmt.Fprintf(w, "%+v\n", input)
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	movie := data.Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "Casablanca",
		Runtime:   102,
		Genres:    []string{"drama", "romance", "war"},
		Version:   1,
	}

	// Create an envelope{"movie": movie} instance and pass it to writeJSON() instead
	// of passing the plain movie struct
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.logger.Println(err)
		app.serverErrorResponse(w, r, err)
	}
}
