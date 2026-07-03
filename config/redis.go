package config

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
)

var RDB *redis.Client

func InitRedis() {
	tlsEnabled := os.Getenv("REDIS_TLS_ENABLED")
	if tlsEnabled == "" {
		tlsEnabled = "true"
	}

	opts := &redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"),
	}

	if tlsEnabled == "true" {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	RDB = redis.NewClient(opts)

	if err := RDB.Ping(RDB.Context()).Err(); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	log.Println("Redis initialized successfully")
}
