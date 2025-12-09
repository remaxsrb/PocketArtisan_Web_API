package main

import (
	"PocketArtisan/config"
	"PocketArtisan/internal/http"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config.InitPostgresDB()
	config.InitRedis()

	r := http.SetupRouter()

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
