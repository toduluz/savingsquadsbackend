package main

import (
	"context"
	"log"
)

func main() {

	// establish connection with MongoDB
	client, err := connectMongoDB()
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	// running Server on port 3000
	server := NewAPIServer(":3000", client)
	server.Run()

}
