package model

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	ID        string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Username  string    `gorm:"uniqueIndex;not null" json:"username"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Phone     string    `gorm:"uniqueIndex;not null" json:"phone"` // <-- ADDED PHONE COLUMN
	Password  string    `gorm:"not null" json:"-"`                 // Hidden from JSON responses
	Role      string    `gorm:"default:CUSTOMER" json:"role"`
	CardUID   *string   `gorm:"uniqueIndex" json:"card_uid,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type SignupRequest struct {
	Name     string `json:"name" binding:"required,min=2"`
	Username string `json:"username" binding:"required,alphanum,min=3"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"required,numeric,min=10,max=15"` // <-- ADDED THIS FIELD (only numbers allowed)
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}
