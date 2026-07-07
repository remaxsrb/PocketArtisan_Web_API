package config

import (
	"PocketArtisan/internal/entities"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

type migration struct {
	table   string
	migrate func() error
}

func InitPostgresDB() {
	initSSLMode()
	DB = mustConnectDB()
	runMigrations()
	runIndexes()
	runSeeds()
	log.Println("Postgres ready")
}

var sslMode string

func initSSLMode() {
	sslMode = os.Getenv("POSTGRES_SSLMODE")
	if sslMode == "" {
		sslMode = "require"
	}
}

func mustConnectDB() *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Europe/Belgrade",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_PORT"),
		sslMode,
	)
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get underlying sql.DB:", err)
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

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
			table: "craftsmen",
			migrate: func() error {
				if err := DB.AutoMigrate(&entities.Craftsman{}); err != nil {
					return err
				}
				return DB.Exec(`UPDATE craftsmen SET approved_at = NOW() WHERE approved_at IS NULL`).Error
			},
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
			table:   "craft_product_categories",
			migrate: func() error { return DB.AutoMigrate(&entities.CraftProductCategory{}) },
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
		`CREATE INDEX IF NOT EXISTS idx_craftsmen_approved_at ON craftsmen (approved_at)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_completed_at ON orders (completed_at)`,
	}
	for _, idx := range indexes {
		if err := DB.Exec(idx).Error; err != nil {
			log.Fatalf("Failed to create index: %v\nQuery: %s", err, idx)
		}
	}
}
