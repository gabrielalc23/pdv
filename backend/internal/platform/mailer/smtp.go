package mailer

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

type SMTPConfig struct {
	Host      string
	Port      int
	Username  string
	Password  string
	From      string
	StartTLS  bool
	Timeout   time.Duration
	TLSConfig *tls.Config
}

type SMTPMailer struct {
	config    SMTPConfig
	from      *mail.Address
	tlsConfig *tls.Config
}

const defaultSMTPTimeout = 10 * time.Second

func NewSMTPMailer(config SMTPConfig) (*SMTPMailer, error) {
	config.Host = strings.TrimSpace(config.Host)
	if !validSMTPHost(config.Host) {
		return nil, fmt.Errorf("smtp host is invalid")
	}
	if config.Port < 1 || config.Port > 65535 {
		return nil, fmt.Errorf("smtp port must be between 1 and 65535")
	}
	from, err := parseMailbox(config.From)
	if err != nil {
		return nil, fmt.Errorf("smtp from address is invalid")
	}
	if config.Username == "" && config.Password != "" {
		return nil, fmt.Errorf("smtp username is required when a password is configured")
	}
	if config.Timeout < 0 {
		return nil, fmt.Errorf("smtp timeout must be positive")
	}
	if config.Timeout == 0 {
		config.Timeout = defaultSMTPTimeout
	}

	tlsConfig := config.TLSConfig
	if tlsConfig == nil {
		tlsConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	} else {
		tlsConfig = tlsConfig.Clone()
	}
	if tlsConfig.InsecureSkipVerify {
		return nil, fmt.Errorf("smtp TLS certificate verification cannot be disabled")
	}
	if strings.TrimSpace(tlsConfig.ServerName) == "" {
		tlsConfig.ServerName = config.Host
	}
	config.From = from.String()
	config.TLSConfig = nil
	return &SMTPMailer{config: config, from: from, tlsConfig: tlsConfig}, nil
}

func (m *SMTPMailer) SendEmailVerification(ctx context.Context, to, displayName, link string) error {
	email, err := renderVerification(displayName, link)
	if err != nil {
		return err
	}
	return m.send(ctx, to, email)
}

func (m *SMTPMailer) SendPasswordReset(ctx context.Context, to, displayName, link string) error {
	email, err := renderPasswordReset(displayName, link)
	if err != nil {
		return err
	}
	return m.send(ctx, to, email)
}

func (m *SMTPMailer) SendInvitation(ctx context.Context, to, organizationName, inviterName, link string) error {
	email, err := renderInvitation(organizationName, inviterName, link)
	if err != nil {
		return err
	}
	return m.send(ctx, to, email)
}

func (m *SMTPMailer) send(ctx context.Context, to string, email renderedEmail) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	recipient, err := parseMailbox(to)
	if err != nil {
		return fmt.Errorf("mail recipient is invalid")
	}
	message, err := buildMessage(m.from.String(), recipient.String(), email)
	if err != nil {
		return err
	}

	address := net.JoinHostPort(m.config.Host, strconv.Itoa(m.config.Port))
	dialer := net.Dialer{Timeout: m.config.Timeout}
	connection, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return smtpOperationError(ctx, "connect", err)
	}
	deadline := time.Now().Add(m.config.Timeout)
	if contextDeadline, ok := ctx.Deadline(); ok && contextDeadline.Before(deadline) {
		deadline = contextDeadline
	}
	if err := connection.SetDeadline(deadline); err != nil {
		_ = connection.Close()
		return fmt.Errorf("smtp set deadline failed")
	}
	stopCancellationWatch := make(chan struct{})
	defer close(stopCancellationWatch)
	go func() {
		select {
		case <-ctx.Done():
			_ = connection.SetDeadline(time.Now())
		case <-stopCancellationWatch:
		}
	}()

	client, err := smtp.NewClient(connection, m.config.Host)
	if err != nil {
		_ = connection.Close()
		return smtpOperationError(ctx, "initialize", err)
	}
	closed := false
	defer func() {
		if !closed {
			_ = client.Close()
		}
	}()

	if m.config.StartTLS {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return fmt.Errorf("smtp STARTTLS is required but unavailable")
		}
		if err := client.StartTLS(m.tlsConfig.Clone()); err != nil {
			return smtpOperationError(ctx, "STARTTLS", err)
		}
	}
	if m.config.Username != "" {
		auth := smtp.PlainAuth("", m.config.Username, m.config.Password, m.config.Host)
		if err := client.Auth(auth); err != nil {
			return smtpOperationError(ctx, "authenticate", err)
		}
	}
	if err := client.Mail(m.from.Address); err != nil {
		return smtpOperationError(ctx, "set sender", err)
	}
	if err := client.Rcpt(recipient.Address); err != nil {
		return smtpOperationError(ctx, "set recipient", err)
	}
	data, err := client.Data()
	if err != nil {
		return smtpOperationError(ctx, "begin message", err)
	}
	if _, err := data.Write(message); err != nil {
		_ = data.Close()
		return smtpOperationError(ctx, "write message", err)
	}
	if err := data.Close(); err != nil {
		return smtpOperationError(ctx, "finish message", err)
	}
	if err := client.Quit(); err != nil {
		return smtpOperationError(ctx, "quit", err)
	}
	closed = true
	return nil
}

func parseMailbox(value string) (*mail.Address, error) {
	if strings.TrimSpace(value) == "" || containsHeaderControl(value) {
		return nil, fmt.Errorf("mailbox is empty or contains control characters")
	}
	address, err := mail.ParseAddress(value)
	if err != nil || strings.TrimSpace(address.Address) == "" {
		return nil, fmt.Errorf("mailbox is invalid")
	}
	return address, nil
}

func validSMTPHost(host string) bool {
	if host == "" || containsHeaderControl(host) {
		return false
	}
	if net.ParseIP(host) != nil {
		return true
	}
	name := strings.TrimSuffix(host, ".")
	if len(name) == 0 || len(name) > 253 {
		return false
	}
	for _, label := range strings.Split(name, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, character := range label {
			if (character < 'a' || character > 'z') && (character < 'A' || character > 'Z') && (character < '0' || character > '9') && character != '-' {
				return false
			}
		}
	}
	return true
}

func smtpOperationError(ctx context.Context, operation string, _ error) error {
	if contextErr := ctx.Err(); contextErr != nil {
		return contextErr
	}
	// SMTP responses are controlled by the remote server and may reflect
	// credentials or message data. Keep adapter errors phase-specific only.
	return fmt.Errorf("smtp %s failed", operation)
}
