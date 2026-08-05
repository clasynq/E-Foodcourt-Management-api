package handler

import (
	"net/http"
	"order-kitchen-service/internal/service"

	"github.com/gin-gonic/gin"
)

// OrderHandler serves as the HTTP entry points for handling Manager Orders API endpoints
type OrderHandler struct {
	srv *service.OrderService
}

// NewOrderHandler initializes and returns a pointer to an OrderHandler
func NewOrderHandler(srv *service.OrderService) *OrderHandler {
	return &OrderHandler{srv: srv}
}

// GetActiveOrders processes GET /api/manager/orders
// It responds with a list of live active orders retrieved from order-kitchen database.
func (h *OrderHandler) GetActiveOrders(c *gin.Context) {
	ctx := c.Request.Context()
	orders, err := h.srv.GetActiveOrders(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to load active orders",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, orders)
}

// UpdateOrderStatus processes PUT /api/manager/orders/:id/status
// It binds the target status from request body and saves the database updates.
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload. 'status' is required"})
		return
	}

	ctx := c.Request.Context()
	err := h.srv.UpdateOrderStatus(ctx, id, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update order status",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order status updated successfully",
		"id":      id,
		"status":  req.Status,
	})
}

