package mailer

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
)

// SMTPConfig is the connection detail for an SMTP relay. It works with any
// standard provider — Gmail (smtp.gmail.com:587 + App Password), Brevo, SES,
// Postmark, etc. — using STARTTLS + PLAIN auth on the submission port.
type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string // e.g. "SimpliTrade <you@gmail.com>"
}

// Configured reports whether enough is set to actually send mail.
func (c SMTPConfig) Configured() bool {
	return c.Host != "" && c.Port != "" && c.Username != "" && c.Password != "" && c.From != ""
}

// SMTPMailer sends real email through an SMTP relay using the standard library
// (no external dependency). net/smtp.SendMail negotiates STARTTLS automatically
// when the server advertises it, which Gmail and other providers do on :587.
type SMTPMailer struct {
	cfg SMTPConfig
}

func NewSMTPMailer(cfg SMTPConfig) *SMTPMailer { return &SMTPMailer{cfg: cfg} }

func (m *SMTPMailer) SendPasswordResetCode(_ context.Context, to, code string, ttlMinutes int) error {
	subject := "Your SimpliTrade password reset code"
	body := fmt.Sprintf(
		"Your password reset code is:\r\n\r\n    %s\r\n\r\n"+
			"It expires in %d minutes. If you didn't request this, you can ignore this email.\r\n",
		code, ttlMinutes)

	msg := strings.Join([]string{
		"From: " + m.cfg.From,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)
	addr := m.cfg.Host + ":" + m.cfg.Port
	if err := smtp.SendMail(addr, auth, m.cfg.Username, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("smtp: send: %w", err)
	}
	return nil
}
