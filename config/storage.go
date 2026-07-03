package config

import (
	"log"
	"os"

	"PocketArtisan/internal/modules/files/storage"
)

// InitStorage builds the file Storage implementation based on environment
// configuration. When R2 credentials are present it uses Cloudflare R2,
// otherwise it falls back to local disk storage.
func InitStorage(baseURL string) storage.Storage {
	if os.Getenv("R2_ACCESS_KEY_ID") != "" {
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
