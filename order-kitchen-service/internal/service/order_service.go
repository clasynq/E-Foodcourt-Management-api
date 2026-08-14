package service

import (
	"context"
	"fmt"
	"order-kitchen-service/internal/model"
	"order-kitchen-service/internal/repository"
	"strings"
	"time"
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

// UpdateOrderStatus delegates the order status update query to the repository layer after normalizing the status
func (s *OrderService) UpdateOrderStatus(ctx context.Context, id string, status string) error {
	normalized := NormalizeStatus(status)
	return s.repo.UpdateOrderStatus(id, normalized)
}

// NormalizeStatus maps frontend statuses (or lowercase equivalents) to the standard backend uppercase statuses
func NormalizeStatus(status string) string {
	switch strings.ToLower(status) {
	case "queued", "pending":
		return "PENDING"
	case "cooking", "preparing":
		return "PREPARING"
	case "ready":
		return "READY"
	case "picked-up", "completed":
		return "COMPLETED"
	default:
		return strings.ToUpper(status)
	}
}

// CreateOrder creates a new order in the database, generating a unique ID and initializing default status
func (s *OrderService) CreateOrder(ctx context.Context, req model.CreateOrderRequest) (*model.Order, error) {
	// Generate order ID like DS-XXXX where XXXX is randomized 5 digits
	orderID := fmt.Sprintf("DS-%05d", time.Now().UnixNano()%100000)

	// Calculate default prep time: standard 10 minutes
	prepTime := 10

	order := model.Order{
		ID:           orderID,
		CustomerName: req.CustomerName,
		Items:        req.Items,
		ItemsCount:   req.ItemsCount,
		TotalAmount:  req.TotalAmount,
		Status:       "PENDING",
		Priority:     req.Priority,
		PrepTime:     prepTime,
		CreatedAt:    time.Now(),
	}

	if order.Priority == "" {
		order.Priority = "normal"
	}

	err := s.repo.CreateOrder(&order)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// GetOrdersByCustomer fetches orders for a specific customer
func (s *OrderService) GetOrdersByCustomer(ctx context.Context, customerName string) ([]model.Order, error) {
	return s.repo.GetOrdersByCustomer(customerName)
}

