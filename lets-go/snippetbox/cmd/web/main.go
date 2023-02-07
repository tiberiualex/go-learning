package main

import (
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"

	"tiberiualex-golearning-snippetbox/internal/models"

	// we're not using anything from the package, but we need its init() function
	// to register itself with the database/sql package, so we alias it to the blank identifier
	// so the compiler won't complain about unused imports
	_ "github.com/go-sql-driver/mysql"
)

type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	snippets      *models.SnippetModel
	templateCache map[string]*template.Template
}

func main() {
	// Define a new command line flag, with a default value of ":4000"
	addr := flag.String("addr", ":4000", "HTTP network address")

	// Define a new command-line flag for the MySQL DSN (Data Source Name: connection string) string.
	dsn := flag.String("dsn", "web:pass@/snippetbox?parseTime=true", "MySQL data source name")

	// This must be called to parse the flags. If you read addr before
	// parsing, it will always have its default value
	flag.Parse()

	// Use log.New() to create a logger for writing information messages. This takes
	// three parameters: the destimation to write the logs to (os.Stdout), a string
	// prefix for message (INFO followed by a tab), and flags to indicate what
	// additional information to include (local date and time). Note that the flags
	// are joined using the bitwise OR operator |.
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	// Create a logger for wirting error messages in the same way, but use the sdrerr as
	// the destination and use the log.Lshortfile flag to incldue the relevant
	// filename and line number
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// To keep the main() function tidy, we're creating the connection pool
	// in a separate openDB() function. We're passing the new function the
	// DSN (connection string) from the command-line flag
	db, err := openDB(*dsn)

	if err != nil {
		errorLog.Fatal(err)
	}

	// We also defer a call to db.Close() so that the connection pool is closed
	// before the main() function exits.
	defer db.Close()

	// Initialize a new template cache...
	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	// initialize a new instance of our application struct, containing
	// the dependencies
	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		snippets:      &models.SnippetModel{DB: db},
		templateCache: templateCache,
	}

	// Initialize a new http.Server struct. We set the Addr and Handler fields so
	// that the serer uses the same network address and routes as before, and set
	// the ErrorLog field so that the server now uses the custom errorLog logger in
	// the even of any problems
	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	// Write messages using the two new loggers, instead of the standard logger
	infoLog.Printf("Starting server on %s", *addr)
	// Call the ListenAndServe() method on our new http.Server struct
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

// The openDB() function wraps sql.Open() and returns a sql.DB connection pool
// for a given DSN (Data Source Name: connection string)
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)

	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
