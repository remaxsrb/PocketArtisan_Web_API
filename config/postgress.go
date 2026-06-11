package config

import (
	"PocketArtisan/internal/modules/cart"
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
	log.Println("Performing initial database migration...")

	migrations := []migration{
		{
			table:   "carts",
			migrate: func() error { return DB.AutoMigrate(&cart.Cart{}) },
		},
		{
			table:   "cart_items",
			migrate: func() error { return DB.AutoMigrate(&cart.CartItem{}) },
		},
		{
			table: "users",
			migrate: func() error {
				if err := DB.Exec("CREATE TYPE IF NOT EXISTS gender AS ENUM ('male', 'female')").Error; err != nil {
					log.Printf("Warning: failed to create gender enum type: %v", err)
				}
				return DB.AutoMigrate(&users.User{})
			},
		},
		{
			table:   "craftsmen",
			migrate: func() error { return DB.AutoMigrate(&users.Craftsman{}) },
		},
		{
			table:   "craftsman_applications",
			migrate: func() error { return DB.AutoMigrate(&craftsman_application.CraftsmanApplication{}) },
		},
		{
			table: "products",
			migrate: func() error {
				return DB.AutoMigrate(&product.Product{}, &product.ProductImage{}, &product.ProductVideo{})
			},
		},
	}

	for _, m := range migrations {
		if !tableExists(m.table) {
			if err := m.migrate(); err != nil {
				log.Fatalf("Failed to migrate table %q: %v", m.table, err)
			}
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

func tableExists(name string) bool {
	var exists bool
	err := DB.Raw("SELECT EXISTS (SELECT FROM pg_tables WHERE tablename = ?)", name).Scan(&exists).Error
	if err != nil {
		log.Fatalf("Failed to check if table %q exists: %v", name, err)
	}
	return exists
}
