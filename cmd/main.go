package main

import (
	"context"
	"flag"
	"os"
	"strings"

	"github.com/toduluz/savingsquadsbackend/api"
	"github.com/toduluz/savingsquadsbackend/internal/data"
	"github.com/toduluz/savingsquadsbackend/internal/jsonlog"
)

func main() {
	// Declare an instance of the config struct.
	var cfg api.Config

	// Initialize a new jsonlog.Logger which writes any messages *at or above* the INFO
	// severity level to the standard out stream.
	logger := jsonlog.NewLogger(os.Stdout, jsonlog.LevelInfo)

	// err := godotenv.Load()
	// if err != nil {
	// 	logger.PrintFatal(err, nil)
	// }

	// Read the value of the port and env command-line flags into the config struct.
	// We default to using the port number 4000 and the environment "development" if no
	// corresponding flags are provided.
	flag.IntVar(&cfg.Port, "port", 4000, "API server port")
	flag.StringVar(&cfg.Env, "env", "development", "Environment (development|staging|production")

	mongoConnectionString := os.Getenv("MONGOURILOCAL")
	// Read the connection pool settings from command-line flags into the config struct.
	flag.IntVar(&cfg.Db.MaxOpenConns, "db-max-open-conns", 25,
		"MongoDB max open connections")
	flag.StringVar(&cfg.Db.MaxIdleTime, "db-max-idle-time", "15m",
		"MongoDB max connection idle time")
	flag.StringVar(&cfg.Db.ConnectionString, "db-connection-string", mongoConnectionString, "MongoDB connection string")
	flag.StringVar(&cfg.Db.DatabaseName, "db-database-name", "testDB", "MongoDB database name")

	// Use flag.Func function to process the -cors-trusted-origins command line flag. In this we
	// use the strings.Field function to split the flag value into slice based on whitespace
	// characters and assign it to our config struct. Importantly, if the -cors-trusted-origins
	// flag is not present, contains the empty string, or contains only whitespace, then
	// strings.Fields will return an empty []string slice.
	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.Cors.TrustedOrigins = strings.Fields(val)
		return nil
	})

	jwtSecret := os.Getenv("JWTSECRET")
	// Parse the JWT signing secret from the command-line-flag. Notice that we leave the
	// default value as the empty string if no flag is provided.
	flag.StringVar(&cfg.Jwt.Secret, "jwt-secret", jwtSecret, "JWT secret")

	flag.Parse()

	// Call the openDB() helper function (see below) to create teh connection pool,
	// passing in the config struct. If this returns an error,
	// we log it and exit the Application immediately.
	db, err := api.OpenDB(cfg)
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

	// Declare an instance of the Application struct, containing the config struct and the infoLog.
	app := &api.Application{
		Config: cfg,
		Logger: logger,
		Models: data.NewModels(db),
	}

	// Call app.server() to start the server.
	if err := app.Serve(); err != nil {
		logger.PrintFatal(err, nil)
	}
}
