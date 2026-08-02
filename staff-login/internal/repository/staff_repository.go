package repository

import (
	"staff-login/internal/model"

	"gorm.io/gorm"
)

type StaffRepository interface {
	Create(member *model.StaffMember) error
	FindByEmail(email string) (*model.StaffMember, error)
}

type staffRepository struct {
	db *gorm.DB
}

func NewStaffRepository(db *gorm.DB) StaffRepository {
	return &staffRepository{db: db}
}

func (r *staffRepository) Create(member *model.StaffMember) error {
	return r.db.Create(member).Error
}

func (r *staffRepository) FindByEmail(email string) (*model.StaffMember, error) {
	var member model.StaffMember
	err := r.db.Where("email = ?", email).First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}
