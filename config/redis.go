package config

import (
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
)

var RDB *redis.Client

func InitRedis() {
	RDB = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
	})
	if err := RDB.Ping(RDB.Context()).Err(); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	log.Println("Redis initialized successfully")
}
