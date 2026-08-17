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

	var account StudentWalletAccount
	err = db.Table("student_wallet_accounts").Where("email = ?", "surajitsutradhar010@gmail.com").First(&account).Error
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Found DB Row for surajit:\n")
	fmt.Printf("  StudentID: %s\n", account.StudentID)
	fmt.Printf("  Name:      %s\n", account.Name)
	fmt.Printf("  Email:     %s\n", account.Email)
	fmt.Printf("  Balance:   %.2f\n", account.Balance)
}
