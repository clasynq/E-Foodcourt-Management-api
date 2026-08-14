package model

import "time"

// StudentWalletAccount represents a GORM database schema for student RFID/digital wallet accounts.
type StudentWalletAccount struct {
	StudentID  string    `gorm:"primaryKey;type:varchar(50)" json:"studentId"`
	Name       string    `gorm:"type:varchar(100);not null" json:"name"`
	Email      string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	RFIDCardID string    `gorm:"column:rfid_card_id;type:varchar(50);uniqueIndex;not null" json:"rfidCardId"`
	Department string    `gorm:"type:varchar(100)" json:"department"`
	Avatar     string    `gorm:"type:text" json:"avatar"`
	Balance    float64   `gorm:"type:decimal(10,2);not null;default:0.0" json:"balance"`
	Phone      string    `gorm:"type:varchar(20)" json:"phone"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// RechargeRecord represents a GORM database schema for financial recharge logs.
type RechargeRecord struct {
	ID              string    `gorm:"primaryKey;type:varchar(50)" json:"id"`
	StudentID       string    `gorm:"type:varchar(50);not null" json:"studentId"`
	StudentName     string    `gorm:"type:varchar(100)" json:"studentName"`
	StudentEmail    string    `gorm:"type:varchar(100)" json:"studentEmail"`
	RFIDCardID      string    `gorm:"column:rfid_card_id;type:varchar(50)" json:"rfidCardId"`
	Amount          float64   `gorm:"type:decimal(10,2);not null" json:"amount"`
	Method          string    `gorm:"type:varchar(20);not null" json:"method"`      // "nfc" or "manual"
	PaymentType     string    `gorm:"type:varchar(20);not null" json:"paymentType"` // "cash", "upi", "card"
	PreviousBalance float64   `gorm:"type:decimal(10,2);not null" json:"previousBalance"`
	NewBalance      float64   `gorm:"type:decimal(10,2);not null" json:"newBalance"`
	Timestamp       time.Time `json:"timestamp"`
	RechargedBy     string    `gorm:"type:varchar(100)" json:"rechargedBy"`
	Status          string    `gorm:"type:varchar(20);not null" json:"status"`      // "success" or "failed"
}

// NfcRechargeRequest is the request payload for NFC tap-and-pay recharge
type NfcRechargeRequest struct {
	RFIDCardID  string  `json:"rfidCardId" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	PaymentType string  `json:"paymentType" binding:"required"` // "cash", "upi", "card"
	RechargedBy string  `json:"rechargedBy"`
}

// ManualRechargeRequest is the request payload for cashier-initiated manual recharge
type ManualRechargeRequest struct {
	StudentID   string  `json:"studentId" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Email       string  `json:"email" binding:"required,email"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	PaymentType string  `json:"paymentType" binding:"required"` // "cash", "upi", "card"
	RechargedBy string  `json:"rechargedBy"`
}

// DeductBalanceRequest is the request payload for digital wallet deductions (e.g. order checkout)
type DeductBalanceRequest struct {
	Email  string  `json:"email" binding:"required,email"`
	Amount float64 `json:"amount" binding:"required,gt=0"`
}
