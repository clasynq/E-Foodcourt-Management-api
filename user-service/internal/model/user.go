package model

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	ID        string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	Username  string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Email     string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Phone     string    `gorm:"type:varchar(20);uniqueIndex;not null" json:"phone"`
	Password  string    `gorm:"type:varchar(255);not null" json:"-"`
	Role      string    `gorm:"type:varchar(20);default:CUSTOMER" json:"role"`
	CardUID   *string   `gorm:"type:varchar(100);uniqueIndex" json:"card_uid,omitempty"`
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

type SendOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required,len=6,numeric"`
}
