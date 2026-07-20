package mailer

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

type Mailer interface {
	SendEmailVerification(ctx context.Context, to, displayName, link string) error
	SendPasswordReset(ctx context.Context, to, displayName, link string) error
	SendInvitation(ctx context.Context, to, organizationName, inviterName, link string) error
}

// LogMailer is development-only and intentionally never logs recipient data or links.
type LogMailer struct {
	enabled bool
}

func NewLogMailer(environment string) (*LogMailer, error) {
	switch strings.ToLower(strings.TrimSpace(environment)) {
	case "development", "test":
		return &LogMailer{enabled: true}, nil
	default:
		return nil, fmt.Errorf("log mailer is only available in development or test")
	}
}

func (m LogMailer) SendEmailVerification(ctx context.Context, _, _, _ string) error {
	return m.logAccepted(ctx, "verification")
}

func (m LogMailer) SendPasswordReset(ctx context.Context, _, _, _ string) error {
	return m.logAccepted(ctx, "password_reset")
}

func (m LogMailer) SendInvitation(ctx context.Context, _, _, _, _ string) error {
	return m.logAccepted(ctx, "invitation")
}

func (m LogMailer) logAccepted(ctx context.Context, messageType string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !m.enabled {
		return fmt.Errorf("log mailer is not configured for development or test")
	}
	slog.InfoContext(ctx, "message accepted by development mailer", "message_type", messageType)
	return nil
}
