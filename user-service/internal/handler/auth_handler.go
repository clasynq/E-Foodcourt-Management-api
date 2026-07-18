package handler

import (
	"net/http"
	"strings"

	"user-service/internal/model"
	"user-service/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler holds the controllers for user registration and login requests.
type AuthHandler struct {
	svc service.AuthService
}

// NewAuthHandler creates a new instance of AuthHandler.
func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Signup processes registrations.
func (h *AuthHandler) Signup(c *gin.Context) {
	var req model.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.svc.Register(req)
	if err != nil {
		// Capture GORM / PostgreSQL index unique constraint violations (e.g. email/username duplicate)
		errMsg := err.Error()
		if strings.Contains(errMsg, "duplicate key value violates unique constraint") {
			if strings.Contains(errMsg, "email") {
				c.JSON(http.StatusConflict, gin.H{"error": "Email is already registered"})
				return
			}
			if strings.Contains(errMsg, "username") {
				c.JSON(http.StatusConflict, gin.H{"error": "Username is already taken"})
				return
			}
			if strings.Contains(errMsg, "phone") {
				c.JSON(http.StatusConflict, gin.H{"error": "Phone number is already in use"})
				return
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user":    user,
	})
}

// Login validates user credentials and returns a JWT access token.
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, user, err := h.svc.Authenticate(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"user": gin.H{
			"id":       user.ID,
			"name":     user.Name,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}
