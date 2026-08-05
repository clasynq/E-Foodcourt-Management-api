package config

import (
	"context"
	"log"
	"manager-dashboard/internal/model"
	"os"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB  *gorm.DB
	RDB *redis.Client
)

func InitDB() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN variable id required")
	}
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("ailed to connect to database: %v", err)
	}
	err = DB.AutoMigrate(&model.InventoryItem{}, &model.LocalOrder{})
	if err != nil {
		log.Fatalf("ailed to run migrations: %v", err)
	}
}

// Replace lines 38-57 with:
func InitRedis() {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Fatalf("Redis isn't configured perfectly. Please configure REDIS_URL in .env")
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse redis url: %v", err)
	}

	// FIX: Assign directly to the package-level RDB variable
	RDB = redis.NewClient(opts)

	ctx := context.Background()
	if err := RDB.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Redis connection established successfully.")
}
