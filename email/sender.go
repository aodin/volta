package email

import (
	"fmt"
	"net/smtp"

	"github.com/aodin/config"
)

// Do not use port 465, it expects TLS from the start
// http://stackoverflow.com/a/11664176

// Sender is a common interface for sending emails.
type Sender interface {
	Send(to, subject, body string) error
}

// DefaultSender implements the Sender interface.
type DefaultSender struct {
	c config.SMTP
}

// Send sends the given body to the to address with the given subject.
func (sender DefaultSender) Send(to, subject, body string) error {
	// Create the auth credentials using the given config
	auth := smtp.PlainAuth(
		"",
		sender.c.User,
		sender.c.Password,
		sender.c.Host,
	)

	// Create the email
	// HTML and UTF-8 please
	// TODO Multiple recipients
	mail := Email{
		From:    sender.c.FromAddress(),
		To:      to,
		Subject: subject,
		Header:  map[string]string{"Content-Type": "text/html; charset=UTF-8"},
		Body:    body,
	}

	return smtp.SendMail(
		fmt.Sprintf("%s:%d", sender.c.Host, sender.c.Port),
		auth,
		sender.c.From, // Do not use the alias!
		[]string{to},
		[]byte(mail.String()),
	)
}

// NewSender returns a new DefaultSender that uses the given SMTP config.
func NewSender(c config.SMTP) DefaultSender {
	return DefaultSender{c: c}
}
