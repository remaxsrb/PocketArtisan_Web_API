//go:build integration

package order_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"PocketArtisan/internal/entities"
	"PocketArtisan/internal/modules/payment"
	"PocketArtisan/internal/modules/utils/fonts"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	testDB    *gorm.DB
	testFonts *fonts.Service // nil if assets dir missing — create tests skip
)

func TestMain(m *testing.M) {
	dsn := testEnvOrDefault("TEST_DATABASE_URL", "host=localhost user=postgres password=postgres dbname=pocketartisan_test port=5432 sslmode=disable")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "integration: cannot connect to test DB: %v\n", err)
		os.Exit(1)
	}

	testDB = db

	if f, ferr := fonts.NewService(testEnvOrDefault("TEST_ASSETS_DIR", "../../../assets")); ferr == nil {
		testFonts = f
	}

	os.Exit(m.Run())
}

func testEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// withTx wraps fn in a DB transaction that is always rolled back.
// Every test gets an isolated snapshot of the DB with no leftover state.
func withTx(t *testing.T, fn func(tx *gorm.DB)) {
	t.Helper()
	tx := testDB.Begin()
	if tx.Error != nil {
		t.Fatalf("begin transaction: %v", tx.Error)
	}
	t.Cleanup(func() { tx.Rollback() })
	fn(tx)
}

// ── Fixtures ──────────────────────────────────────────────────────────────────

type testFixtures struct {
	Customer  entities.User
	Craftsman entities.User
	Product   entities.Product
}

// seedFixtures inserts a customer, a craftsman, and a product into tx.
// All rows are inside the rolled-back transaction, so the DB stays clean.
func seedFixtures(t *testing.T, tx *gorm.DB) testFixtures {
	t.Helper()

	customer := entities.User{
		Username: "test_customer",
		Email:    "customer@test.local",
		Role:     "user",
		Gender:   "M",
	}
	if err := tx.Create(&customer).Error; err != nil {
		t.Fatalf("seed customer: %v", err)
	}

	craftsman := entities.User{
		Username: "test_craftsman",
		Email:    "craftsman@test.local",
		Role:     "craftsman",
		Gender:   "M",
	}
	if err := tx.Create(&craftsman).Error; err != nil {
		t.Fatalf("seed craftsman: %v", err)
	}

	product := entities.Product{
		CraftsmanID: craftsman.ID,
		Name:        "Handmade Sword",
		Price:       3200.00,
		Description: "A fine handmade sword",
		Available:   true,
		Hidden:      false,
	}
	if err := tx.Create(&product).Error; err != nil {
		t.Fatalf("seed product: %v", err)
	}

	return testFixtures{Customer: customer, Craftsman: craftsman, Product: product}
}

// ctxFor returns a context carrying the given user ID, matching what the JWT middleware sets.
func ctxFor(userID uint64) context.Context {
	return context.WithValue(context.Background(), "user_id", userID)
}

// freshMock returns a new MockGateway. Never share one between tests.
func freshMock() *payment.MockGateway {
	return payment.NewMockGateway()
}
