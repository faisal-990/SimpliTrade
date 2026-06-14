// Package mailer is the outbound-email seam. The app depends on the Mailer
// interface, never a concrete transport, so swapping log/SMTP/SES/Postmark is a
// one-line wiring change with no call-site edits.
package mailer

import (
	"context"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/utils"
)

// Mailer sends transactional email.
type Mailer interface {
	// SendPasswordResetCode delivers a one-time passcode to the given address.
	// ttlMinutes is how long the code stays valid (for the message copy).
	SendPasswordResetCode(ctx context.Context, to, code string, ttlMinutes int) error
}

// New returns an SMTP mailer when SMTP is fully configured, otherwise the log
// mailer (so the app always boots and the flow is testable without a provider).
func New(smtp SMTPConfig) Mailer {
	if smtp.Configured() {
		return NewSMTPMailer(smtp)
	}
	utils.Logger().Warn("mailer: SMTP not configured — using log mailer (reset codes go to the server log)")
	return NewLogMailer()
}

// LogMailer "sends" mail by logging it. It lets the full reset flow run with no
// email provider configured: in development you read the code from the server
// log. Used as the fallback when SMTP isn't configured.
type LogMailer struct{}

func NewLogMailer() *LogMailer { return &LogMailer{} }

func (LogMailer) SendPasswordResetCode(_ context.Context, to, code string, ttlMinutes int) error {
	utils.Logger().Info("password reset code (dev log mailer)", "to", to, "code", code, "valid_minutes", ttlMinutes)
	return nil
}
