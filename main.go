package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	var (
		usernameFlag   = flag.String("u", "", "Username")
		passwordFlag   = flag.String("p", "", "Password")
		hostFlag       = flag.String("h", "", "Host")
		dbNameFlag     = flag.String("d", "", "Database Name")
		collectionFlag = flag.String("c", "", "Collection Name")
		queryFlag      = flag.String("q", "", "Optional query")
	)
	flag.Parse()

	username := *usernameFlag
	password := *passwordFlag
	host := *hostFlag
	dbName := *dbNameFlag
	collection := *collectionFlag
	query := *queryFlag

	if username == "" || password == "" || host == "" || dbName == "" || collection == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Set up MongoDB connection
	uri := fmt.Sprintf("mongodb://%s:%s@%s", username, password, host)
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Error creating MongoDB client: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	// Access database and collection
	db := client.Database(dbName)
	coll := db.Collection(collection)

	// Optionally apply query
	filter := make(map[string]interface{})
	if query != "" {
		filter = map[string]interface{}{
			"$text": map[string]interface{}{
				"$search": query,
			},
		}
	}

	// Project to exclude _id field
	projection := bson.M{"_id": 0}

	// Execute query and print results to stdout as JSON
	cursor, err := coll.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		log.Fatalf("Error querying MongoDB: %v", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			log.Fatalf("Error decoding result: %v", err)
		}
		jsonResult, err := json.Marshal(result)
		if err != nil {
			log.Fatalf("Error encoding result to JSON: %v", err)
		}
		fmt.Println(string(jsonResult))
	}
	if err := cursor.Err(); err != nil {
		log.Fatalf("Cursor error: %v", err)
	}
}
