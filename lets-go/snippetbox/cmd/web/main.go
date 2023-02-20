package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"tiberiualex-golearning-snippetbox/internal/models"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"

	// we're not using anything from the package, but we need its init() function
	// to register itself with the database/sql package, so we alias it to the blank identifier
	// so the compiler won't complain about unused imports
	_ "github.com/go-sql-driver/mysql"
)

// go run ./cmd/web

type application struct {
	errorLog       *log.Logger
	infoLog        *log.Logger
	snippets       *models.SnippetModel
	templateCache  map[string]*template.Template
	sessionManager *scs.SessionManager
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

	// Use the scs.New() function to initialize a new session manager. Then we
	// configure it to use our MySQL database as the session store, and set a
	// lifetime of 12 hours (so that the sessions automatically expire 12 hours
	// after first being created).
	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	// Make sure that the Secure attribute is set on our session cookies.
	// Setting this means that the cookie will only be sent by a user's web
	// browser when a HTTPS connection is being used (and won't be sent over an
	// unsecure HTTP connection).
	sessionManager.Cookie.Secure = true

	// initialize a new instance of our application struct, containing
	// the dependencies
	app := &application{
		errorLog:       errorLog,
		infoLog:        infoLog,
		snippets:       &models.SnippetModel{DB: db},
		templateCache:  templateCache,
		sessionManager: sessionManager,
	}

	// Initialize a tls.Config struct to hold the non-default TLS settings we
	// want the server to use. In this case, the only thing that we're changing
	// is the curve preferences value, so that only the elliptic curves with
	// assembly implementations are used.
	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	// Initialize a new http.Server struct. We set the Addr and Handler fields so
	// that the serer uses the same network address and routes as before, and set
	// the ErrorLog field so that the server now uses the custom errorLog logger in
	// the even of any problems
	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
		// Set the server's TLS config
		TLSConfig: tlsConfig,
		// Add Idle, Read and Write timeouts to the server
		// By default, Go enables keep-alives on all accepted connections, this helps
		// reduce latency (especially for HTTPS) because a client can reuse the same
		// connection for multiple requests without having to repeat the handshake.
		// By default, keep-alive connections will be automatically closed after a
		// couple of minutes. There's no way to increase this default with the default
		// go net.Listener, but you can reduce it via the IddleTimeout setting. In our
		// case, all keep-alives will be closed after 1 minute of inactivity
		IdleTimeout: time.Minute,
		// Setting this can avoid slow-client attacks. After the timeout, Go will close the
		// underlying connection, so the user wont' receive any HTTP(S) response
		// Also, if you set this, but not IdleTimeout, IddleTimeout will default to the
		// same setting as ReadTimeout, so make sure to set IdleTimeout explicitly
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Write messages using the two new loggers, instead of the standard logger
	infoLog.Printf("Starting server on %s", *addr)
	// Call the ListenAndServe() method on our new http.Server struct
	// Later: use the ListenAndServeTLS() method to start the HTTPS server. We
	// pass in the paths to the TLS certificate and corresponding private key as
	// the two parameters
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
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
