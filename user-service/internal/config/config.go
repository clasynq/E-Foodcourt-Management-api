package config

import (
	"log"
	"os"

	"user-service/internal/model"

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
	err = db.AutoMigrate(&model.User{})
	if err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}

	log.Println("Database structures migrated successfully.")
	return db
}
