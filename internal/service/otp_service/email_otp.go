package otp_service

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/nikhil/url-shortner-backend/internal/middleware/logger"
	"github.com/nikhil/url-shortner-backend/internal/repository"
	"github.com/nikhil/url-shortner-backend/internal/service/email_service"
	"gopkg.in/gomail.v2"
	"time"
)

type EmailOTPService struct {
	emailService email_service.IEmailService
	otpRepo      repository.IOTPRepository
}

func NewOTPService(emailService email_service.IEmailService, otpRepo repository.IOTPRepository) IOTPService {
	return &EmailOTPService{
		emailService: emailService,
		otpRepo:      otpRepo,
	}
}

// GenerateOTP generates a 6-digit OTP based on the email and current timestamp
func (o *EmailOTPService) GenerateOTP(email string) string {
	// Get the current timestamp as a string
	currentTimestamp := time.Now().Format(time.RFC3339)

	// Combine the email with the current timestamp
	combinedSeed := email + currentTimestamp

	// Hash the combined seed using SHA-256
	hash := sha256.New()
	hash.Write([]byte(combinedSeed))
	hashedSeed := hash.Sum(nil)

	// Use part of the hash as a numeric seed for generating OTP
	seed := binary.BigEndian.Uint32(hashedSeed[:4])

	// Generate the 6-digit OTP from the seed
	otp := seed % 1000000 // Ensure it is a 6-digit number

	// Return the OTP, padded with leading zeros if necessary
	return fmt.Sprintf("%06d", otp)
}

func (o *EmailOTPService) SendOTP(email string, otp string) error {
	// Configure email
	m := gomail.NewMessage()
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Your OTP Code")

	// Create email body with styling
	emailBody := fmt.Sprintf(`
    <html>
    <head>
        <style>
            body {
                font-family: Arial, sans-serif;
                background-color: #f4f7fa;
                color: #333;
                padding: 20px;
                margin: 0;
            }
            .container {
                background-color: #ffffff;
                border-radius: 8px;
                box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
                padding: 30px;
                max-width: 600px;
                margin: 20px auto;
            }
            h2 {
                color: #2d9cdb;
                text-align: center;
            }
            p {
                font-size: 16px;
                line-height: 1.6;
                text-align: center;
            }
            .otp-code {
                font-size: 24px;
                font-weight: bold;
                color: #2d9cdb;
                background-color: #f1f8ff;
                padding: 10px 20px;
                border-radius: 6px;
                display: inline-block;
            }
            .footer {
                font-size: 12px;
                color: #777;
                text-align: center;
                margin-top: 20px;
            }
            .footer a {
                color: #2d9cdb;
                text-decoration: none;
            }
        </style>
    </head>
    <body>
        <div class="container">
            <h2>Your OTP Code</h2>
            <p>Your one-time password is:</p>
            <p class="otp-code">%s</p>
            <p>This code will expire in 5 minutes.</p>
            <p>If you didn't request this code, please ignore this email.</p>
            <div class="footer">
                <p>For security reasons, never share your OTP with anyone.</p>
                <p><a href="nikhilkmtnk29@gmail.com">Contact Support</a> if you need help.</p>
            </div>
        </div>
    </body>
    </html>
`, otp)

	m.SetBody("text/html", emailBody)

	return o.emailService.SendEmail(m)
}

func (o *EmailOTPService) SaveOTP(ctx *gin.Context, email string, otp string) error {
	return o.otpRepo.SaveOTP(ctx, email, otp)
}

func (o *EmailOTPService) VerifyOTP(ctx *gin.Context, email string, otp string) error {
	log := logger.GetLogger(ctx)
	cachedOTP, err := o.otpRepo.GetOTP(ctx, email)
	if err != nil {
		log.Errorf("Failed to verify OTP: %v", err)
		return err
	}
	if cachedOTP != otp {
		log.Errorf("Wrong OTP entered")
		return errors.New("wrong OTP")
	}
	return nil
}

func (o *EmailOTPService) DeleteOTP(ctx *gin.Context, email string) error {
	return o.otpRepo.DeleteOTP(ctx, email)
}
