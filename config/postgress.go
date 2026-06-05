package config

import (
	"PocketArtisan/internal/modules/craftsman_application"
	"PocketArtisan/internal/modules/product"
	"PocketArtisan/internal/modules/users"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

	var userTableExists bool
	err = DB.Raw("SELECT EXISTS (SELECT FROM pg_tables WHERE tablename = 'users')").Scan(&userTableExists).Error
	if err != nil {
		log.Fatal("Failed to check if users table exists:", err)
	}

	var craftsmanTableExists bool
	err = DB.Raw("SELECT EXISTS (SELECT FROM pg_tables WHERE tablename = 'craftsmen')").Scan(&craftsmanTableExists).Error
	if err != nil {
		log.Fatal("Failed to check if craftsmen table exists:", err)
	}


	var craftsmanApplicationTableExists bool
	err = DB.Raw("SELECT EXISTS (SELECT FROM pg_tables WHERE tablename = 'craftsman_applications')").Scan(&craftsmanApplicationTableExists).Error
	if err != nil {
		log.Fatal("Failed to check if users table exists:", err)
	}

	var productTableExists bool
	err = DB.Raw("SELECT EXISTS (SELECT FROM pg_tables WHERE tablename = 'products')").Scan(&productTableExists).Error
	if err != nil {
		log.Fatal("Failed to check if users table exists:", err)
	}

	log.Println("Performing initial database migration...")

	if !userTableExists {
		

		err = DB.Exec("CREATE TYPE gender AS ENUM ('male', 'female')").Error
		if err != nil {
			log.Printf("Warning: failed to create gender enum type: %v", err)
		}

		if err := DB.AutoMigrate(&users.User{}); err != nil {
			log.Fatal("Failed to migrate users model:", err)
		}

	}

	if !craftsmanTableExists {
		if err := DB.AutoMigrate(&users.Craftsman{}); err != nil {
			log.Fatal("Failed to migrate craftsmen model:", err)
		}
	}

	if !craftsmanApplicationTableExists {
		if err := DB.AutoMigrate(&craftsman_application.CraftsmanApplication{}); err != nil {
			log.Fatal("Failed to migrate craftsman_application model:", err)
		}
	}

	if !productTableExists {
		if err := DB.AutoMigrate(&product.Product{}); err != nil {
			log.Fatal("Failed to migrate products model:", err)
		}
	}

	// Indexes

	// Users Index
	err = DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_users_created_at_id
		ON users (created_at DESC, id DESC)`).Error

	if err != nil {
		log.Fatal("Failed to create users pagination index:", err)
	}

	log.Println("Postgres ready")
}
