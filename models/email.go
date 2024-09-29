package models

import (
	"fmt"

	"github.com/go-mail/mail/v2"
)

const (
	DefaultSender = "support@lenslocked.com"
)

type EmailService struct {
	DefaultSender string
	dialer        *mail.Dialer
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

type Email struct {
	To        string
	From      string
	Subject   string
	Plaintext string
	Html      string
}

func NewEmailService(config SMTPConfig) *EmailService {
	es := EmailService{
		dialer: mail.NewDialer(config.Host, config.Port, config.Username, config.Password),
	}
	return &es
}

func (es *EmailService) WhoFrom(email *Email) string {
	switch {
	case email.From != "":
		return email.From
	case es.DefaultSender != "":
		return es.DefaultSender
	default:
		return DefaultSender
	}
}
func (es *EmailService) Send(email Email) error {
	msg := mail.NewMessage()
	msg.SetHeader("To", email.To)
	msg.SetHeader("From", es.WhoFrom(&email))
	msg.SetHeader("Subject", email.Subject)

	switch {
	case email.Plaintext != "" && email.Html != "":
		msg.SetBody("text/plain", email.Plaintext)
		msg.AddAlternative("text/html", email.Html)
	case email.Plaintext != "":
		msg.SetBody("text/plain", email.Plaintext)
	case email.To != "":
		msg.SetBody("text/html", email.Html)
	}

	err := es.dialer.DialAndSend(msg)
	if err != nil {
		return fmt.Errorf("email send: %v", err)
	}
	return nil
}

func (es *EmailService) ForgetPassword(to, resetUrl string) error {
	email := Email{
		To:        to,
		Subject:   "Reset your password",
		Plaintext: "To reset your password, please visit the following link: " + resetUrl,
		Html:      `<p>To reset your password, please visit the following link: <a href="` + resetUrl + `">Reset Password</a></p>`,
	}
	err := es.Send(email)
	if err != nil {
		return fmt.Errorf("forgot password: %v", err)
	}
	return nil
}
