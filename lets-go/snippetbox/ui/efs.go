package ui

import (
	"embed"
)

// The comment below is a coment directive instructing Go
// to store the files from ui/html and ui/static folders in
// an embed.FS embedded filesystem referenced by the global variable Files

//go:embed "html" "static"
var Files embed.FS
