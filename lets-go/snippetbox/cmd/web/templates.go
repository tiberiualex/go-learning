package main

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"time"

	"tiberiualex-golearning-snippetbox/internal/models"
	"tiberiualex-golearning-snippetbox/ui"
)

// Define a templateData type to act as a holding structure for
// any dynamic data that we want to pass to our HTML templates.
// We need this because `html/template` package allows us to pass
// in one and only one item of dynamic data when rendering a template.
// So we're creating a struct to hold all the data in different fields
type templateData struct {
	CurrentYear     int
	Snippet         *models.Snippet
	Snippets        []*models.Snippet
	Form            any
	Flash           string
	IsAuthenticated bool
	// We need the nosurf.Token() function to get the CSRF token, then add it to a hidden field
	// so our forms will still work
	CSRFToken string
}

// Create a humanData function which returns a nicely formatted string
// representation of a time.Time object
func humanDate(t time.Time) string {
	// Return the empty string if time has the zero value.
	if t.IsZero() {
		return ""
	}

	return t.UTC().Format("02 Jan 2006 at 15:04")
}

// Initialize a template.FuncMap object and store it in a global variable. This is
// essentially a string-keyed map which acts as a lookup between the names of our
// custom template functions and the functions themselves.
var functions = template.FuncMap{
	"humanDate": humanDate,
}

func newTemplateCache() (map[string]*template.Template, error) {
	// initialize a new map to act as the cache
	cache := map[string]*template.Template{}

	// Use the filepath.Glob() function to get a slice of all filepaths that
	// match the pattern "./ui/html/pages/*.tmpl". This will essentially give
	// us a slice of all the filepaths for our application 'page' templates
	// like [ui/html/pages/home.tmpl ui/html/pages/view/tmpl]
	// old code pages, err := filepath.Glob("./ui/html/pages/*.tmpl")

	// Use the fs.Glob() to get a slice of all filepaths in the ui.Files embedded
	// filesystem which match the pattern 'html/pages/*tmpl'. This essentially
	// gives us a slice of all the 'page' templates for the application, just
	// like before
	pages, err := fs.Glob(ui.Files, "html/pages/*.tmpl")
	if err != nil {
		return nil, err
	}

	// Loop through the page filepaths one-by-one.
	for _, page := range pages {
		// Extract the file name (like 'home.tmpl') from the full filepath
		// and assign it to the name variable
		name := filepath.Base(page)

		// Create a slice containing the filepath patterns for the templates we
		// want to parse.
		patterns := []string{
			"html/base.tmpl",
			"html/partials/*.tmpl",
			page,
		}

		// Use ParseFS() instead of ParseFiles() to parse the template files
		// from the ui.files embedded filesystem
		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		// Add the template set to the map, using the name of the page
		// (like 'home.tmpl' as the key)
		cache[name] = ts
	}

	// Return the map.
	return cache, nil
}
