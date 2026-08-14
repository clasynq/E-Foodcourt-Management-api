package config

import (
	"context"
	"log"
	"order-kitchen-service/internal/model"
	"os"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	// DB stores the globally shared GORM PostgreSQL connection instance
	DB  *gorm.DB
	// RDB stores the globally shared Redis client session connection instance
	RDB *redis.Client
)

// InitDB reads the PostgreSQL DSN config from env and establishes a connection pool.
// It also runs auto-migrations to keep order schemas in sync.
func InitDB() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN variable is required")
	}
	var err error
	
	// Establish database connection with GORM, defaulting logger level to Silent
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	
	// AutoMigrate will dynamically create or update the orders tables
	err = DB.AutoMigrate(&model.Order{}, &model.FoodCategory{}, &model.FoodItem{})
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
}

// InitRedis parses the REDIS_URL environment config and opens a connection pool to Redis,
// verifying the connectivity using a Ping pinging verification call.
func InitRedis() {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Fatalf("Redis isn't configured perfectly. Please configure REDIS_URL in .env")
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse redis url: %v", err)
	}
	
	RDB = redis.NewClient(opts)

	ctx := context.Background()
	if err := RDB.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v. Continuing without Redis.", err)
	} else {
		log.Println("Redis connection established successfully.")
	}
}

