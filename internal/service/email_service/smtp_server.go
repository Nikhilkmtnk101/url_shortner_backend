package email_service

import (
	"fmt"
	"sync"

	"github.com/nikhil/url-shortner-backend/config"
	"gopkg.in/gomail.v2"
)

type smtpEmailService struct {
	emailConfig config.EmailConfig
}

var (
	instance *smtpEmailService
	once     sync.Once
)

// GetSMTPEmailService returns the singleton instance of smtpEmailService
func GetSMTPEmailService(emailConfig config.EmailConfig) IEmailService {
	once.Do(func() {
		instance = &smtpEmailService{
			emailConfig: emailConfig,
		}
	})
	return instance
}

// SendEmail sends mail in a thread-safe manner
func (e *smtpEmailService) SendEmail(message *gomail.Message) error {
	message.SetHeader("From", e.emailConfig.FromEmail)
	dialer := gomail.Dialer{
		Host:     e.emailConfig.SMTPHost,
		Port:     e.emailConfig.SMTPPort,
		Username: e.emailConfig.SMTPUsername,
		Password: e.emailConfig.SMTPPassword,
	}
	if err := dialer.DialAndSend(message); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}
	return nil
}
