package main

import "tiberiualex-golearning-snippetbox/internal/models"

// Define a templateData type to act as a holding structure for
// any dynamic data that we want to pass to our HTML templates.
// We need this because `html/template` package allows us to pass
// in one and only one item of dynamic data when rendering a template.
// So we're creating a struct to hold all the data in different fields
type templateData struct {
	Snippet *models.Snippet
}
