package data

import (
	"database/sql"
	"errors"
	"time"

	"github.com.go-learning.greenlight/internal/validator"
	"github.com/lib/pq"
)

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

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}

// Define MovieModel struct type which wraps a sql.DB connection pool.
type MovieModel struct {
	DB *sql.DB
}

// The Insert() method accepts a pointer to a movie struct, which should contain the
// data for the new record
func (m MovieModel) Insert(movie *Movie) error {
	// Define the SQL query for inserting a new record in the movies table and returning
	// the system-generated data
	query := `
		INSERT INTO movies (title, year, runtime, genres)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version`

	// Create an args slice containing the values for the placeholder parameters from
	// the movie struct. Declaring this slice immediately next to our SQL query helps to
	// make it nice and clear *what values are being used where* in the query
	// interface{} here is basically 'any' type
	// pq.Array convers the string slice we have to a pq.StringArray which the driver can
	// use to translate it to a Postgres text[] array column
	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	// Use the QueryRow() method to execute the SQL query on our connection pool,
	// passing in the args slice as a variadic parameter and scanning the system
	// generated id, created_at and version valeus into the movie struct
	// Calling Scan here using the movie pointer, we're mutating the movie struct that
	// we got as an argument
	return m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (m MovieModel) Get(id int64) (*Movie, error) {
	// The PosgreSQL bigserial type that we're using for the movie ID starts
	// auto-incremeting at 1 by default, so we know that no moveis will have ID values
	// less than that. To avoid making an unnecessary database call, we take a shortcut
	// and return ErrRecordNotFound error straight away
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	// Define the SQL query for retrieving the movie data.
	query := `
		SELECT id, created_at, title, year, runtime, genres, version
		FROM movies
		WHERE id=$1`

	// Declare a Movie struct to hold the data returned by the query.
	var movie Movie

	// Execute the query using the QueryRow() method, passing in the provided id value
	// as a placeholder paramter, and scan the response data into the fields of the
	// Movie struct. Importantly, notice that we need to convert the scan target for the
	// genres column using the pq.Array() adapter function again.
	err := m.DB.QueryRow(query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
	)

	// Handle any errors. If there was no matching movie found, Scan() will return
	// a sql.ErrNoRows error. We check for this and return our custom ErrRecordNotFonud
	// error instead
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	// Otherwise, return a pointer to the Movie struct.
	return &movie, nil
}

func (m MovieModel) Update(movie *Movie) error {
	return nil
}

func (m MovieModel) Delete(id int64) error {
	return nil
}
