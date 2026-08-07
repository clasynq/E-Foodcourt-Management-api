package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"order-kitchen-service/internal/model"
	"order-kitchen-service/internal/repository"
)

// MenuService houses all core business logic transformations for Categories and Menu items
type MenuService struct {
	repo *repository.MenuRepository
}

// NewMenuService initializes and returns a pointer to a MenuService
func NewMenuService(repo *repository.MenuRepository) *MenuService {
	return &MenuService{repo: repo}
}

// ListCategories delegates category fetching directly to repository
func (s *MenuService) ListCategories(ctx context.Context) ([]model.FoodCategory, error) {
	return s.repo.ListCategories()
}

// ListFoodItems retrieves all foods and categories, mapping IDs to Names in memory
func (s *MenuService) ListFoodItems(ctx context.Context) ([]model.FoodItemResponse, error) {
	items, err := s.repo.ListFoodItems()
	if err != nil {
		return nil, err
	}
	categories, err := s.repo.ListCategories()
	if err != nil {
		return nil, err
	}
	catMap := make(map[string]string)
	for _, c := range categories {
		catMap[c.ID] = c.Name
	}

	var resp []model.FoodItemResponse
	for _, item := range items {
		catName := catMap[item.CategoryID]
		if catName == "" {
			catName = "Uncategorized"
		}
		resp = append(resp, item.ToResponse(catName))
	}
	return resp, nil
}

// CreateFoodItem builds a new FoodItem GORM object and persists it
func (s *MenuService) CreateFoodItem(ctx context.Context, req model.CreateFoodItemRequest) (*model.FoodItemResponse, error) {
	foodID := fmt.Sprintf("food-%d", time.Now().UnixNano())

	item := model.FoodItem{
		ID:                foodID,
		Name:              req.Name,
		Description:       req.Description,
		Price:             req.Price,
		OriginalPrice:     req.OriginalPrice,
		Image:             req.Image,
		CategoryID:        req.CategoryID,
		Type:              req.Type,
		MealTime:          strings.Join(req.MealTime, ","),
		PreparationTime:   req.PreparationTime,
		StockCount:        req.StockCount,
		MaxStock:          req.MaxStock,
		IsAvailable:       req.StockCount > 0,
		IsTodaysSpecial:   req.IsTodaysSpecial,
		IsPopular:         req.IsPopular,
		IsBestSeller:      req.IsBestSeller,
		IsNewlyAdded:      req.IsNewlyAdded,
		Ingredients:       strings.Join(req.Ingredients, ","),
		Allergens:         strings.Join(req.Allergens, ","),
		NutritionCalories: req.NutritionCalories,
		NutritionProtein:  req.NutritionProtein,
		NutritionCarbs:    req.NutritionCarbs,
		NutritionFat:      req.NutritionFat,
		NutritionFiber:    req.NutritionFiber,
		CreatedAt:         time.Now(),
	}

	if err := s.repo.CreateFoodItem(&item); err != nil {
		return nil, err
	}

	categories, _ := s.repo.ListCategories()
	catName := "Uncategorized"
	for _, c := range categories {
		if c.ID == req.CategoryID {
			catName = c.Name
			break
		}
	}

	resp := item.ToResponse(catName)
	return &resp, nil
}

// UpdateFoodItem applies detail changes of an item in the database
func (s *MenuService) UpdateFoodItem(ctx context.Context, id string, req model.CreateFoodItemRequest) (*model.FoodItemResponse, error) {
	updates := map[string]interface{}{
		"name":               req.Name,
		"description":        req.Description,
		"price":              req.Price,
		"original_price":     req.OriginalPrice,
		"image":              req.Image,
		"category_id":        req.CategoryID,
		"type":               req.Type,
		"meal_time":          strings.Join(req.MealTime, ","),
		"preparation_time":   req.PreparationTime,
		"stock_count":        req.StockCount,
		"max_stock":          req.MaxStock,
		"is_available":       req.StockCount > 0,
		"is_todays_special":  req.IsTodaysSpecial,
		"is_popular":         req.IsPopular,
		"is_best_seller":     req.IsBestSeller,
		"is_newly_added":     req.IsNewlyAdded,
		"ingredients":        strings.Join(req.Ingredients, ","),
		"allergens":          strings.Join(req.Allergens, ","),
		"nutrition_calories": req.NutritionCalories,
		"nutrition_protein":  req.NutritionProtein,
		"nutrition_carbs":    req.NutritionCarbs,
		"nutrition_fat":      req.NutritionFat,
		"nutrition_fiber":    req.NutritionFiber,
	}

	if err := s.repo.UpdateFoodItem(id, updates); err != nil {
		return nil, err
	}

	item, err := s.repo.FindFoodByID(id)
	if err != nil {
		return nil, err
	}

	categories, _ := s.repo.ListCategories()
	catName := "Uncategorized"
	for _, c := range categories {
		if c.ID == item.CategoryID {
			catName = c.Name
			break
		}
	}

	resp := item.ToResponse(catName)
	return &resp, nil
}

// UpdateStock updates quantity and auto-sets availability
func (s *MenuService) UpdateStock(ctx context.Context, id string, stock int) error {
	updates := map[string]interface{}{
		"stock_count":  stock,
		"is_available": stock > 0,
	}
	return s.repo.UpdateFoodItem(id, updates)
}

// ToggleAvailability updates availability and handles empty stock count fallbacks
func (s *MenuService) ToggleAvailability(ctx context.Context, id string, isAvailable bool) error {
	item, err := s.repo.FindFoodByID(id)
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"is_available": isAvailable,
	}

	// Auto-fill stock count to 15 if toggled available while empty
	if isAvailable && item.StockCount == 0 {
		updates["stock_count"] = 15
	}

	return s.repo.UpdateFoodItem(id, updates)
}

// DeleteFoodItem deletes a food record
func (s *MenuService) DeleteFoodItem(ctx context.Context, id string) error {
	return s.repo.DeleteFoodItem(id)
}
