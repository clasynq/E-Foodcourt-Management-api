package main

import (
	"log"
	"net/http"
	"os"

	"manager-dashboard/internal/config"
	"manager-dashboard/internal/handler"
	"manager-dashboard/internal/repository"
	"manager-dashboard/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment config
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")

	config.InitDB()
	config.InitRedis()

	// Wire GORM repository, service logic, and handler controllers
	repo := repository.NewDashboardRepository(config.DB)
	srv := service.NewDashboardService(repo, config.RDB)
	hdl := handler.NewDashboardHandler(srv)

	// Bind Gin router engine
	r := gin.Default()

	// CORS Setup to allow frontend connection
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	api := r.Group("/api")
	{
		manager := api.Group("/manager")
		{
			manager.GET("/overview", hdl.GetOverview)

			// Inventory CRUD Routes
			manager.GET("/inventory", hdl.ListInventoryItems)
			manager.POST("/inventory", hdl.CreateInventoryItem)
			manager.PUT("/inventory/:id", hdl.UpdateInventoryItem)
			manager.POST("/inventory/:id/restock", hdl.RestockInventoryItem)
			manager.DELETE("/inventory/:id", hdl.DeleteInventoryItem)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8087"
	}

	log.Printf("Manager Dashboard Service listening on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server launch failed: %v", err)
	}
}
