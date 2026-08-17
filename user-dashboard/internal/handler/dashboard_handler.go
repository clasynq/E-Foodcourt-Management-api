package handler

import (
	"log"
	"net/http"
	"user-dashboard/internal/model"
	"user-dashboard/internal/service"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	srv *service.DashboardService
}

func NewDashboardHandler(srv *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{srv: srv}
}

// GetOverview outputs dashboard statistics card and recent orders list
func (h *DashboardHandler) GetOverview(c *gin.Context) {
	log.Printf("[Dashboard Debug] Headers: X-User-Id='%s', X-User-Name='%s', X-User-Email='%s'", 
		c.GetHeader("X-User-Id"), c.GetHeader("X-User-Name"), c.GetHeader("X-User-Email"))

	userID := c.Query("userId")
	if userID == "" {
		userID = c.GetHeader("X-User-Id")
	}

	customerName := c.Query("customer")
	if customerName == "" {
		customerName = c.GetHeader("X-User-Name")
	}

	userEmail := c.Query("email")
	if userEmail == "" {
		userEmail = c.GetHeader("X-User-Email")
	}

	// Fallback values for test environments
	if userID == "" {
		userID = "user-default-stu-001"
	}
	if customerName == "" {
		customerName = "Arindam Roy"
	}
	if userEmail == "" {
		userEmail = "arindam@dinesynq.com"
	}

	ctx := c.Request.Context()
	resp, err := h.srv.GetOverview(ctx, userID, customerName, userEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to load student dashboard overview",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// AddRewards updates or inserts student reward points balance
func (h *DashboardHandler) AddRewards(c *gin.Context) {
	var req model.AddRewardsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	err := h.srv.AddRewards(ctx, req.UserID, req.Points)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update student rewards",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Student reward points updated successfully",
		"userId":  req.UserID,
		"points":  req.Points,
	})
}
