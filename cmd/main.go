package main

import (
	"PocketArtisan/config"
	"PocketArtisan/internal/container"
	"PocketArtisan/internal/http"
	"PocketArtisan/internal/modules/auth"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config.InitPostgresDB()
	config.InitRedis()
	config.InitCrypto()

	auth.InitJWTService(24 * time.Hour)

	appContainer := container.NewAppContainer(
		config.DB,
		config.RDB,
		auth.GetJWTService(),
	)

	r := http.SetupRouter(appContainer)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
