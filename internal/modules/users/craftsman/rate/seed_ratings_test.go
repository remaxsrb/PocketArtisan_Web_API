package rate

import (
	"fmt"
	"math/rand"
	"os"
	"testing"

	"PocketArtisan/internal/entities"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func connectSeedDB(t *testing.T) *gorm.DB {
	t.Helper()
	_ = godotenv.Load("../../../../../.env")
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
		t.Fatalf("Failed to connect to database: %v", err)
	}
	return db
}

func TestSeedCraftsmanRatings(t *testing.T) {
	db := connectSeedDB(t)

	var craftsmen []entities.Craftsman
	if err := db.Find(&craftsmen).Error; err != nil {
		t.Fatalf("Failed to fetch craftsmen: %v", err)
	}
	if len(craftsmen) == 0 {
		t.Fatal("No craftsmen found — run the craftsman creation step first")
	}

	t.Logf("Seeding ratings for %d craftsmen...", len(craftsmen))

	for i := range craftsmen {
		c := &craftsmen[i]
		c.NumberOfRatings = 10 + rand.Intn(91)          // 10–100
		c.Rating = 3.0 + rand.Float64()*2.0             // 3.0–5.0

		if err := db.Save(c).Error; err != nil {
			t.Errorf("Failed to update craftsman ID %d: %v", c.ID, err)
			continue
		}
		t.Logf("Craftsman ID %d → rating=%.2f, count=%d", c.ID, c.Rating, c.NumberOfRatings)
	}

	t.Log("Rating seed completed.")
}