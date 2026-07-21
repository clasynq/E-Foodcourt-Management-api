package utils

import (
	"fmt"
	"net/smtp"
	"os"
	"strconv"
)

// SendOTPEmail sends a styled HTML OTP email using Brevo SMTP configurations from .env
func SendOTPEmail(toEmail, otpCode string) error {
	host := os.Getenv("EMAIL_HOST")
	portStr := os.Getenv("EMAIL_PORT")
	user := os.Getenv("EMAIL_HOST_USER")
	pass := os.Getenv("EMAIL_HOST_PASSWORD")
	from := os.Getenv("DEFAULT_FROM_EMAIL")

	if host == "" || user == "" || pass == "" {
		return fmt.Errorf("SMTP settings are incomplete in .env file")
	}

	port := 587
	if p, err := strconv.Atoi(portStr); err == nil && p > 0 {
		port = p
	}

	auth := smtp.PlainAuth("", user, pass, host)

	senderAddress := from
	if senderAddress == "" {
		senderAddress = user
	}

	// Construct Email Message (Headers + Body)
	subject := "Subject: Smart Food Court - Your Verification OTP\r\n"
	contentType := "Content-Type: text/plain; charset=UTF-8\r\n"
	fromHeader := fmt.Sprintf("From: Smart Food Court <%s>\r\n", senderAddress)
	toHeader := fmt.Sprintf("To: %s\r\n", toEmail)

	body := fmt.Sprintf("Hello,\n\nYour Smart Food Court one-time verification code (OTP) is: %s\n\nThis OTP is valid for 5 minutes. Do not share this code with anyone.\n", otpCode)

	msg := []byte(fromHeader + toHeader + subject + contentType + "\r\n" + body)
	addr := fmt.Sprintf("%s:%d", host, port)

	return smtp.SendMail(addr, auth, senderAddress, []string{toEmail}, msg)
}
