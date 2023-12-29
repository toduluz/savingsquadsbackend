package test

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/toduluz/savingsquadsbackend/api"
	"github.com/toduluz/savingsquadsbackend/internal/data"
	"github.com/toduluz/savingsquadsbackend/internal/jsonlog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func newIntegrationTestApplication(t *testing.T) (*api.Application, func()) {
	t.Helper()

	var cfg api.Config

	cfg.Jwt.Secret = "secret"

	// Get the MongoDB URI from the environment variable.
	mongoURI := os.Getenv("MONGOURILOCAL")
	if mongoURI == "" {
		t.Fatal("MONGOURILOCAL not set")
	}

	// Connect to the MongoDB server.
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		t.Fatal(err)
	}

	// Create a new database.
	db := client.Database("test")

	// Declare an instance of the Application struct, containing the Config struct and the infoLog.
	app := &api.Application{
		Config: cfg,
		Logger: jsonlog.NewLogger(io.Discard, jsonlog.LevelOff),
		Models: data.NewModels(db),
	}

	return app, func() {
		err := db.Drop(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		err = client.Disconnect(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	}
}
