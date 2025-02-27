package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"flag"
	"fmt"
	"log"
	"shpankids/app"
	firestorekvs "shpankids/infra/database/firestore"
	"shpankids/infra/database/kvstore"
	"shpankids/shpankids"
)

const projectID = "shpankids"

// TIP <p>Press <shortcut actionId="ShowIntentionActions"/> when your caret is at the underlined text
// to see how GoLand suggests fixing the warning.</p><p>Alternatively, if available, click the lightbulb to view possible fixes.</p>
func main() {

	// Sets your Google Cloud Platform project ID.

	// get a flag for localdev
	runtimeEnv := flag.String("runtime-env", "prod", "have a separate flag for local development")
	flag.Parse()

	err := shpankids.DetectSecrets()
	if err != nil {
		log.Fatalf("Failed to detect secrets: %v", err)
	}

	var kvs kvstore.RawJsonStore
	if *runtimeEnv == "dev" {
		kvs = kvstore.NewInMemoryRawJsonStore()
	} else {
		fs, err := firestore.NewClient(context.Background(), projectID)
		if err != nil {
			log.Fatalf("Failed to create client firestore client: %v", err)
		}
		//Close client when done with
		kvs = firestorekvs.NewFirestoreKvs(fs, nil)
		defer fs.Close()
	}

	log.Fatalf(fmt.Sprintf("%v", app.Start(kvs)))

}
