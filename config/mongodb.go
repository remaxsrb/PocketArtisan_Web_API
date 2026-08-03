package config

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var MongoDB *mongo.Database

func InitMongoDB() {
	uri := os.Getenv("MONGO_URI")
	if password := os.Getenv("MONGO_PASSWORD"); password != "" {
		uri = strings.Replace(uri, "<db_password>", password, 1)
	}
	dbName := os.Getenv("MONGO_DB_NAME")

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(opts)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}

	MongoDB = client.Database(dbName)
	log.Println("MongoDB initialized successfully")
}
