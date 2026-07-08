package main

import (
	"PocketArtisan/config"
	"PocketArtisan/internal/container"
	"PocketArtisan/internal/http"
	"PocketArtisan/internal/modules/auth"
	"PocketArtisan/internal/modules/mail"
	"PocketArtisan/internal/modules/order"
	"PocketArtisan/internal/modules/order/reviewreminder"
	"PocketArtisan/internal/modules/payment"
	"PocketArtisan/internal/modules/users"
	"PocketArtisan/internal/modules/utils/fonts"
	"PocketArtisan/internal/modules/utils/timeutil"
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config.InitPostgresDB()
	config.InitRedis()
	config.InitCrypto()
	//config.InitMongoDB()

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

	mailProvider := os.Getenv("MAIL_PROVIDER")
	mailer, err := mail.NewService(mailProvider)
	if err != nil {
		log.Fatalf("failed to initialize mail service: %v", err)
	}

	timeService := timeutil.NewService()

	appContainer := container.NewAppContainer(
		config.DB,
		config.RDB,
		jwtService,
		fileStorage,
		fontService,
		wrappedGateway,
		timeService,
		mailer,
	)

	reviewReminderSvc := reviewreminder.NewService(
		order.NewGormRepository(config.DB),
		users.NewGormRepository(config.DB),
		mailer,
		5*time.Minute,
	)

	c := cron.New()
	_, err = c.AddFunc(os.Getenv("REVIEW_REMINDER_CRON_SCHEDULE"), func() {
		if err := reviewReminderSvc.Execute(context.Background()); err != nil {
			log.Printf("review reminder run failed: %v", err)
		}
	})
	if err != nil {
		log.Fatalf("failed to schedule review reminder cron: %v", err)
	}
	c.Start()

	r := http.SetupRouter(appContainer)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
