package main

import (
	"PocketArtisan/config"
	"PocketArtisan/internal/container"
	"PocketArtisan/internal/http"
	"PocketArtisan/internal/modules/auth"
	"PocketArtisan/internal/modules/files/storage"
	"PocketArtisan/internal/modules/utils/fonts"
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

	jwtService := auth.InitJWTService(24 * time.Hour)

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	localStorage := storage.NewLocalStorage("./uploads", baseURL+"/api/files")

	fontService, err := fonts.NewService("./assets")
	if err != nil {
		log.Fatalf("failed to load fonts: %v", err)
	}

	appContainer := container.NewAppContainer(
		config.DB,
		config.RDB,
		jwtService,
		localStorage,
		fontService,
	)

	r := http.SetupRouter(appContainer)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
