package main

import (
	"context"
	"expvar"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/toduluz/savingsquadsbackend/internal/data"
	"github.com/toduluz/savingsquadsbackend/internal/jsonlog"
	"github.com/toduluz/savingsquadsbackend/internal/mailer"
	"github.com/toduluz/savingsquadsbackend/internal/vcs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Set version of application corresponding to value of vcs.Version.
var (
	version = vcs.Version()
)

// Define a config struct.
type config struct {
	port int
	env  string
	// Add a new jwt struct containing a single string field for the JWT signing secret.
	jwt struct {
		secret string
	}
	// db struct field holds the configuration settings for our database connection pool.
	db struct {
		maxOpenConns     int
		maxIdleTime      string
		connectionString string
		databaseName     string
	}
	// Add a new limiter struct containing fields for the request-per-second and burst
	// values, and a boolean field which we can use to enable/disable rate limiting.
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
	cors struct {
		trustedOrigins []string
	}
}

// Define an application struct to hold dependencies for our HTTP handlers, helpers, and
// middleware.
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func main() {
	// Declare an instance of the config struct.
	var cfg config

	// Initialize a new jsonlog.Logger which writes any messages *at or above* the INFO
	// severity level to the standard out stream.
	logger := jsonlog.NewLogger(os.Stdout, jsonlog.LevelInfo)

	err := godotenv.Load()
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	// Read the value of the port and env command-line flags into the config struct.
	// We default to using the port number 4000 and the environment "development" if no
	// corresponding flags are provided.
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production")

	mongoConnectionString := os.Getenv("MONGOURI")
	// Read the connection pool settings from command-line flags into the config struct.
	// Notice the default values that we're using?
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25,
		"PostgreSQL max open connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m",
		"PostgreSQL max connection idle time")
	flag.StringVar(&cfg.db.connectionString, "db-connection-string", mongoConnectionString, "MongoDB connection string")
	flag.StringVar(&cfg.db.databaseName, "db-database-name", "testDB", "MongoDB database name")

	// Read the limiter settings from the command-line flags into the config struct.
	// We use true as the default for 'enabled' setting.
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	// Read the SMTP server configuration settings into the config struct, using the
	// Mailtrap settings as teh default values.
	mtUser := os.Getenv("MAILTRAP_USER")
	mtPw := os.Getenv("MAILTRAP_PW")
	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 2525, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", mtUser, "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", mtPw, "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "DoNotReply <3fc3f54366-09689f+1@inbox.mailtrap.io>", "SMTP sender")

	// Use flag.Func function to process the -cors-trusted-origins command line flag. In this we
	// use the strings.Field function to split the flag value into slice based on whitespace
	// characters and assign it to our config struct. Importantly, if the -cors-trusted-origins
	// flag is not present, contains the empty string, or contains only whitespace, then
	// strings.Fields will return an empty []string slice.
	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	// Parse the JWT signing secret from the command-line-flag. Notice that we leave the
	// default value as the empty string if no flag is provided.
	flag.StringVar(&cfg.jwt.secret, "jwt-secret", "", "JWT secret")
	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	// If the version flag value is true, then print out the version number and immediately exit.
	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		os.Exit(0)
	}

	// Call the openDB() helper function (see below) to create teh connection pool,
	// passing in the config struct. If this returns an error,
	// we log it and exit the application immediately.
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	// Defer a call to db.Close() so that the connection pool is closed before the main()
	// function exits.
	defer func() {
		if err := db.Client().Disconnect(context.Background()); err != nil {
			logger.PrintFatal(err, nil)
		}
	}()

	logger.PrintInfo("database connection pool established", nil)

	// Publish a new "version" varaible in the expar var handler containing our application
	// version number.
	expvar.NewString("version").Set(version)

	// Publish the number of activate goroutines.
	expvar.Publish("goroutines", expvar.Func(func() interface{} {
		return runtime.NumGoroutine()
	}))

	// Publish the database connection pool statistics.
	expvar.Publish("database", expvar.Func(func() interface{} {
		var result bson.M
		err := db.RunCommand(context.Background(), bson.D{{"dbStats", 1}}).Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		return result
	}))

	// Publish the current Unix timestamp.
	expvar.Publish("timestamp", expvar.Func(func() interface{} {
		return time.Now().Unix()
	}))

	// Declare an instance of the application struct, containing the config struct and the infoLog.
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	// Call app.server() to start the server.
	if err := app.serve(); err != nil {
		logger.PrintFatal(err, nil)
	}
}

func openDB(cfg config) (*mongo.Database, error) {
	// Set client options
	clientOptions := options.Client().ApplyURI(cfg.db.connectionString)
	clientOptions.SetMaxPoolSize(uint64(cfg.db.maxOpenConns)) // Set the maximum connection pool size

	maxConnectionIdleTime, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	clientOptions.SetMaxConnIdleTime(maxConnectionIdleTime) // Set the maximum connection idle time

	// Create a context with a 5-second timeout deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Connect to the MongoDB server with context
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Call Ping to check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Return a handle to the specified database
	return client.Database(cfg.db.databaseName), nil
}
