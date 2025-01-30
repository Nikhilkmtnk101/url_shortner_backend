package otp_service

import (
	"fmt"
	"github.com/nikhil/url-shortner-backend/internal/service/email_service"
	"gopkg.in/gomail.v2"
)

type EmailOTPService struct {
	emailService email_service.IEmailService
}

func NewOTPService(emailService email_service.IEmailService) IOTPService {
	return &EmailOTPService{
		emailService: emailService,
	}
}

func (o *EmailOTPService) GenerateOTP(email string) string {
	return "{"
}

func (o *EmailOTPService) SendOTP(email string, otp string) error {
	// Configure email
	m := gomail.NewMessage()
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Your OTP Code")

	// Create email body
	emailBody := fmt.Sprintf(`
        <h2>Your OTP Code</h2>
        <p>Your one-time password is: <strong>%s</strong></p>
        <p>This code will expire in 5 minutes.</p>
        <p>If you didn't request this code, please ignore this email.</p>
    `, otp)

	m.SetBody("text/html", emailBody)
	
	return o.emailService.SendEmail(m)
}

func (o *EmailOTPService) VerifyOTP(email string, otp string) error {
	return nil
}
