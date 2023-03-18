package data

import "time"

// All the fields are exported (start with a capital leter), which is necessary
// for them to be visible to Go's encoding/json package. Any fields which aren't
// exported won't be included when encoding a struct to JSON
// Use struct tags to control how the keys appear in the JSON-encoded output
type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"` // Omit this from the JSON output
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"` // Omit if the field is empty (false, 0, "",empty array/map, nil)
	// Use the Runtime type instead of int32. Note that the omitempty directive will
	// still work on this: if the Runtime field has the underlying value 0, then it will
	// be considered empty and omitted -- and the MarshalJSON() method we just made
	// won't be called at all
	Runtime Runtime  `json:"runtime,omitempty,string"` // The string directive will force the field to be converted to string in the JSON output
	Genres  []string `json:"genres,omitempty"`
	Version int32    `json:"version"`
}
