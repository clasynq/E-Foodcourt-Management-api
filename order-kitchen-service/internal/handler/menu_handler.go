package handler

import (
	"net/http"

	"order-kitchen-service/internal/model"
	"order-kitchen-service/internal/service"

	"github.com/gin-gonic/gin"
)

// MenuHandler receives request inputs for Category and FoodItem CRUD endpoints
type MenuHandler struct {
	srv *service.MenuService
}

// NewMenuHandler initializes and returns a pointer to a MenuHandler
func NewMenuHandler(srv *service.MenuService) *MenuHandler {
	return &MenuHandler{srv: srv}
}

// ListCategories processes GET /api/manager/categories
func (h *MenuHandler) ListCategories(c *gin.Context) {
	ctx := c.Request.Context()
	categories, err := h.srv.ListCategories(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
		return
	}
	c.JSON(http.StatusOK, categories)
}

// ListFoodItems processes GET /api/manager/menu
func (h *MenuHandler) ListFoodItems(c *gin.Context) {
	ctx := c.Request.Context()
	items, err := h.srv.ListFoodItems(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch menu items", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// CreateFoodItem processes POST /api/manager/menu
func (h *MenuHandler) CreateFoodItem(c *gin.Context) {
	var req model.CreateFoodItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	resp, err := h.srv.CreateFoodItem(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create food item", "details": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// UpdateFoodItem processes PUT /api/manager/menu/:id
func (h *MenuHandler) UpdateFoodItem(c *gin.Context) {
	id := c.Param("id")
	var req model.CreateFoodItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	resp, err := h.srv.UpdateFoodItem(ctx, id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update food item", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// UpdateStock processes PUT /api/manager/menu/:id/stock
func (h *MenuHandler) UpdateStock(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdateStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	err := h.srv.UpdateStock(ctx, id, req.Stock)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update stock", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Stock count updated successfully"})
}

// ToggleAvailability processes PUT /api/manager/menu/:id/availability
func (h *MenuHandler) ToggleAvailability(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdateAvailabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	err := h.srv.ToggleAvailability(ctx, id, req.IsAvailable)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update availability", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Availability updated successfully"})
}

// DeleteFoodItem processes DELETE /api/manager/menu/:id
func (h *MenuHandler) DeleteFoodItem(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()
	err := h.srv.DeleteFoodItem(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete item", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Menu item deleted successfully"})
}
