package handler

import (
	"net/http"

	"manager-dashboard/internal/model"
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

// ListInventoryItems returns all raw kitchen ingredients
func (h *DashboardHandler) ListInventoryItems(c *gin.Context) {
	ctx := c.Request.Context()
	items, err := h.srv.ListInventoryItems(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load inventory items"})
		return
	}
	c.JSON(http.StatusOK, items)
}

// CreateInventoryItem inserts a new ingredient record
func (h *DashboardHandler) CreateInventoryItem(c *gin.Context) {
	var req model.InventoryItem
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	resp, err := h.srv.CreateInventoryItem(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create inventory item", "details": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// UpdateInventoryItem updates details of an ingredient
func (h *DashboardHandler) UpdateInventoryItem(c *gin.Context) {
	id := c.Param("id")
	var req model.InventoryItem
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	resp, err := h.srv.UpdateInventoryItem(ctx, id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update inventory item", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// RestockInventoryItem handles a restocking event increasing the stock count
func (h *DashboardHandler) RestockInventoryItem(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Amount float64 `json:"amount" binding:"required,gt=0"`
		Notes  string  `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload. 'amount' greater than 0 is required"})
		return
	}

	ctx := c.Request.Context()
	resp, err := h.srv.RestockInventoryItem(ctx, id, req.Amount, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restock item", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// DeleteInventoryItem removes an ingredient
func (h *DashboardHandler) DeleteInventoryItem(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()
	err := h.srv.DeleteInventoryItem(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete item", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Inventory item deleted successfully"})
}
