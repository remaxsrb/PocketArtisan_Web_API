package config

import (
	"log"
	"os"
	"sync"
)

type Crypto struct {
	JwtKeySecret    string
	TurnstileSecret string
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

		turnstile := os.Getenv("TURNSTILE_SECRET")
		if turnstile == "" {
			log.Fatal("TURNSTILE_SECRET must be set")
		}

		cryptoInstance = &Crypto{
			JwtKeySecret:    jwt,
			TurnstileSecret: turnstile,
		}
	})
}

func GetCrypto() *Crypto {
	if cryptoInstance == nil {
		log.Fatal("Crypto not initialized. Call InitCrypto() first")
	}
	return cryptoInstance
}
