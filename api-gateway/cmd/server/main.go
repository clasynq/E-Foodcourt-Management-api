package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"api-gateway/internal/config"
	"api-gateway/internal/handler"
	"api-gateway/internal/proxy"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment configurations (forcing override of global vars)
	_ = godotenv.Overload(".env")
	_ = godotenv.Overload("../.env")
	_ = godotenv.Overload("../../.env")

	// Ensure secret key is loaded
	if os.Getenv("SECRECT_KEY") == "" {
		// Fallback to user-service key if loaded in parent environment
		log.Println("Warning: SECRECT_KEY is empty in local env. Making sure fallback keys are checked.")
	}

	// Initialize config
	config.InitRedis()

	// Get target microservice URLs
	userServiceURL := os.Getenv("USER_SERVICE_URL")
	walletServiceURL := os.Getenv("WALLET_SERVICE_URL")
	diningIoTServiceURL := os.Getenv("DINING_IOT_SERVICE_URL")
	orderKitchenServiceURL := os.Getenv("ORDER_KITCHEN_SERVICE_URL")
	aiAnalyticsServiceURL := os.Getenv("AI_ANALYTICS_SERVICE_URL")
	userDashboardServiceURL := os.Getenv("USER_DASHBOARD_SERVICE_URL")
	
	if userServiceURL == "" {
		userServiceURL = "http://localhost:8081"
	}
	if walletServiceURL == "" {
		walletServiceURL = "http://localhost:8082"
	}
	if diningIoTServiceURL == "" {
		diningIoTServiceURL = "http://localhost:8083"
	}
	if orderKitchenServiceURL == "" {
		orderKitchenServiceURL = "http://localhost:8084"
	}
	if aiAnalyticsServiceURL == "" {
		aiAnalyticsServiceURL = "http://localhost:8085"
	}
	if userDashboardServiceURL == "" {
		userDashboardServiceURL = "http://localhost:8088"
	}
	staffLoginServiceURL := os.Getenv("STAFF_LOGIN_SERVICE_URL")
	if staffLoginServiceURL == "" {
		staffLoginServiceURL = "http://localhost:8086"
	}
	managerDashboardServiceURL := os.Getenv("MANAGER_DASHBOARD_SERVICE_URL")
	if managerDashboardServiceURL == "" {
		managerDashboardServiceURL = "http://localhost:8087"
	}

	log.Printf("[Gateway Config] User Auth Service -> %s", userServiceURL)
	log.Printf("[Gateway Config] Wallet Service -> %s", walletServiceURL)
	log.Printf("[Gateway Config] Dining/IoT Service -> %s", diningIoTServiceURL)
	log.Printf("[Gateway Config] Order/Kitchen Service -> %s", orderKitchenServiceURL)
	log.Printf("[Gateway Config] AI/Analytics Service -> %s", aiAnalyticsServiceURL)
	log.Printf("[Gateway Config] User Dashboard Service -> %s", userDashboardServiceURL)
	log.Printf("[Gateway Config] Staff Login Service -> %s", staffLoginServiceURL)
	log.Printf("[Gateway Config] Manager Dashboard Service -> %s", managerDashboardServiceURL)

	r := gin.Default()

	// Mount global middlewares
	r.Use(handler.CORSMiddleware())
	r.Use(handler.AuthMiddleware())

	// Dynamic proxy router using prefix matching to resolve path overlap conflicts
	r.Any("/api/*any", func(c *gin.Context) {
		path := c.Request.URL.Path

		var targetURL string
		switch {
		case strings.HasPrefix(path, "/api/auth"):
			targetURL = userServiceURL
		case strings.HasPrefix(path, "/api/wallet"):
			targetURL = walletServiceURL
		case strings.HasPrefix(path, "/api/dining"):
			targetURL = diningIoTServiceURL
		case strings.HasPrefix(path, "/api/student/dashboard"):
			targetURL = userDashboardServiceURL
		case strings.HasPrefix(path, "/api/student"):
			targetURL = orderKitchenServiceURL
		case strings.HasPrefix(path, "/api/chef"):
			targetURL = orderKitchenServiceURL
		case strings.HasPrefix(path, "/api/analytics"):
			targetURL = aiAnalyticsServiceURL
		case strings.HasPrefix(path, "/api/staff"):
			targetURL = staffLoginServiceURL
		case strings.HasPrefix(path, "/api/manager/orders") || strings.HasPrefix(path, "/api/manager/menu") || strings.HasPrefix(path, "/api/manager/categories") || strings.HasPrefix(path, "/api/manager/uploads"):
			targetURL = orderKitchenServiceURL
		case strings.HasPrefix(path, "/api/manager"):
			targetURL = managerDashboardServiceURL
		default:
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Not Found",
				"message": "Gateway route mapping not found for path: " + path,
			})
			return
		}

		// Execute target reverse proxy handler
		proxy.ReverseProxyHandler(targetURL)(c)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("API Gateway running on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Gateway server start failed: %v", err)
	}
}
