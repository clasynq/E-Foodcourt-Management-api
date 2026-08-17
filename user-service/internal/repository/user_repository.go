package repository

import (
	"user-service/internal/model"

	"gorm.io/gorm"
)

// UserRepository defines the standard operations for user data management.
type UserRepository interface {
	Create(user *model.User) error
	FindByEmail(email string) (*model.User, error)
	Update(use *model.User) error
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository returns a new instance of UserRepository.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create inserts a new user record into the database.
func (r *userRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// FindByEmail searches for a user in the database by their email address or username.
func (r *userRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ? OR username = ?", email, email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update an existing user record in the DB
func (r *userRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}
