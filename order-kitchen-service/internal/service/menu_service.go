package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

	isLiveToday := true
	if req.IsLiveToday != nil {
		isLiveToday = *req.IsLiveToday
	}
	targetStock := req.MaxStock
	if req.TargetStock != nil {
		targetStock = *req.TargetStock
	} else if targetStock == 0 {
		targetStock = 50
	}
	platesCooked := 0
	if req.PlatesCooked != nil {
		platesCooked = *req.PlatesCooked
	}

	savedImage, err := saveBase64Image(req.Image)
	if err != nil {
		savedImage = req.Image
	}

	item := model.FoodItem{
		ID:                foodID,
		Name:              req.Name,
		Description:       req.Description,
		Price:             req.Price,
		OriginalPrice:     req.OriginalPrice,
		Image:             savedImage,
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
		IsLiveToday:       isLiveToday,
		TargetStock:       targetStock,
		PlatesCooked:      platesCooked,
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
	savedImage, err := saveBase64Image(req.Image)
	if err != nil {
		savedImage = req.Image
	}

	updates := map[string]interface{}{
		"name":               req.Name,
		"description":        req.Description,
		"price":              req.Price,
		"original_price":     req.OriginalPrice,
		"image":              savedImage,
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

	if req.IsLiveToday != nil {
		updates["is_live_today"] = *req.IsLiveToday
	}
	if req.TargetStock != nil {
		updates["target_stock"] = *req.TargetStock
	}
	if req.PlatesCooked != nil {
		updates["plates_cooked"] = *req.PlatesCooked
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

// UpdatePrep updates the plates cooked and target stock of a menu item, auto-adjusting stock and availability
func (s *MenuService) UpdatePrep(ctx context.Context, id string, updates map[string]interface{}) error {
	// If plates_cooked is updated, let's adjust stock_count and is_available just like the store does!
	if cookedVal, ok := updates["plates_cooked"].(int); ok {
		item, err := s.repo.FindFoodByID(id)
		if err == nil {
			target := item.TargetStock
			if tVal, ok2 := updates["target_stock"].(int); ok2 {
				target = tVal
			}
			if target <= 0 {
				target = 9999
			}
			cooked := cookedVal
			if cooked > target {
				cooked = target
			}
			if cooked < 0 {
				cooked = 0
			}
			updates["plates_cooked"] = cooked

			diff := cooked - item.PlatesCooked
			newStock := item.StockCount + diff
			if newStock < 0 {
				newStock = 0
			}
			updates["stock_count"] = newStock
			updates["is_available"] = newStock > 0 && cooked < target
		}
	}
	return s.repo.UpdateFoodItem(id, updates)
}

// ToggleLiveToday toggles the live status of a menu item
func (s *MenuService) ToggleLiveToday(ctx context.Context, id string, isLiveToday bool) error {
	updates := map[string]interface{}{
		"is_live_today": isLiveToday,
	}
	return s.repo.UpdateFoodItem(id, updates)
}

// ListStudentFoodItems fetches available and live menu items mapped as response DTOs
func (s *MenuService) ListStudentFoodItems(ctx context.Context) ([]model.FoodItemResponse, error) {
	items, err := s.repo.ListStudentFoodItems()
	if err != nil {
		return nil, err
	}

	categories, _ := s.repo.ListCategories()
	catMap := make(map[string]string)
	for _, c := range categories {
		catMap[c.ID] = c.Name
	}

	var resp []model.FoodItemResponse
	for _, item := range items {
		catName := "Uncategorized"
		if name, ok := catMap[item.CategoryID]; ok {
			catName = name
		}
		resp = append(resp, item.ToResponse(catName))
	}
	return resp, nil
}

func saveBase64Image(base64Str string) (string, error) {
	if !strings.HasPrefix(base64Str, "data:image/") {
		return base64Str, nil
	}

	parts := strings.Split(base64Str, ",")
	if len(parts) != 2 {
		return "", errors.New("invalid base64 image format")
	}

	header := parts[0]
	body := parts[1]

	ext := ".png"
	if strings.Contains(header, "image/jpeg") || strings.Contains(header, "image/jpg") {
		ext = ".jpg"
	} else if strings.Contains(header, "image/gif") {
		ext = ".gif"
	} else if strings.Contains(header, "image/webp") {
		ext = ".webp"
	}

	dec, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		return "", err
	}

	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return "", err
	}

	fileName := fmt.Sprintf("img-%d%s", time.Now().UnixNano(), ext)
	filePath := filepath.Join(uploadDir, fileName)

	if err := os.WriteFile(filePath, dec, 0644); err != nil {
		return "", err
	}

	return fmt.Sprintf("/api/manager/uploads/%s", fileName), nil
}
