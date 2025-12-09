package config

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"PocketArtisan/internal/modules/users/common"
)

var DB *gorm.DB

func InitPostgresDB() {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Europe/Belgrade",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_PORT"),
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	var tableExists bool
	err = DB.Raw("SELECT EXISTS (SELECT FROM pg_tables WHERE tablename = 'users')").Scan(&tableExists).Error
	if err != nil {
		log.Fatal("Failed to check if users table exists:", err)
	}

	if !tableExists {
		err = DB.Exec("CREATE TYPE gender AS ENUM ('male', 'female')").Error
		if err != nil {
			log.Printf("Warning: failed to create gender enum type: %v", err)
		}
		if err := DB.AutoMigrate(&common.User{}); err != nil {
			log.Fatal("Failed to migrate models:", err)
		}
		log.Println("Database initialized successfully")
	} else {
		log.Println("Database already exists, skipping initialization")
	}
}
