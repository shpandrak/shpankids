package database

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
)

func TryMeDude(ctx context.Context) {
	// Sets your Google Cloud Platform project ID.
	projectID := "shpankids"

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	// Close client when done with
	defer client.Close()

	usersCollection := client.Collection("users")
	const userEmail = "shpandrak@gmail.com"
	_, err = usersCollection.Doc(userEmail).Set(ctx, map[string]interface{}{
		"email": userEmail,
		"first": "Amit",
		"last":  "Lieberman",
		"born":  1981,
	})
	if err != nil {
		log.Fatalf("Failed adding bootstrapping users: %v", err)
	}
}
