package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"api-gateway/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT structure signed by user-service
type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// CachedUser represents the JSON schema cached in Redis under user:id:<userID>
type CachedUser struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// CORSMiddleware provides cross-origin headers configuration
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-User-Id, X-User-Role")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// AuthMiddleware intercepts requests to parse JWT, check Redis blacklist, and inject header information.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Bypass auth for public auth endpoints and options requests
		if c.Request.Method == "OPTIONS" ||
			path == "/api/auth/login" ||
			path == "/api/auth/signup" ||
			path == "/api/auth/send-otp" ||
			path == "/api/auth/verify-otp" ||
			path == "/api/auth/forgot-password" ||
			path == "/api/auth/reset-password" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must be Bearer token"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 1. Check if token is blacklisted in Redis
		ctx := context.Background()
		isBlacklisted, err := config.RDB.Exists(ctx, "blacklist:"+tokenString).Result()
		if err == nil && isBlacklisted > 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has been invalidated (logged out)"})
			c.Abort()
			return
		}

		// 2. Parse token using SECRECT_KEY
		secretKey := os.Getenv("SECRECT_KEY")
		if secretKey == "" {
			// Fallback to JWT_SECRET if SECRECT_KEY is not set
			secretKey = os.Getenv("JWT_SECRET")
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token", "details": err.Error()})
			c.Abort()
			return
		}

		// 3. Inject standard identity headers
		c.Request.Header.Set("X-User-Id", claims.UserID)
		c.Request.Header.Set("X-User-Role", claims.Role)

		// 4. Try retrieving username/email from Redis cache to enrich headers
		userCacheKey := "user:id:" + claims.UserID
		cachedUserData, err := config.RDB.Get(ctx, userCacheKey).Result()
		if err == nil && cachedUserData != "" {
			var user CachedUser
			if err := json.Unmarshal([]byte(cachedUserData), &user); err == nil {
				c.Request.Header.Set("X-User-Name", user.Name)
				c.Request.Header.Set("X-User-Email", user.Email)
			}
		}

		c.Next()
	}
}
