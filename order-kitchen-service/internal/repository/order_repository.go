package repository

import (
	"order-kitchen-service/internal/model"
	"strings"

	"gorm.io/gorm"
)

// OrderRepository handles all database interactions for the Order model using GORM
type OrderRepository struct {
	db *gorm.DB
}

// NewOrderRepository initializes and returns a pointer to an OrderRepository
func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// GetActiveOrders queries PostgreSQL for orders whose status is in:
// "PENDING", "PREPARING", "CONFIRMED", or "READY". It excludes completed/cancelled orders
// and sorts them by creation timestamp descending.
func (r *OrderRepository) GetActiveOrders() ([]model.Order, error) {
	var orders []model.Order
	err := r.db.Where("status IN (?)", []string{"PENDING", "PREPARING", "CONFIRMED", "READY"}).
		Order("created_at desc").
		Find(&orders).Error
	return orders, err
}

// UpdateOrderStatus updates the status string for a specific order identified by ID
func (r *OrderRepository) UpdateOrderStatus(id string, status string) error {
	return r.db.Model(&model.Order{}).Where("id = ?", id).Update("status", status).Error
}

// CreateOrder inserts a new order record into the database
func (r *OrderRepository) CreateOrder(order *model.Order) error {
	return r.db.Create(order).Error
}

// GetOrdersByCustomer queries orders placed by a specific customer (case-insensitive query)
func (r *OrderRepository) GetOrdersByCustomer(customerName string) ([]model.Order, error) {
	var orders []model.Order
	query := r.db.Order("created_at desc")
	if customerName != "" {
		query = query.Where("LOWER(customer_name) = ?", strings.ToLower(customerName))
	}
	err := query.Find(&orders).Error
	return orders, err
}

