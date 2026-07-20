package mailer

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestSMTPConfigValidation(t *testing.T) {
	valid := SMTPConfig{Host: "smtp.example.test", Port: 587, From: "PDV <mail@example.test>"}
	tests := []struct {
		name   string
		mutate func(*SMTPConfig)
	}{
		{name: "empty host", mutate: func(config *SMTPConfig) { config.Host = "" }},
		{name: "invalid host", mutate: func(config *SMTPConfig) { config.Host = "bad host" }},
		{name: "zero port", mutate: func(config *SMTPConfig) { config.Port = 0 }},
		{name: "large port", mutate: func(config *SMTPConfig) { config.Port = 65536 }},
		{name: "invalid from", mutate: func(config *SMTPConfig) { config.From = "not-an-email" }},
		{name: "header injection", mutate: func(config *SMTPConfig) { config.From = "mail@example.test\r\nBcc: other@example.test" }},
		{name: "password without username", mutate: func(config *SMTPConfig) { config.Password = "secret" }},
		{name: "negative timeout", mutate: func(config *SMTPConfig) { config.Timeout = -time.Second }},
		{name: "disabled certificate verification", mutate: func(config *SMTPConfig) { config.TLSConfig = &tls.Config{InsecureSkipVerify: true} }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := valid
			test.mutate(&config)
			if _, err := NewSMTPMailer(config); err == nil {
				t.Fatal("expected configuration error")
			}
		})
	}
}

func TestSMTPConfigDefaultsAndClonesTLS(t *testing.T) {
	originalTLS := &tls.Config{}
	mailer, err := NewSMTPMailer(SMTPConfig{Host: "smtp.example.test", Port: 587, From: "mail@example.test", TLSConfig: originalTLS})
	if err != nil {
		t.Fatal(err)
	}
	if mailer.config.Timeout != defaultSMTPTimeout {
		t.Fatalf("timeout = %v, want %v", mailer.config.Timeout, defaultSMTPTimeout)
	}
	if mailer.tlsConfig.ServerName != "smtp.example.test" {
		t.Fatalf("TLS ServerName = %q", mailer.tlsConfig.ServerName)
	}
	if originalTLS.ServerName != "" || mailer.tlsConfig == originalTLS {
		t.Fatal("caller's TLS config was not cloned")
	}
}

func TestSMTPMailerDeliversMultipartMessage(t *testing.T) {
	server := startFakeSMTP(t, false, "")
	mailer, err := NewSMTPMailer(SMTPConfig{Host: server.host, Port: server.port, From: "PDV <mail@example.test>", Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	const link = "https://app.example.test/reset#token=reset-secret"
	if err := mailer.SendPasswordReset(context.Background(), "user@example.test", "User", link); err != nil {
		t.Fatal(err)
	}
	message := server.waitMessage(t)
	if !strings.Contains(message, "multipart/alternative") || !strings.Contains(message, "text/plain") || !strings.Contains(message, "text/html") {
		t.Fatal("message is not multipart/alternative")
	}
	if !strings.Contains(message, "reset-secret") {
		t.Fatal("SMTP message does not contain the reset link")
	}
}

func TestSMTPMailerRequiresAdvertisedSTARTTLS(t *testing.T) {
	server := startFakeSMTP(t, false, "")
	mailer, err := NewSMTPMailer(SMTPConfig{Host: server.host, Port: server.port, From: "mail@example.test", StartTLS: true, Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	err = mailer.SendEmailVerification(context.Background(), "user@example.test", "User", "https://app.example.test/verify#token=secret")
	if err == nil || !strings.Contains(err.Error(), "STARTTLS is required") {
		t.Fatalf("error = %v", err)
	}
}

func TestSMTPErrorDoesNotExposeCredentials(t *testing.T) {
	const username = "smtp-user-secret"
	const password = "smtp-password-secret"
	server := startFakeSMTP(t, true, password)
	mailer, err := NewSMTPMailer(SMTPConfig{Host: "localhost", Port: server.port, From: "mail@example.test", Username: username, Password: password, Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	err = mailer.SendInvitation(context.Background(), "user@example.test", "Org", "Inviter", "https://app.example.test/invitation#token=secret")
	if err == nil {
		t.Fatal("expected authentication failure")
	}
	if strings.Contains(err.Error(), username) || strings.Contains(err.Error(), password) || strings.Contains(err.Error(), base64.StdEncoding.EncodeToString([]byte("\x00"+username+"\x00"+password))) {
		t.Fatalf("SMTP error contains credentials: %v", err)
	}
}

type fakeSMTP struct {
	host     string
	port     int
	messages chan string
}

func startFakeSMTP(t *testing.T, advertiseAuth bool, reflectedSecret string) *fakeSMTP {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	host, portText, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatal(err)
	}
	server := &fakeSMTP{host: host, port: port, messages: make(chan string, 1)}
	go func() {
		connection, acceptErr := listener.Accept()
		if acceptErr != nil {
			return
		}
		defer connection.Close()
		reader := bufio.NewReader(connection)
		writer := bufio.NewWriter(connection)
		writeSMTPResponse(writer, "220 fake smtp ready")
		for {
			line, readErr := reader.ReadString('\n')
			if readErr != nil {
				return
			}
			command := strings.TrimSpace(line)
			switch {
			case strings.HasPrefix(command, "EHLO") || strings.HasPrefix(command, "HELO"):
				if advertiseAuth {
					writeSMTPResponse(writer, "250-fake smtp", "250 AUTH PLAIN")
				} else {
					writeSMTPResponse(writer, "250 fake smtp")
				}
			case strings.HasPrefix(command, "AUTH"):
				writeSMTPResponse(writer, "535 authentication rejected: "+reflectedSecret)
			case strings.HasPrefix(command, "MAIL FROM:") || strings.HasPrefix(command, "RCPT TO:"):
				writeSMTPResponse(writer, "250 accepted")
			case command == "DATA":
				writeSMTPResponse(writer, "354 end with dot")
				var message strings.Builder
				for {
					dataLine, dataErr := reader.ReadString('\n')
					if dataErr != nil {
						return
					}
					if dataLine == ".\r\n" {
						break
					}
					if strings.HasPrefix(dataLine, "..") {
						dataLine = dataLine[1:]
					}
					message.WriteString(dataLine)
				}
				server.messages <- message.String()
				writeSMTPResponse(writer, "250 queued")
			case command == "QUIT":
				writeSMTPResponse(writer, "221 bye")
				return
			default:
				writeSMTPResponse(writer, "500 unknown command")
			}
		}
	}()
	return server
}

func writeSMTPResponse(writer *bufio.Writer, lines ...string) {
	for _, line := range lines {
		_, _ = fmt.Fprintf(writer, "%s\r\n", line)
	}
	_ = writer.Flush()
}

func (server *fakeSMTP) waitMessage(t *testing.T) string {
	t.Helper()
	select {
	case message := <-server.messages:
		return message
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for SMTP message")
		return ""
	}
}
