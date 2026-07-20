package mailer

import (
	"bytes"
	"fmt"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/textproto"
	"strings"
	"time"
)

func buildMessage(from, to string, email renderedEmail) ([]byte, error) {
	if containsHeaderControl(from) || containsHeaderControl(to) || containsHeaderControl(email.subject) {
		return nil, fmt.Errorf("mail header contains invalid characters")
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if _, err := fmt.Fprintf(&body,
		"Date: %s\r\nFrom: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/alternative; boundary=%q\r\n\r\n",
		time.Now().UTC().Format(time.RFC1123Z), from, to, mime.QEncoding.Encode("UTF-8", email.subject), writer.Boundary(),
	); err != nil {
		return nil, fmt.Errorf("write mail headers: %w", err)
	}
	if err := writeMIMEPart(writer, "text/plain; charset=UTF-8", email.text); err != nil {
		return nil, err
	}
	if err := writeMIMEPart(writer, "text/html; charset=UTF-8", email.html); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart mail: %w", err)
	}
	return body.Bytes(), nil
}

func writeMIMEPart(writer *multipart.Writer, contentType, content string) error {
	header := make(textproto.MIMEHeader)
	header.Set("Content-Type", contentType)
	header.Set("Content-Transfer-Encoding", "quoted-printable")
	part, err := writer.CreatePart(header)
	if err != nil {
		return fmt.Errorf("create MIME part: %w", err)
	}
	encoded := quotedprintable.NewWriter(part)
	if _, err := encoded.Write([]byte(normalizeCRLF(content))); err != nil {
		return fmt.Errorf("encode MIME part: %w", err)
	}
	if err := encoded.Close(); err != nil {
		return fmt.Errorf("close MIME part: %w", err)
	}
	return nil
}

func containsHeaderControl(value string) bool {
	return strings.ContainsAny(value, "\r\n\x00")
}

func normalizeCRLF(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	return strings.ReplaceAll(value, "\n", "\r\n")
}
