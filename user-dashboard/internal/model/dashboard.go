package model

import "time"

type StudentReward struct {
	UserID    string    `gorm:"type:uuid;primaryKey" json:"user_id"`
	Points    int       `gorm:"type:integer;default:0;not null" json:"points"`
	UpdatedAt time.Time `json:"updated_at"`
}

type StatsCard struct {
	Title  string `json:"title"`
	Value  string `json:"value"`
	Change string `json:"change"`
	Icon   string `json:"icon"`
}

type RecentOrder struct {
	ID        string    `json:"id"`
	Items     string    `json:"items"`
	Total     float64   `json:"total"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type DashboardOverviewResponse struct {
	Stats        []StatsCard   `json:"stats"`
	RecentOrders []RecentOrder `json:"recentOrders"`
}

type AddRewardsRequest struct {
	UserID string `json:"userId" binding:"required"`
	Points int    `json:"points" binding:"required,gt=0"`
}
