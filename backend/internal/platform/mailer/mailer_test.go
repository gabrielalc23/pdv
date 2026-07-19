package mailer

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestNewLogMailerEnvironmentValidation(t *testing.T) {
	for _, environment := range []string{"development", "DEVELOPMENT", "test", " test "} {
		if _, err := NewLogMailer(environment); err != nil {
			t.Fatalf("NewLogMailer(%q): %v", environment, err)
		}
	}
	for _, environment := range []string{"", "production", "staging", "dev"} {
		if _, err := NewLogMailer(environment); err == nil {
			t.Fatalf("NewLogMailer(%q) unexpectedly succeeded", environment)
		}
	}
}

func TestLogMailerRedactsMessageData(t *testing.T) {
	var output bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&output, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	mailer, err := NewLogMailer("test")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := mailer.SendEmailVerification(ctx, "recipient-secret@example.test", "Body Secret", "https://example.test/#token=verify-secret"); err != nil {
		t.Fatal(err)
	}
	if err := mailer.SendPasswordReset(ctx, "recipient-secret@example.test", "Body Secret", "https://example.test/#token=reset-secret"); err != nil {
		t.Fatal(err)
	}
	if err := mailer.SendInvitation(ctx, "recipient-secret@example.test", "Organization Secret", "Inviter Secret", "https://example.test/#token=invitation-secret"); err != nil {
		t.Fatal(err)
	}

	logs := output.String()
	for _, secret := range []string{"recipient-secret", "Body Secret", "Organization Secret", "Inviter Secret", "example.test", "verify-secret", "reset-secret", "invitation-secret"} {
		if strings.Contains(logs, secret) {
			t.Fatal("log contains sensitive message data")
		}
	}
	for _, messageType := range []string{"verification", "password_reset", "invitation"} {
		if !strings.Contains(logs, messageType) {
			t.Fatalf("log does not contain safe message type %q", messageType)
		}
	}
}

func TestLogMailerHonorsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := (LogMailer{}).SendEmailVerification(ctx, "to@example.test", "Name", "https://example.test"); err != context.Canceled {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

func TestZeroValueLogMailerIsDisabled(t *testing.T) {
	if err := (LogMailer{}).SendPasswordReset(context.Background(), "to@example.test", "Name", "https://example.test"); err == nil {
		t.Fatal("zero-value log mailer must not be usable")
	}
}

var _ Mailer = (*LogMailer)(nil)
var _ Mailer = (*SMTPMailer)(nil)
