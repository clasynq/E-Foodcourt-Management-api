package config

import (
	"context"
	"log"
	"os"

	"user-service/internal/model"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// InitDB connects to the PostgreSQL database using DB_DSN environment variable
// and synchronizes the table schemas.
func InitDB() *gorm.DB {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("Error: DB_DSN environment variable is not set in .env file.")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	log.Println("Database connection established successfully.")

	// AutoMigrate will dynamically create/sync the users table
	err = db.AutoMigrate(
		&model.User{},
	)
	if err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}

	log.Println("Database structures migrated successfully.")
	return db
}

// Add this to imports:
// "github.com/redis/go-redis/v9"
// InitRedis connects to the Redis cache using REDIS_URL

func InitRedis() *redis.Client {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Fatalf("Redis isn't configured perfectly please onfigure the REDIS in the env")
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse redis url: %v", err)
	}
	rdb := redis.NewClient(opts)

	//ping the redis server to verify that redis is alive or not
	ctx := context.Background() // import "context" at the top of the file
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Redis connection established succesasfully.")
	return rdb
}
