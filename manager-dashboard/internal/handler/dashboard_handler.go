package handler

import (
	"net/http"

	"manager-dashboard/internal/service"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	srv *service.DashboardService
}

func NewDashboardHandler(srv *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{srv: srv}
}

// GetOverview outputs dashboard card stats, pending orders list, and low inventory warnings
func (h *DashboardHandler) GetOverview(c *gin.Context) {
	ctx := c.Request.Context()
	
	resp, err := h.srv.GetOverview(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to load manager dashboard metrics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}
