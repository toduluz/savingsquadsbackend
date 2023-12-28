package data

import (
	"context"
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func newTestDB(t *testing.T) (*mongo.Database, func()) {
	t.Helper()

	// Get the MongoDB URI from the environment variable.
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		t.Fatal("MONGO_URI not set")
	}

	// Connect to the MongoDB server.
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		t.Fatal(err)
	}

	// Create a new database.
	db := client.Database("test")

	// Return the database handle and a function to close it.
	return db, func() {
		err := db.Drop(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	}
}
