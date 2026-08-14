package config

import (
	"context"
	"log"
	"os"
	"time"
	"wallet-service/internal/model"

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
		log.Fatal("DB_DSN environment variable is required")
	}
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = DB.AutoMigrate(&model.StudentWalletAccount{}, &model.RechargeRecord{})
	if err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	seedMockStudents()
}

func InitRedis() {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Fatalf("Redis is not configured. Please define REDIS_URL in env")
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

	log.Println("Redis connection established successfully.")
}

func seedMockStudents() {
	var count int64
	DB.Model(&model.StudentWalletAccount{}).Count(&count)
	if count > 0 {
		log.Println("Mock students already seeded. Skipping seeder.")
		return
	}

	mockStudents := []model.StudentWalletAccount{
		{
			StudentID:  "STU-2026-001",
			Name:       "Arindam Roy",
			Email:      "arindam@dinesynq.com",
			RFIDCardID: "RFID-001-AR",
			Department: "Computer Science",
			Avatar:     "https://api.dicebear.com/9.x/avataaars/svg?seed=Arindam",
			Balance:    2450.00,
			Phone:      "+91 98765 43210",
			UpdatedAt:  time.Now(),
		},
		{
			StudentID:  "STU-2026-002",
			Name:       "Priya Sharma",
			Email:      "priya@dinesynq.com",
			RFIDCardID: "RFID-002-PS",
			Department: "Electronics",
			Avatar:     "https://api.dicebear.com/9.x/avataaars/svg?seed=Priya",
			Balance:    1200.00,
			Phone:      "+91 98765 43211",
			UpdatedAt:  time.Now(),
		},
		{
			StudentID:  "STU-2026-003",
			Name:       "Sneha Patel",
			Email:      "sneha@dinesynq.com",
			RFIDCardID: "RFID-003-SP",
			Department: "Mechanical",
			Avatar:     "https://api.dicebear.com/9.x/avataaars/svg?seed=Sneha",
			Balance:    850.00,
			Phone:      "+91 98765 43215",
			UpdatedAt:  time.Now(),
		},
		{
			StudentID:  "STU-2026-004",
			Name:       "Amit Gupta",
			Email:      "amit@dinesynq.com",
			RFIDCardID: "RFID-004-AG",
			Department: "Civil Engineering",
			Avatar:     "https://api.dicebear.com/9.x/avataaars/svg?seed=Amit",
			Balance:    500.00,
			Phone:      "+91 98765 43216",
			UpdatedAt:  time.Now(),
		},
		{
			StudentID:  "STU-2026-005",
			Name:       "Kavita Nair",
			Email:      "kavita@dinesynq.com",
			RFIDCardID: "RFID-005-KN",
			Department: "Information Tech",
			Avatar:     "https://api.dicebear.com/9.x/avataaars/svg?seed=Kavita",
			Balance:    310.00,
			Phone:      "+91 98765 43217",
			UpdatedAt:  time.Now(),
		},
	}

	for _, student := range mockStudents {
		if err := DB.Create(&student).Error; err != nil {
			log.Printf("Warning: Failed to seed student %s: %v", student.Name, err)
		}
	}

	log.Println("Mock students seeded successfully.")
}
