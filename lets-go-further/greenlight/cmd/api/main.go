package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com.go-learning.greenlight/internal/data"
	"github.com.go-learning.greenlight/internal/jsonlog"
	"github.com.go-learning.greenlight/internal/mailer"

	// Import the pq driver so that it can register itself with the database/sql
	// package. Note that we alias this import to the blank identifier, to stop the Go
	// compiler complaining that the package isn't being used
	_ "github.com/lib/pq"
)

// Declare a string containing the application version number. Later we'll generate
// this automatically at build time, but for now we'll just store the version number
// as a hard-coded global constant
const version = "1.0.0"

// Define a config struct to hold all the configuration settings for our application.
// We will read in these configuration settings from command-line flags when the app
// starts
type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}

	// Add a new limiter struct containing fields for the requests-per-second and burst
	// values, and a boolean field which we can use to enable/disable rate limiting
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}

	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}

	// Add a cors struct and trustedOrigins field with the type []string
	cors struct {
		trustedOrigins []string
	}
}

// Define an application struct to hold the depedencies for our HTTP handlers, helpers,
// and middleware. At the moment this only contains a copy of the config struct and
// a logger, but it will grow to incldue a lot more as our build progresses
type application struct {
	config config
	logger *jsonlog.Logger
	// Add a models field to hold our new Models struct
	models data.Models
	// Update the application struct to hold a new Mailer instance
	mailer mailer.Mailer
	// Include a sync.WaitGroup in the application struct. The zero-value for a
	// sync.WaitGroup type is a valid, useable, sync.WaitGroup with a 'counter' value of 0,
	// so we don't need to do anything else to initialize it before we can use it
	wg sync.WaitGroup
}

func main() {
	var cfg config

	// Read the value of the port and env command-line flags into the config struct. We
	// default to using the port 4000 and the environment "development" if no corresponding
	// flags are provided
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	// Read the DSN value from the db-dsn command-line flag into the config struct. We
	// default to using our development DSN if no flag is provided.
	// Use the value of the GREENLIGHT_DB_DSN environment variable as the default value
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PosgreSQL DSN")

	// Read the connection pool settings from command-line flags into the config struct.
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PosgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL mex connection idle time")

	// Create command line flags to read the setting values into the config struct.
	// Notice that we use true as the default for the 'enabled' setting
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	// Read the SMTP server configuration settings into the config struct, using the
	// Mailtrap settings as the default values.
	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "random", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "random", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Greenlight <no-reply@tiberiualex.github.io>", "SMTP sender")

	// Use the flag.Func() function to process the -cors-trusted-origins command line
	// flag. In this we use the strings.FIelds() function to split the flag value into a
	// slice based on twhitespace characters and assign it to our config struct.
	// Importantly, if the -cors-trusted-origins flag is not present, contains the empty
	// string, or contains only whitespace, then string.Fields() will return an empty
	// []string slice
	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	flag.Parse()

	// Initialize a new logger which writes messages to the standard out stream,
	// prefixed with the current date and time.
	// logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// Initialize a new jsonlog.Logger which writes any messages *at or above* the INFO
	// severity level to the standard out stream
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	// Call the openDB() helper function to create the connection pool,
	// passing in the config struct. If this returns an error, we log it and exit the
	// application immediately
	db, err := openDB(cfg)
	if err != nil {
		// Use the PrintFatal() method to write a log entry containing the error at the
		// FATAL level and exit. We have no additional properties to include in the log
		// entry, so we pass nil as the second parameter
		logger.PrintFatal(err, nil)
	}

	// Defer a call to db.Close() so that the connection pool is closed before the
	// main() function exists.
	defer db.Close()

	// Also log a message to say that the connection pool has been successfully
	// established
	logger.PrintInfo("database connection pool established", nil)

	// Publish a new "version" variable in the expvar handler containing our application
	// version number (currently the constant "1.0.0")
	expvar.NewString("version").Set(version)

	// Publish the number of active goroutines
	expvar.Publish("goroutines", expvar.Func(func() interface{} {
		return runtime.NumGoroutine()
	}))

	// Publish the database connection pool statistics
	expvar.Publish("database", expvar.Func(func() interface{} {
		return db.Stats()
	}))

	// Publish the current Unix timestamp
	expvar.Publish("timestamp", expvar.Func(func() interface{} {
		return time.Now().Unix()
	}))

	// Declare an instance of the application struct, containing the config struct and
	// the logger
	app := &application{
		config: cfg,
		logger: logger,
		// Use the data.NewModels() function to initialize a Models struct, passing in the
		// connection pool as a parameter
		models: data.NewModels(db),
		// Initialize a new Mailer instance using the settings from the command line
		// flags, and add it to the application struct
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	err = app.serve()
	logger.PrintFatal(err, nil)
}

// Returns a sql.DB connection pool
func openDB(cfg config) (*sql.DB, error) {
	// Use the sql.Open() to create an empty connection pool, using the DSN from the config
	// struct
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// Set the maximum number of open (in-use + idle) connections in the pool. Note that
	// passing a value less than or equal to 0 will mean there is no limit.
	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	// Set the maximum number of idle conections in the pool. Again, passing a value
	// less than or equal to 0 will mean there is no limit.
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	// Use the time.ParseDuration() function to convert the idle timeout duration string
	// to a time.Duration type.
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	// Set the maximum idle timeout.
	db.SetConnMaxIdleTime(duration)

	// Create a context with a 5-second timeout deadline
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use PingContext() to establish a new connection to the database, passing it the
	// context we created above as a parameter. If the connection couldn't be
	// established successfully within the 5 second deadline, then this will return an
	// error
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	// Return the sql.DB connection pool
	return db, nil
}
