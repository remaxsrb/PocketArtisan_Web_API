package main

import (
	"PocketArtisan/config"
	"PocketArtisan/internal/container"
	"PocketArtisan/internal/http"
	"PocketArtisan/internal/modules/auth"
	"PocketArtisan/internal/modules/payment"
	"PocketArtisan/internal/modules/utils/fonts"
	"PocketArtisan/internal/modules/utils/timeutil"
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

	fileStorage := config.InitStorage(baseURL)

	fontService, err := fonts.NewService("./assets")
	if err != nil {
		log.Fatalf("failed to load fonts: %v", err)
	}

	provider := os.Getenv("PAYMENT_PROVIDER")
	gateway, err := payment.NewGateway(provider)
	if err != nil {
		log.Fatalf("failed to initialize payment gateway: %v", err)
	}
	wrappedGateway := payment.NewBreakerGateway(gateway, 5, 30*time.Second)

	timeService := timeutil.NewService()

	appContainer := container.NewAppContainer(
		config.DB,
		config.RDB,
		jwtService,
		fileStorage,
		fontService,
		wrappedGateway,
		timeService,
	)

	r := http.SetupRouter(appContainer)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
