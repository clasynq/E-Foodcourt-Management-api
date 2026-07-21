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

	senderAddress := user
	replyToHeader := ""
	if from != "" {
		replyToHeader = fmt.Sprintf("Reply-To: %s\r\n", from)
	}

	// Construct Email Message (Headers + Body)
	subject := "Subject: Smart Food Court - Your Sign-up Verification OTP\r\n"
	contentType := "MIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n"
	fromHeader := fmt.Sprintf("From: Smart Food Court <%s>\r\n", senderAddress)
	toHeader := fmt.Sprintf("To: %s\r\n", toEmail)

	body := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<body style="font-family: Arial, sans-serif; background-color: #f4f4f4; padding: 20px;">
			<div style="max-width: 500px; margin: 0 auto; background: #ffffff; padding: 30px; border-radius: 8px;">
				<h2 style="color: #333333; text-align: center;">Food Court Verification Code</h2>
				<p style="color: #666666;">Hello,</p>
				<p style="color: #666666;">Your one-time security verification code (OTP) is:</p>
				<div style="background: #eef2ff; color: #4f46e5; font-size: 32px; font-weight: bold; text-align: center; padding: 15px; border-radius: 6px; letter-spacing: 5px; margin: 20px 0;">
					%s
				</div>
				<p style="color: #888888; font-size: 13px;">This OTP is valid for <strong>5 minutes</strong>. Do not share this code with anyone.</p>
			</div>
		</body>
		</html>
	`, otpCode)

	msg := []byte(fromHeader + toHeader + replyToHeader + subject + contentType + "\r\n" + body)
	addr := fmt.Sprintf("%s:%d", host, port)

	return smtp.SendMail(addr, auth, senderAddress, []string{toEmail}, msg)
}
