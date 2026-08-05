package model

import (
	"time"
)

// InventoryItem represents raw kitchen ingredients
type InventoryItem struct {
	ID            string    `gorm:"primaryKey;type:varchar(50)" json:"id"`
	Name          string    `gorm:"type:varchar(100);not null" json:"name"`
	Category      string    `gorm:"type:varchar(100)" json:"category"`
	CurrentStock  float64   `gorm:"type:decimal(10,2);not null" json:"currentStock"`
	MinStock      float64   `gorm:"type:decimal(10,2);not null" json:"minStock"`
	MaxStock      float64   `gorm:"type:decimal(10,2);not null" json:"maxStock"`
	Unit          string    `gorm:"type:varchar(20);not null" json:"unit"`
	Status        string    `gorm:"type:varchar(20)" json:"status"` // "In Stock", "Low Stock", "Out of Stock"
	LastRestocked time.Time `json:"lastRestocked"`
	Notes         string    `gorm:"type:text" json:"notes"`
}

// LocalOrder simulates order placement for metrics calculation
type LocalOrder struct {
	ID           string    `gorm:"primaryKey;type:varchar(50)" json:"id"`
	CustomerName string    `gorm:"type:varchar(100);not null" json:"customer"`
	ItemsCount   int       `gorm:"not null" json:"items"`
	TotalAmount  float64   `gorm:"type:decimal(10,2);not null" json:"total"`
	Status       string    `gorm:"type:varchar(50);not null" json:"status"` // "PENDING", "PREPARING", "READY", "COMPLETED"
	Priority     string    `gorm:"type:varchar(20);not null" json:"priority"` // "normal", "rush"
	PrepTime     int       `gorm:"not null" json:"prepTime"` // in minutes
	CreatedAt    time.Time `json:"createdAt"`
}

// StatsCard represents a widget block on the dashboard UI
type StatsCard struct {
	Title     string `json:"title"`
	Value     string `json:"value"`
	Change    string `json:"change"`
	Icon      string `json:"icon"`
	BgColor   string `json:"bgColor"`
	TextColor string `json:"textColor"`
}

// LowInventoryAlert contains low stock formatting
type LowInventoryAlert struct {
	Message string   `json:"message"`
	Items   []string `json:"items"`
}

// ManagerOverviewResponse is the complete payload for GET /api/manager/overview
type ManagerOverviewResponse struct {
	Stats             []StatsCard        `json:"stats"`
	PendingOrders     []LocalOrder       `json:"pendingOrders"`
	LowInventoryAlert *LowInventoryAlert `json:"lowInventoryAlert,omitempty"`
}
