package repository

import (
	"strings"
	"time"
	"wallet-service/internal/model"

	"gorm.io/gorm"
)

type WalletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) GetStudentByRfid(rfid string) (*model.StudentWalletAccount, error) {
	var student model.StudentWalletAccount
	err := r.db.Where("UPPER(rfid_card_id) = ?", strings.ToUpper(rfid)).First(&student).Error
	if err != nil {
		return nil, err
	}
	return &student, nil
}

func (r *WalletRepository) GetStudentByEmail(email string) (*model.StudentWalletAccount, error) {
	var student model.StudentWalletAccount
	err := r.db.Where("LOWER(email) = ?", strings.ToLower(email)).First(&student).Error
	if err != nil {
		return nil, err
	}
	return &student, nil
}

func (r *WalletRepository) GetStudentById(id string) (*model.StudentWalletAccount, error) {
	var student model.StudentWalletAccount
	err := r.db.Where("UPPER(student_id) = ?", strings.ToUpper(id)).First(&student).Error
	if err != nil {
		return nil, err
	}
	return &student, nil
}

func (r *WalletRepository) GetStudentByIdOrEmail(query string) (*model.StudentWalletAccount, error) {
	var student model.StudentWalletAccount
	q := strings.ToLower(strings.TrimSpace(query))
	err := r.db.Where("LOWER(student_id) = ? OR LOWER(email) = ? OR LOWER(name) LIKE ?", q, q, "%"+q+"%").First(&student).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound && strings.Contains(q, "@") {
			// Auto-create wallet account with default starter balance for presentation flow
			prefix := strings.Split(q, "@")[0]
			student = model.StudentWalletAccount{
				StudentID:  "STU-" + strings.ToUpper(prefix),
				Name:       strings.Title(strings.ReplaceAll(prefix, ".", " ")),
				Email:      q,
				RFIDCardID: "RFID-" + strings.ToUpper(prefix),
				Balance:    0.00, // Default 0.00 balance
				UpdatedAt:  time.Now(),
			}
			if createErr := r.db.Create(&student).Error; createErr == nil {
				return &student, nil
			}
		}
		return nil, err
	}
	return &student, nil
}

func (r *WalletRepository) UpdateStudentBalance(studentID string, newBalance float64) error {
	return r.db.Model(&model.StudentWalletAccount{}).
		Where("student_id = ?", studentID).
		Updates(map[string]interface{}{
			"balance":    newBalance,
			"updated_at": time.Now(),
		}).Error
}

func (r *WalletRepository) CreateStudentWallet(wallet *model.StudentWalletAccount) error {
	wallet.UpdatedAt = time.Now()
	return r.db.Create(wallet).Error
}

func (r *WalletRepository) CreateRechargeRecord(record *model.RechargeRecord) error {
	return r.db.Create(record).Error
}

func (r *WalletRepository) GetRechargeHistory(studentID string) ([]model.RechargeRecord, error) {
	var history []model.RechargeRecord
	var err error
	if studentID != "" {
		err = r.db.Where("student_id = ?", studentID).Order("timestamp desc").Find(&history).Error
	} else {
		err = r.db.Order("timestamp desc").Find(&history).Error
	}
	return history, err
}

func (r *WalletRepository) GetRechargeSummaryToday() (int, float64, error) {
	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var stats struct {
		Count  int64
		Amount float64
	}

	err := r.db.Model(&model.RechargeRecord{}).
		Select("count(id) as count, coalesce(sum(amount), 0.0) as amount").
		Where("timestamp >= ? AND status = ?", startOfToday, "success").
		Scan(&stats).Error

	return int(stats.Count), stats.Amount, err
}

func (r *WalletRepository) GetNfcRechargesCount() (int, error) {
	var count int64
	err := r.db.Model(&model.RechargeRecord{}).Where("method = ? AND status = ?", "nfc", "success").Count(&count).Error
	return int(count), err
}

func (r *WalletRepository) GetManualRechargesCount() (int, error) {
	var count int64
	err := r.db.Model(&model.RechargeRecord{}).Where("method = ? AND status = ?", "manual", "success").Count(&count).Error
	return int(count), err
}
