package repository

import (
	"errors"
	"time"
	"user-dashboard/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

// GetRewards retrieves the reward details for a user. If not exists, returns default 0 points.
func (r *DashboardRepository) GetRewards(userID string) (*model.StudentReward, error) {
	var reward model.StudentReward
	err := r.db.Where("user_id = ?", userID).First(&reward).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Return default 0 reward points structure
			return &model.StudentReward{
				UserID:    userID,
				Points:    0,
				UpdatedAt: time.Now(),
			}, nil
		}
		return nil, err
	}
	return &reward, nil
}

// AddRewards increments user reward points or inserts a new row if it doesn't exist
func (r *DashboardRepository) AddRewards(userID string, points int) error {
	reward := model.StudentReward{
		UserID:    userID,
		Points:    points,
		UpdatedAt: time.Now(),
	}

	// Use Upsert (on conflict update points = points + excluded.points)
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"points":     gorm.Expr("student_rewards.points + ?", points),
			"updated_at": time.Now(),
		}),
	}).Create(&reward).Error
}
