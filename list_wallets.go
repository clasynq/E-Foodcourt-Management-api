package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type StudentWalletAccount struct {
	StudentID  string  `gorm:"primaryKey;column:student_id"`
	Name       string  `gorm:"column:name"`
	Email      string  `gorm:"column:email"`
	Balance    float64 `gorm:"column:balance"`
}

func main() {
	connStr := "postgres://postgres:suro1234@localhost:5432/foodcourt_wallet?sslmode=disable"
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	var accounts []StudentWalletAccount
	err = db.Table("student_wallet_accounts").Find(&accounts).Error
	if err != nil {
		log.Fatalf("Error querying wallets: %v", err)
	}

	fmt.Printf("Total Wallets: %d\n", len(accounts))
	for _, acc := range accounts {
		fmt.Printf("  - ID: %s, Name: %s, Email: %s, Balance: %.2f\n", acc.StudentID, acc.Name, acc.Email, acc.Balance)
	}
}
