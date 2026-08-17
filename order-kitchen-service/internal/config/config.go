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

	seedCategoriesOnly()
}

func seedCategoriesOnly() {
	var categoryCount int64
	DB.Model(&model.FoodCategory{}).Count(&categoryCount)
	if categoryCount == 0 {
		log.Println("Seeding default food categories...")
		categories := []model.FoodCategory{
			{ID: "cat-1", Name: "South Indian", Slug: "south-indian", Icon: "🫓", Description: "Authentic South Indian delicacies", Image: "https://images.unsplash.com/photo-1668236543090-82bbe5ce830c?w=400&h=300&fit=crop"},
			{ID: "cat-2", Name: "North Indian", Slug: "north-indian", Icon: "🍛", Description: "Rich curries, naans, rotis and thalis", Image: "https://images.unsplash.com/photo-1631452180519-c014fe946bc7?w=400&h=300&fit=crop"},
			{ID: "cat-3", Name: "Chinese", Slug: "chinese", Icon: "🥡", Description: "Indo-Chinese favorites", Image: "https://images.unsplash.com/photo-1585032226651-759b368d7246?w=400&h=300&fit=crop"},
			{ID: "cat-4", Name: "Snacks", Slug: "snacks", Icon: "🍟", Description: "Quick bites", Image: "https://images.unsplash.com/photo-1626777552726-4a6b54c97e46?w=400&h=300&fit=crop"},
			{ID: "cat-5", Name: "Beverages", Slug: "beverages", Icon: "☕", Description: "Chai, coffee, juices, shakes", Image: "https://images.unsplash.com/photo-1517701604599-bb29b565090c?w=400&h=300&fit=crop"},
			{ID: "cat-6", Name: "Desserts", Slug: "desserts", Icon: "🍰", Description: "Sweet endings", Image: "https://images.unsplash.com/photo-1563805042-7684c019e1cb?w=400&h=300&fit=crop"},
			{ID: "cat-7", Name: "Biryani", Slug: "biryani", Icon: "🍚", Description: "Aromatic rice dishes", Image: "https://images.unsplash.com/photo-1563379091339-03b21ab4a4f8?w=400&h=300&fit=crop"},
			{ID: "cat-8", Name: "Fast Food", Slug: "fast-food", Icon: "🍔", Description: "Burgers, pizzas, wraps, and fries", Image: "https://images.unsplash.com/photo-1568901346375-23c9450c58cd?w=400&h=300&fit=crop"},
		}
		for _, cat := range categories {
			DB.Create(&cat)
		}
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

