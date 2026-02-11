package config

import (
	"log"
	"os"
	"sync"
)

type Crypto struct {
	JwtKeySecret string
}

var (
	cryptoInstance *Crypto
	once           sync.Once
)

func InitCrypto() {
	once.Do(func() {

		jwt := os.Getenv("JWT_SECRET")
		if jwt == "" {
			log.Fatal("JWT_SECRET must be set")
		}

		cryptoInstance = &Crypto{
			JwtKeySecret: jwt,
		}
	})
}

func GetCrypto() *Crypto {
	if cryptoInstance == nil {
		log.Fatal("Crypto not initialized. Call InitCrypto() first")
	}
	return cryptoInstance
}
