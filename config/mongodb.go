package config

import (
	"context"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var MongoDB *mongo.Database

func InitMongoDB() {
	uri := os.Getenv("MONGO_URI")

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}

	MongoDB = client.Database(dbNameFromURI(uri))
	log.Println("MongoDB initialized successfully")
}

func dbNameFromURI(uri string) string {
	u, err := url.Parse(uri)
	if err != nil {
		log.Fatal("Invalid MONGO_URI:", err)
	}
	return strings.TrimPrefix(u.Path, "/")
}
