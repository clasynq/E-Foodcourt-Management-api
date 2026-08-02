package handler

import (
	"net/http"
	"strings"

	"staff-login/internal/model"
	"staff-login/internal/service"

	"github.com/gin-gonic/gin"
)

type StaffHandler struct {
	svc service.StaffService
}

func NewStaffHandler(svc service.StaffService) *StaffHandler {
	return &StaffHandler{svc: svc}
}

func (h *StaffHandler) InitiateLogin(c *gin.Context) {
	var req model.InitiateLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.InitiateLogin(req); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Staff credentials verified. 2FA verification code has been dispatched.",
	})
}

func (h *StaffHandler) VerifyLogin(c *gin.Context) {
	var req model.VerifyLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, member, err := h.svc.VerifyLogin(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Staff login successful",
		"token":   token,
		"user": gin.H{
			"id":       member.ID,
			"name":     member.Name,
			"email":    member.Email,
			"role":     member.Role,
			"is_active": member.IsActive,
		},
	})
}

func (h *StaffHandler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header is required"})
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Authorization header format"})
		return
	}
	tokenString := parts[1]

	err := h.svc.BlacklistToken(tokenString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log out: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Staff logout successful"})
}
