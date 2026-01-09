package multislog

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"time"
)

// emailHandler passes minLevel and above slog messages to the smtp client.
// It satisfies slog.Handler.
type emailHandler struct {
	smtpClient *smtpClient
	Level      slog.Level
}

// newEmailHandler creates a custom slog.Handler that emits emails.
func newEmailHandler(sc *smtpClient, level slog.Level) *emailHandler {
	return &emailHandler{
		Level:      level,
		smtpClient: sc,
	}
}

// Enabled determines if a slog message will be passed to the smtp client.
func (eh *emailHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= eh.Level
}

// Handle the emailing of the slog message.
func (eh *emailHandler) Handle(_ context.Context, r slog.Record) error {
	if r.Level < eh.Level {
		return nil
	}

	var buf bytes.Buffer
	r.Attrs(func(a slog.Attr) bool {
		fmt.Fprintf(&buf, "%s=%v ", a.Key, a.Value)
		return true
	})

	msg := fmt.Sprintf(
		"Level: %s\nTime: %s\nMessage: %s\nAttributes: %s",
		r.Level.String(),
		r.Time.Format(time.RFC3339),
		r.Message,
		buf.String(),
	)
	err := eh.smtpClient.Send("Log Alert", msg, eh.smtpClient.Recipient)
	if err != nil {
		return fmt.Errorf("logger couldn't send emailr: %w", err)
	}
	return nil
}

// WithAttrs satisfies handler interface.
func (eh *emailHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return eh
}

// WithGroup satisfies handler interface.
func (eh *emailHandler) WithGroup(_ string) slog.Handler {
	return eh
}
