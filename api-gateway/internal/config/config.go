package config

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func InitRedis() {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Fatalf("REDIS_URL environment variable is required")
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}

	RDB = redis.NewClient(opts)

	ctx := context.Background()
	if err := RDB.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("API Gateway established Redis connection successfully.")
}
