package repository

import (
	"order-kitchen-service/internal/model"

	"gorm.io/gorm"
)

// MenuRepository handles PostgreSQL GORM database CRUD queries for Category and FoodItem models
type MenuRepository struct {
	db *gorm.DB
}

// NewMenuRepository initializes and returns a pointer to a MenuRepository
func NewMenuRepository(db *gorm.DB) *MenuRepository {
	return &MenuRepository{db: db}
}

// ListCategories fetches all food categories from the database
func (r *MenuRepository) ListCategories() ([]model.FoodCategory, error) {
	var categories []model.FoodCategory
	err := r.db.Find(&categories).Error
	return categories, err
}

// ListFoodItems fetches all food items ordered by creation time descending
func (r *MenuRepository) ListFoodItems() ([]model.FoodItem, error) {
	var items []model.FoodItem
	err := r.db.Order("created_at desc").Find(&items).Error
	return items, err
}

// FindFoodByID searches the database for a specific food item by its unique ID
func (r *MenuRepository) FindFoodByID(id string) (*model.FoodItem, error) {
	var item model.FoodItem
	err := r.db.First(&item, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// CreateFoodItem inserts a new food item record into the database
func (r *MenuRepository) CreateFoodItem(item *model.FoodItem) error {
	return r.db.Create(item).Error
}

// UpdateFoodItem applies modifications to a food item using map variables
func (r *MenuRepository) UpdateFoodItem(id string, updates map[string]interface{}) error {
	return r.db.Model(&model.FoodItem{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteFoodItem deletes a specific food item record matching the ID
func (r *MenuRepository) DeleteFoodItem(id string) error {
	return r.db.Delete(&model.FoodItem{}, "id = ?", id).Error
}

// ListStudentFoodItems fetches all food items that are available and live today
func (r *MenuRepository) ListStudentFoodItems() ([]model.FoodItem, error) {
	var items []model.FoodItem
	err := r.db.Where("is_available = ? AND is_live_today = ?", true, true).Order("created_at desc").Find(&items).Error
	return items, err
}
