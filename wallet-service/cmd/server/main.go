package main

import (
	"log"
	"net/http"
	"os"
	"wallet-service/internal/config"
	"wallet-service/internal/handler"
	"wallet-service/internal/repository"
	"wallet-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables (forcing override of global vars)
	_ = godotenv.Overload(".env")
	_ = godotenv.Overload("../.env")
	_ = godotenv.Overload("../../.env")

	// Initialize config
	config.InitDB()
	config.InitRedis()

	// Dependency Injection
	repo := repository.NewWalletRepository(config.DB)
	srv := service.NewWalletService(repo)
	hdl := handler.NewWalletHandler(srv)

	// Create Gin router
	r := gin.Default()

	// CORS middleware configuration
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

	// Routing setup
	api := r.Group("/api/wallet")
	{
		api.GET("/balance", hdl.GetBalance)
		api.GET("/student", hdl.GetStudent)
		api.GET("/history", hdl.GetHistory)
		api.GET("/summary", hdl.GetSummary)

		api.POST("/recharge/nfc", hdl.NfcRecharge)
		api.POST("/recharge/manual", hdl.ManualRecharge)
		api.POST("/recharge/online", hdl.OnlineRecharge)
		api.POST("/deduct", hdl.DeductBalance)
		api.POST("/webhook/razorpay", hdl.RazorpayWebhook)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	log.Printf("Wallet Service running on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server launch failed: %v", err)
	}
}
