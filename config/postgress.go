package config

import (
	"PocketArtisan/internal/entities"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

type migration struct {
	table   string
	migrate func() error
}

func InitPostgresDB() {
	DB = mustConnectDB()
	runMigrations()
	runIndexes()
	log.Println("Postgres ready")
}

func mustConnectDB() *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Europe/Belgrade",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_PORT"),
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	return db
}

func runMigrations() {
	log.Println("Performing database migrations...")

	migrations := []migration{
		{
			table:   "carts",
			migrate: func() error { return DB.AutoMigrate(&entities.Cart{}) },
		},
		{
			table:   "cart_items",
			migrate: func() error { return DB.AutoMigrate(&entities.CartItem{}) },
		},
		{
			table: "users",
			migrate: func() error {
				return DB.AutoMigrate(&entities.User{})
			},
		},
		{
			table:   "crafts",
			migrate: func() error { return DB.AutoMigrate(&entities.Craft{}) },
		},
		{
			table:   "craftsmen",
			migrate: func() error { return DB.AutoMigrate(&entities.Craftsman{}) },
		},
		{
			table:   "craftsman_applications",
			migrate: func() error { return DB.AutoMigrate(&entities.CraftsmanApplication{}) },
		},
		{
			table:   "product_categories",
			migrate: func() error { return DB.AutoMigrate(&entities.ProductCategory{}) },
		},
		{
			table: "products",
			migrate: func() error {
				return DB.AutoMigrate(&entities.Product{}, &entities.ProductImage{}, &entities.ProductVideo{})
			},
		},
		{
			table: "orders",
			migrate: func() error {
				return DB.AutoMigrate(&entities.Order{}, &entities.OrderItem{})
			},
		},
		{
			table:   "craftsman_rating_records",
			migrate: func() error { return DB.AutoMigrate(&entities.CraftsmanRatingRecord{}) },
		},
	}

	for _, m := range migrations {
		if err := m.migrate(); err != nil {
			log.Fatalf("Failed to migrate table %q: %v", m.table, err)
		}
	}
}

func runIndexes() {
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_users_created_at_id ON users (created_at DESC, id DESC)`,
	}
	for _, idx := range indexes {
		if err := DB.Exec(idx).Error; err != nil {
			log.Fatalf("Failed to create index: %v\nQuery: %s", err, idx)
		}
	}
}
