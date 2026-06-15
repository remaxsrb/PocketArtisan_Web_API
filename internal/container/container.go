package container

import (
	"PocketArtisan/internal/modules/auth"
	"PocketArtisan/internal/modules/files/storage"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// AppContainer holds all application dependencies
type AppContainer struct {
	DB         *gorm.DB
	RDB        *redis.Client
	JWTService auth.JWTService
	Storage    storage.Storage
}

func NewAppContainer(db *gorm.DB, rdb *redis.Client, jwtService auth.JWTService, s storage.Storage) *AppContainer {
	return &AppContainer{
		DB:         db,
		RDB:        rdb,
		JWTService: jwtService,
		Storage:    s,
	}
}
