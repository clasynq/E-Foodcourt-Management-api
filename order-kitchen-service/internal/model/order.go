package model

import (
	"strings"
	"time"
)

// Order represents the GORM database schema for student food court orders.
// It maps directly to the "orders" table in the "foodcourt_order" database,
// tracking order details, status lifecycle, total amount, and preparation speed.
type Order struct {
	// ID holds the unique order identifier (e.g. "DS-20260724-0012")
	ID string `gorm:"primaryKey;type:varchar(50)" json:"id"`

	// CustomerName holds the name of the student who placed the order
	CustomerName string `gorm:"type:varchar(100);not null" json:"customer"`

	// Items stores a comma-separated text list of food items ordered (e.g. "Biryani, Soda")
	Items string `gorm:"type:text" json:"items"`

	// ItemsCount holds the count of items in the order
	ItemsCount int `gorm:"not null" json:"itemsCount"`

	// TotalAmount represents the price of the order in INR
	TotalAmount float64 `gorm:"type:decimal(10,2);not null" json:"total"`

	// Status tracks the order stage: PENDING, PREPARING, CONFIRMED, READY, COMPLETED, CANCELLED
	Status string `gorm:"type:varchar(50);not null" json:"status"`

	// Priority represents order urgency: "normal" or "rush"
	Priority string `gorm:"type:varchar(20);not null" json:"priority"`

	// PrepTime estimates or tracks the preparation time in minutes
	PrepTime int `gorm:"not null" json:"prepTime"`

	// CreatedAt marks when the order was placed
	CreatedAt time.Time `json:"createdAt"`
}

// FoodCategory represents a GORM table schema for a food category
type FoodCategory struct {
	ID          string `gorm:"primaryKey;type:varchar(50)" json:"id"`
	Name        string `gorm:"type:varchar(100);not null" json:"name"`
	Slug        string `gorm:"type:varchar(100);uniqueIndex" json:"slug"`
	Icon        string `gorm:"type:varchar(50)" json:"icon"`
	Description string `gorm:"type:text" json:"description"`
	Image       string `gorm:"type:text" json:"image"`
}

// FoodItem represents a GORM table schema for individual dishes in the database
type FoodItem struct {
	ID                string    `gorm:"primaryKey;type:varchar(50)" json:"id"`
	Name              string    `gorm:"type:varchar(100);not null" json:"name"`
	Description       string    `gorm:"type:text" json:"description"`
	Price             float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	OriginalPrice     float64   `gorm:"type:decimal(10,2)" json:"originalPrice"`
	Image             string    `gorm:"type:text" json:"image"`
	CategoryID        string    `gorm:"type:varchar(50);not null" json:"categoryId"`
	Type              string    `gorm:"type:varchar(20);not null" json:"type"`      // "veg", "non-veg", "vegan", "egg"
	MealTime          string    `gorm:"type:varchar(255);not null" json:"mealTime"` // comma-separated list e.g., "lunch,dinner"
	Rating            float64   `gorm:"type:decimal(3,2);default:4.5" json:"rating"`
	TotalRatings      int       `gorm:"default:0" json:"totalRatings"`
	PreparationTime   int       `gorm:"not null" json:"preparationTime"` // in minutes
	IsAvailable       bool      `gorm:"default:true" json:"isAvailable"`
	IsPopular         bool      `gorm:"default:false" json:"isPopular"`
	IsTodaysSpecial   bool      `gorm:"default:false" json:"isTodaysSpecial"`
	IsBestSeller      bool      `gorm:"default:false" json:"isBestSeller"`
	IsNewlyAdded      bool      `gorm:"default:true" json:"isNewlyAdded"`
	IsLiveToday       bool      `gorm:"default:true" json:"isLiveToday"`
	TargetStock       int       `gorm:"default:50" json:"targetStock"`
	PlatesCooked      int       `gorm:"default:0" json:"platesCooked"`
	StockCount        int       `gorm:"not null" json:"stockCount"`
	MaxStock          int       `gorm:"not null" json:"maxStock"`
	SoldCount         int       `gorm:"default:0" json:"soldCount"`
	Ingredients       string    `gorm:"type:text" json:"ingredients"` // comma-separated e.g. "Paneer,Tomato"
	Allergens         string    `gorm:"type:text" json:"allergens"`   // comma-separated e.g. "Dairy,Gluten"
	NutritionCalories int       `json:"nutritionCalories"`
	NutritionProtein  float64   `json:"nutritionProtein"`
	NutritionCarbs    float64   `json:"nutritionCarbs"`
	NutritionFat      float64   `json:"nutritionFat"`
	NutritionFiber    float64   `json:"nutritionFiber"`
	CreatedAt         time.Time `json:"createdAt"`
}

// NutritionInfo represents the sub-object expected by Next.js types
type NutritionInfo struct {
	Calories int     `json:"calories"`
	Protein  float64 `json:"protein"`
	Carbs    float64 `json:"carbs"`
	Fat      float64 `json:"fat"`
	Fiber    float64 `json:"fiber"`
}

// FoodItemResponse matches the React/TypeScript structure of a FoodItem
type FoodItemResponse struct {
	ID              string        `json:"id"`
	Name            string        `json:"name"`
	Description     string        `json:"description"`
	Price           float64       `json:"price"`
	OriginalPrice   float64       `json:"originalPrice,omitempty"`
	Image           string        `json:"image"`
	CategoryID      string        `json:"categoryId"`
	CategoryName    string        `json:"categoryName"`
	Type            string        `json:"type"`
	MealTime        []string      `json:"mealTime"`
	Rating          float64       `json:"rating"`
	TotalRatings    int           `json:"totalRatings"`
	PreparationTime int           `json:"preparationTime"`
	IsAvailable     bool          `json:"isAvailable"`
	IsPopular       bool          `json:"isPopular"`
	IsTodaysSpecial bool          `json:"isTodaysSpecial"`
	IsBestSeller    bool          `json:"isBestSeller"`
	IsNewlyAdded    bool          `json:"isNewlyAdded"`
	IsLiveToday     bool          `json:"isLiveToday"`
	TargetStock     int           `json:"targetStock"`
	PlatesCooked    int           `json:"platesCooked"`
	StockCount      int           `json:"stockCount"`
	MaxStock        int           `json:"maxStock"`
	SoldCount       int           `json:"soldCount"`
	Ingredients     []string      `json:"ingredients"`
	Allergens       []string      `json:"allergens"`
	Nutrition       NutritionInfo `json:"nutrition"`
	CreatedAt       string        `json:"createdAt"`
}

// ToResponse converts a GORM database model into a React-compliant JSON DTO
func (f *FoodItem) ToResponse(categoryName string) FoodItemResponse {
	mealTimes := []string{}
	if f.MealTime != "" {
		for _, part := range strings.Split(f.MealTime, ",") {
			mealTimes = append(mealTimes, strings.TrimSpace(part))
		}
	}

	ingredients := []string{}
	if f.Ingredients != "" {
		for _, part := range strings.Split(f.Ingredients, ",") {
			ingredients = append(ingredients, strings.TrimSpace(part))
		}
	}

	allergens := []string{}
	if f.Allergens != "" {
		for _, part := range strings.Split(f.Allergens, ",") {
			allergens = append(allergens, strings.TrimSpace(part))
		}
	}

	return FoodItemResponse{
		ID:              f.ID,
		Name:            f.Name,
		Description:     f.Description,
		Price:           f.Price,
		OriginalPrice:   f.OriginalPrice,
		Image:           f.Image,
		CategoryID:      f.CategoryID,
		CategoryName:    categoryName,
		Type:            f.Type,
		MealTime:        mealTimes,
		Rating:          f.Rating,
		TotalRatings:    f.TotalRatings,
		PreparationTime: f.PreparationTime,
		IsAvailable:     f.IsAvailable,
		IsPopular:       f.IsPopular,
		IsTodaysSpecial: f.IsTodaysSpecial,
		IsBestSeller:    f.IsBestSeller,
		IsNewlyAdded:    f.IsNewlyAdded,
		IsLiveToday:     f.IsLiveToday,
		TargetStock:     f.TargetStock,
		PlatesCooked:    f.PlatesCooked,
		StockCount:      f.StockCount,
		MaxStock:        f.MaxStock,
		SoldCount:       f.SoldCount,
		Ingredients:     ingredients,
		Allergens:       allergens,
		Nutrition: NutritionInfo{
			Calories: f.NutritionCalories,
			Protein:  f.NutritionProtein,
			Carbs:    f.NutritionCarbs,
			Fat:      f.NutritionFat,
			Fiber:    f.NutritionFiber,
		},
		CreatedAt:       f.CreatedAt.Format(time.RFC3339),
	}
}

// Request payloads for endpoints
type CreateFoodItemRequest struct {
	Name              string   `json:"name" binding:"required"`
	Description       string   `json:"description"`
	Price             float64  `json:"price" binding:"required,gt=0"`
	OriginalPrice     float64  `json:"originalPrice"`
	Image             string   `json:"image"`
	CategoryID        string   `json:"categoryId" binding:"required"`
	Type              string   `json:"type" binding:"required"`
	MealTime          []string `json:"selectedMealTimes"`
	PreparationTime   int      `json:"preparationTime"`
	StockCount        int      `json:"stockCount"`
	MaxStock          int      `json:"maxStock"`
	IsTodaysSpecial   bool     `json:"isTodaysSpecial"`
	IsPopular         bool     `json:"isPopular"`
	IsBestSeller      bool     `json:"isBestSeller"`
	IsNewlyAdded      bool     `json:"isNewlyAdded"`
	IsLiveToday       *bool    `json:"isLiveToday"`
	TargetStock       *int     `json:"targetStock"`
	PlatesCooked      *int     `json:"platesCooked"`
	Ingredients       []string `json:"ingredients"`
	Allergens         []string `json:"allergens"`
	NutritionCalories int      `json:"nutritionCalories"`
	NutritionProtein  float64  `json:"nutritionProtein"`
	NutritionCarbs    float64  `json:"nutritionCarbs"`
	NutritionFat      float64  `json:"nutritionFat"`
	NutritionFiber    float64  `json:"nutritionFiber"`
}

type UpdateStockRequest struct {
	Stock int `json:"stock" binding:"required,min=0"`
}

type UpdateAvailabilityRequest struct {
	IsAvailable bool `json:"isAvailable"`
}

type CreateOrderRequest struct {
	CustomerName string  `json:"customer" binding:"required"`
	Items        string  `json:"items" binding:"required"`
	ItemsCount   int     `json:"itemsCount" binding:"required,min=1"`
	TotalAmount  float64 `json:"total" binding:"required,gt=0"`
	Priority     string  `json:"priority"`
}
