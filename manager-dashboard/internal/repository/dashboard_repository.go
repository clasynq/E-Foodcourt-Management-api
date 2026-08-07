package repository

import (
	"time"

	"manager-dashboard/internal/model"

	"gorm.io/gorm"
)

type DashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

// GetLowStockItems finds all ingredients where current stock is less than or equal to minimum stock
func (r *DashboardRepository) GetLowStockItems() ([]model.InventoryItem, error) {
	var items []model.InventoryItem
	err := r.db.Where("current_stock <= min_stock").Find(&items).Error
	return items, err
}

// GetPendingAndPreparingOrders retrieves active kitchen orders sorted by creation time
func (r *DashboardRepository) GetPendingAndPreparingOrders() ([]model.LocalOrder, error) {
	var orders []model.LocalOrder
	err := r.db.Where("status = ? OR status = ?", "PENDING", "PREPARING").
		Order("created_at desc").
		Find(&orders).Error
	return orders, err
}

// GetTodayStats aggregates order counts, total revenue, and average preparation times for the current day
func (r *DashboardRepository) GetTodayStats() (int64, float64, float64, error) {
	var count int64
	var revenue float64
	var avgPrep float64

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Today's total order count
	err := r.db.Model(&model.LocalOrder{}).Where("created_at >= ?", startOfDay).Count(&count).Error
	if err != nil {
		return 0, 0, 0, err
	}

	// Today's Revenue (Completed and ready items)
	r.db.Model(&model.LocalOrder{}).
		Where("created_at >= ? AND status != ?", startOfDay, "CANCELLED").
		Select("COALESCE(SUM(total_amount), 0)").
		Row().
		Scan(&revenue)

	// Avg Prep Time for completed orders
	r.db.Model(&model.LocalOrder{}).
		Where("created_at >= ? AND status = ?", startOfDay, "COMPLETED").
		Select("COALESCE(AVG(prep_time), 0)").
		Row().
		Scan(&avgPrep)

	return count, revenue, avgPrep, nil
}

// GetYesterdayStats fetches stats from the previous 24h block to evaluate changes
func (r *DashboardRepository) GetYesterdayStats() (int64, float64, float64, error) {
	var count int64
	var revenue float64
	var avgPrep float64

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startOfYesterday := startOfDay.Add(-24 * time.Hour)

	// Yesterday's Count
	err := r.db.Model(&model.LocalOrder{}).
		Where("created_at >= ? AND created_at < ?", startOfYesterday, startOfDay).
		Count(&count).Error
	if err != nil {
		return 0, 0, 0, err
	}

	// Yesterday's Revenue
	r.db.Model(&model.LocalOrder{}).
		Where("created_at >= ? AND created_at < ? AND status != ?", startOfYesterday, startOfDay, "CANCELLED").
		Select("COALESCE(SUM(total_amount), 0)").
		Row().
		Scan(&revenue)

	// Yesterday's Avg Prep Time
	r.db.Model(&model.LocalOrder{}).
		Where("created_at >= ? AND created_at < ? AND status = ?", startOfYesterday, startOfDay, "COMPLETED").
		Select("COALESCE(AVG(prep_time), 0)").
		Row().
		Scan(&avgPrep)

	return count, revenue, avgPrep, nil
}

// ListInventoryItems fetches all raw kitchen ingredients ordered by name ASC
func (r *DashboardRepository) ListInventoryItems() ([]model.InventoryItem, error) {
	var items []model.InventoryItem
	err := r.db.Order("name asc").Find(&items).Error
	return items, err
}

// FindInventoryByID retrieves an ingredient by its ID
func (r *DashboardRepository) FindInventoryByID(id string) (*model.InventoryItem, error) {
	var item model.InventoryItem
	err := r.db.First(&item, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// CreateInventoryItem inserts a new inventory record
func (r *DashboardRepository) CreateInventoryItem(item *model.InventoryItem) error {
	return r.db.Create(item).Error
}

// UpdateInventoryItem updates specific fields of an ingredient
func (r *DashboardRepository) UpdateInventoryItem(id string, updates map[string]interface{}) error {
	return r.db.Model(&model.InventoryItem{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteInventoryItem deletes an ingredient matching the ID
func (r *DashboardRepository) DeleteInventoryItem(id string) error {
	return r.db.Delete(&model.InventoryItem{}, "id = ?", id).Error
}
