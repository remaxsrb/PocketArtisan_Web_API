package container

import (
	"PocketArtisan/internal/modules/auth"
	"PocketArtisan/internal/modules/files/storage"
	"PocketArtisan/internal/modules/mail"
	"PocketArtisan/internal/modules/payment"
	"PocketArtisan/internal/modules/utils/fonts"
	"PocketArtisan/internal/modules/utils/timeutil"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// AppContainer holds all application dependencies
type AppContainer struct {
	DB             *gorm.DB
	RDB            *redis.Client
	JWTService     auth.JWTService
	Storage        storage.Storage
	Fonts          *fonts.Service
	BreakerGateway *payment.BreakerGateway
	TimeService    timeutil.Service
	MailService    mail.Service
} 

func NewAppContainer(db *gorm.DB, rdb *redis.Client, jwtService auth.JWTService, s storage.Storage, f *fonts.Service, bg *payment.BreakerGateway, ts timeutil.Service, m mail.Service) *AppContainer {
	return &AppContainer{
		DB:             db,
		RDB:            rdb,
		JWTService:     jwtService,
		Storage:        s,
		Fonts:          f,
		BreakerGateway: bg,
		TimeService:    ts,
	}
}
