package utils

import (
	"fmt"
	"net/smtp"
	"os"
	"strconv"
)

// SendStaffOTPEmail sends a styled OTP email for staff 2FA authentication
func SendStaffOTPEmail(toEmail, otpCode, role string) error {
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

	subject := fmt.Sprintf("Subject: Smart Food Court - Staff (%s) 2FA Verification\r\n", role)
	contentType := "Content-Type: text/plain; charset=UTF-8\r\n"
	fromHeader := fmt.Sprintf("From: Smart Food Court Staff Portal <%s>\r\n", senderAddress)
	toHeader := fmt.Sprintf("To: %s\r\n", toEmail)

	body := fmt.Sprintf("Hello,\n\nYou are attempting to access the DineSynq %s Portal.\n\nYour 2-Factor Authentication (2FA) verification code is: %s\n\nThis code is valid for 5 minutes. If you did not initiate this login request, please change your credentials immediately.\n", role, otpCode)

	msg := []byte(fromHeader + toHeader + subject + contentType + "\r\n" + body)
	addr := fmt.Sprintf("%s:%d", host, port)

	return smtp.SendMail(addr, auth, senderAddress, []string{toEmail}, msg)
}
