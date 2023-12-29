package api

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Define a config struct.
type Config struct {
	Port int
	Env  string
	// Add a new jwt struct containing a single string field for the JWT signing secret.
	Jwt struct {
		Secret string
	}
	// db struct field holds the configuration settings for our database connection pool.
	Db struct {
		MaxOpenConns     int
		MaxIdleTime      string
		ConnectionString string
		DatabaseName     string
	}
	Cors struct {
		TrustedOrigins []string
	}
}

func OpenDB(cfg Config) (*mongo.Database, error) {
	// Set client options
	clientOptions := options.Client().ApplyURI(cfg.Db.ConnectionString)
	clientOptions.SetMaxPoolSize(uint64(cfg.Db.MaxOpenConns)) // Set the maximum connection pool size

	maxConnectionIdleTime, err := time.ParseDuration(cfg.Db.MaxIdleTime)
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
	return client.Database(cfg.Db.DatabaseName), nil
}
