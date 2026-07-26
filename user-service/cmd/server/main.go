package main

import (
	"log"
	"os"

	"user-service/internal/config"
	"user-service/internal/handler"
	"user-service/internal/repository"
	"user-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load local environment configurations if available
	if err := godotenv.Load(); err != nil {
		log.Println("Info: No local .env file found. Reading system environment variables.")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// 2. Initialize Database pool and run migrations
	db := config.InitDB()
	rdb := config.InitRedis()

	// 3. Dependency Injection (Wiring layers together)
	userRepo := repository.NewUserRepository(db)
	authSvc := service.NewAuthService(userRepo, rdb)
	authHandler := handler.NewAuthHandler(authSvc)

	// 4. Initialize Gin Web Engine
	r := gin.Default()

	// 5. Register Endpoint Routes
	authRoutes := r.Group("/api/auth")
	{
		authRoutes.POST("/signup", authHandler.Signup)
		authRoutes.POST("/login", authHandler.Login)
		authRoutes.POST("/logout", authHandler.Logout)
		authRoutes.POST("/send-otp", authHandler.SendOTP)
		authRoutes.POST("/verify-otp", authHandler.VerifyOTP)
		authRoutes.POST("/forgot-password", authHandler.ForgotPassword)
		authRoutes.POST("/reset-password", authHandler.ResetPassword)
	}

	// 6. Start the web server
	log.Printf("User Service is running on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server startup failed: %v", err)
	}
}
