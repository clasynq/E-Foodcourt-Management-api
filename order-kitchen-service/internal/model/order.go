package model

import (
	"time"
)

// Order represents the GORM database schema for student food court orders.
// It maps directly to the "orders" table in the "foodcourt_order" database,
// tracking order details, status lifecycle, total amount, and preparation speed.
type Order struct {
	// ID holds the unique order identifier (e.g. "DS-20260724-0012")
	ID           string    `gorm:"primaryKey;type:varchar(50)" json:"id"`
	
	// CustomerName holds the name of the student who placed the order
	CustomerName string    `gorm:"type:varchar(100);not null" json:"customer"`
	
	// Items stores a comma-separated text list of food items ordered (e.g. "Biryani, Soda")
	Items        string    `gorm:"type:text" json:"items"`
	
	// ItemsCount holds the count of items in the order
	ItemsCount   int       `gorm:"not null" json:"itemsCount"`
	
	// TotalAmount represents the price of the order in INR
	TotalAmount  float64   `gorm:"type:decimal(10,2);not null" json:"total"`
	
	// Status tracks the order stage: PENDING, PREPARING, CONFIRMED, READY, COMPLETED, CANCELLED
	Status       string    `gorm:"type:varchar(50);not null" json:"status"`
	
	// Priority represents order urgency: "normal" or "rush"
	Priority     string    `gorm:"type:varchar(20);not null" json:"priority"`
	
	// PrepTime estimates or tracks the preparation time in minutes
	PrepTime     int       `gorm:"not null" json:"prepTime"`
	
	// CreatedAt marks when the order was placed
	CreatedAt    time.Time `json:"createdAt"`
}

