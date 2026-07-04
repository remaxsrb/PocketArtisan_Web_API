package config

import (
	"log"
	"os"

	"PocketArtisan/internal/modules/files/storage"
)

// InitStorage builds the file Storage implementation based on environment
// configuration. Production uses Cloudflare R2; every other APP_ENV (dev,
// feature branches, unset) uses local disk storage so local work never
// writes to the shared production bucket.
func InitStorage(baseURL string) storage.Storage {
	if os.Getenv("APP_ENV") == "production" {
		r2Storage, err := storage.NewR2Storage(
			os.Getenv("R2_ACCOUNT_ID"),
			os.Getenv("R2_ACCESS_KEY_ID"),
			os.Getenv("R2_SECRET_ACCESS_KEY"),
			os.Getenv("R2_BUCKET_NAME"),
			os.Getenv("R2_ENDPOINT"),
			os.Getenv("R2_PUBLIC_URL"),
		)
		if err != nil {
			log.Fatalf("failed to init R2 storage: %v", err)
		}
		log.Println("Using Cloudflare R2 storage")
		return r2Storage
	}

	log.Println("Using local file storage")
	return storage.NewLocalStorage("./uploads", baseURL+"/api/files")
}
