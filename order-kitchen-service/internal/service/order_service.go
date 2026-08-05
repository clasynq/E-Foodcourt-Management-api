package service

import (
	"context"
	"order-kitchen-service/internal/model"
	"order-kitchen-service/internal/repository"
)

// OrderService contains business logic operations for processing order status tracking
type OrderService struct {
	repo *repository.OrderRepository
}

// NewOrderService initializes and returns a pointer to an OrderService
func NewOrderService(repo *repository.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

// GetActiveOrders manages context boundaries and delegates to the repository layer to load open order flows
func (s *OrderService) GetActiveOrders(ctx context.Context) ([]model.Order, error) {
	return s.repo.GetActiveOrders()
}

// UpdateOrderStatus delegates the order status update query to the repository layer
func (s *OrderService) UpdateOrderStatus(ctx context.Context, id string, status string) error {
	return s.repo.UpdateOrderStatus(id, status)
}

