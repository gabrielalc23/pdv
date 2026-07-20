package mailer

import (
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"strings"
	"testing"
)

func TestBuildMessageCreatesMultipartAlternative(t *testing.T) {
	email, err := renderVerification("Usuário", "https://app.example.test/verify#token=secret")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := buildMessage("PDV <mail@example.test>", "user@example.test", email)
	if err != nil {
		t.Fatal(err)
	}
	message, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	mediaType, parameters, err := mime.ParseMediaType(message.Header.Get("Content-Type"))
	if err != nil {
		t.Fatal(err)
	}
	if mediaType != "multipart/alternative" || parameters["boundary"] == "" {
		t.Fatalf("Content-Type = %q", message.Header.Get("Content-Type"))
	}

	parts := multipart.NewReader(message.Body, parameters["boundary"])
	var contentTypes []string
	for {
		part, err := parts.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		contentTypes = append(contentTypes, part.Header.Get("Content-Type"))
		reader := io.Reader(part)
		if part.Header.Get("Content-Transfer-Encoding") == "quoted-printable" {
			reader = quotedprintable.NewReader(part)
		}
		body, err := io.ReadAll(reader)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), "24 horas") {
			t.Fatal("part does not include expiration")
		}
	}
	if len(contentTypes) != 2 || !strings.HasPrefix(contentTypes[0], "text/plain") || !strings.HasPrefix(contentTypes[1], "text/html") {
		t.Fatalf("MIME parts = %v", contentTypes)
	}
}

func TestBuildMessageRejectsHeaderInjection(t *testing.T) {
	_, err := buildMessage("mail@example.test", "victim@example.test\r\nBcc: attacker@example.test", renderedEmail{subject: "subject", text: "text", html: "html"})
	if err == nil {
		t.Fatal("expected invalid header error")
	}
}
