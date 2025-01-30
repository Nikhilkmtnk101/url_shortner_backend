package email_service

import "gopkg.in/gomail.v2"

type IEmailService interface {
	SendEmail(message *gomail.Message) error
}
