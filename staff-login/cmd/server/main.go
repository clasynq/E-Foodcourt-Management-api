package main

import (
	"log"
	"os"

	"staff-login/internal/config"
	"staff-login/internal/handler"
	"staff-login/internal/repository"
	"staff-login/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load local environment configuration if available (forcing override of global vars)
	if err := godotenv.Overload(); err != nil {
		log.Println("Info: No local .env file found. Reading system environment variables.")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8086"
	}

	// 2. Initialize Database pool and run migrations/seeding
	db := config.InitDB()
	rdb := config.InitRedis()

	// 3. Dependency Injection (Wiring layers together)
	staffRepo := repository.NewStaffRepository(db)
	staffSvc := service.NewStaffService(staffRepo, rdb)
	staffHandler := handler.NewStaffHandler(staffSvc)

	// 4. Initialize Gin Web Engine
	r := gin.Default()

	// 5. Register Endpoint Routes
	staffRoutes := r.Group("/api/staff")
	{
		staffRoutes.POST("/login/initiate", staffHandler.InitiateLogin)
		staffRoutes.POST("/login/verify", staffHandler.VerifyLogin)
		staffRoutes.POST("/login", staffHandler.Login)
		staffRoutes.POST("/logout", staffHandler.Logout)
	}

	// 6. Start the web server
	log.Printf("Staff Login Service is running on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server startup failed: %v", err)
	}
}
