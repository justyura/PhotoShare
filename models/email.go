package models

import (
	"fmt"

	"github.com/go-mail/mail/v2"
)

const (
    DefaultSender = "support@photoshare.com"
)

type Email struct{
    From string
    To string 
    Subject string
    PlainText string
    HTML string
}

type EmailService struct {
    DefaultSender string

    dialer *mail.Dialer
}

type SMTPConfig struct {
    Host string
    Port int
    Username string
    Password string
}

func NewEmailService (config SMTPConfig) *EmailService {
    es := EmailService {
	dialer: mail.NewDialer(config.Host, config.Port, config.Username, config.Password),
    }
    return &es
}

func (es *EmailService) Send(email Email) error {
    msg := mail.NewMessage()
    msg.SetHeader("To", email.To)
    es.setFrom(msg, email)
    msg.SetHeader("Subject", email.Subject)

    switch {
    case email.PlainText != "" && email.HTML != "":
	msg.SetBody("text/plain", email.PlainText)	
	msg.AddAlternative("text/html", email.HTML)
    case email.PlainText != "":
	msg.SetBody("text/html", email.HTML)	
    case email.HTML != "":
	msg.SetBody("text/plain", email.PlainText)	
    }

    err := es.dialer.DialAndSend(msg)
    if err != nil {
	return fmt.Errorf("send err: %w", err)
    }
    return nil
}

func (es *EmailService) setFrom(msg *mail.Message, email Email) {
    var from string
    switch {
    case email.From != "":
	from = email.From
    case es.DefaultSender != "":
	from = es.DefaultSender
    default:
	from = DefaultSender
    }
    msg.SetHeader("From", from)
}
