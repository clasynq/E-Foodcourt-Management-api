package main

import (
	"log"
	"net/http"
	"os"

	"order-kitchen-service/internal/config"
	"order-kitchen-service/internal/handler"
	"order-kitchen-service/internal/repository"
	"order-kitchen-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment configurations from multiple parent path offsets to support run directory variations
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")

	// Initialize Postgres and Redis database connection clients
	config.InitDB()
	config.InitRedis()

	// Wire repository, service logic dependency layer, and router handler controllers
	repo := repository.NewOrderRepository(config.DB)
	srv := service.NewOrderService(repo)
	hdl := handler.NewOrderHandler(srv)

	menuRepo := repository.NewMenuRepository(config.DB)
	menuSrv := service.NewMenuService(menuRepo)
	menuHdl := handler.NewMenuHandler(menuSrv)

	// Create a default Gin HTTP router engine instance
	r := gin.Default()

	// CORS Setup to support frontend requests without domain block issues
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

	// Map API endpoint endpoints router tree structure
	api := r.Group("/api")
	{
		manager := api.Group("/manager")
		{
			manager.GET("/orders", hdl.GetActiveOrders)
			manager.PUT("/orders/:id/status", hdl.UpdateOrderStatus)

			// Menu Management CRUD Routes
			manager.GET("/menu", menuHdl.ListFoodItems)
			manager.GET("/categories", menuHdl.ListCategories)
			manager.POST("/menu", menuHdl.CreateFoodItem)
			manager.PUT("/menu/:id", menuHdl.UpdateFoodItem)
			manager.DELETE("/menu/:id", menuHdl.DeleteFoodItem)
			manager.PUT("/menu/:id/availability", menuHdl.ToggleAvailability)
			manager.PUT("/menu/:id/stock", menuHdl.UpdateStock)
		}
	}

	// Read microservice listening port from environments, defaulting to 8084
	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	// Boot up and listen for incoming HTTP connections
	log.Printf("Order & Kitchen Service listening on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server launch failed: %v", err)
	}
}

