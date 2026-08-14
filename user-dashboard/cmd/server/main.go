package main

import (
	"log"
	"net/http"
	"os"
	"user-dashboard/internal/config"
	"user-dashboard/internal/handler"
	"user-dashboard/internal/repository"
	"user-dashboard/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment config
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")

	// Initialize database and cache connections
	config.InitDB()
	config.InitRedis()

	// Setup repository, service, and handlers
	repo := repository.NewDashboardRepository(config.DB)
	srv := service.NewDashboardService(repo, config.RDB)
	hdl := handler.NewDashboardHandler(srv)

	// Initialize Gin router
	r := gin.Default()

	// CORS Setup to support frontend requests
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-User-Id, X-User-Name, X-User-Role")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Routing group mapping
	api := r.Group("/api")
	{
		student := api.Group("/student")
		{
			dashboard := student.Group("/dashboard")
			{
				dashboard.GET("/overview", hdl.GetOverview)
				dashboard.POST("/rewards", hdl.AddRewards)
			}
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8088"
	}

	log.Printf("User Dashboard Service running on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server launch failed: %v", err)
	}
}
