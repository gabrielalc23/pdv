package mailer

import (
	"strings"
	"testing"
)

func TestVerificationAndPasswordResetTemplates(t *testing.T) {
	tests := []struct {
		name       string
		render     func(string, string) (renderedEmail, error)
		expiration string
	}{
		{name: "verification", render: renderVerification, expiration: "24 horas"},
		{name: "password reset", render: renderPasswordReset, expiration: "30 minutos"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			const link = "https://app.example.test/action#token=secret-token"
			email, err := test.render(`<Admin & "Owner">`, link)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(email.text, link) || !strings.Contains(email.html, link) {
				t.Fatal("rendered template does not contain the action link")
			}
			if !strings.Contains(email.text, test.expiration) || !strings.Contains(email.html, test.expiration) {
				t.Fatalf("rendered template does not contain expiration %q", test.expiration)
			}
			if strings.Contains(email.html, `<Admin & "Owner">`) {
				t.Fatal("HTML template did not escape the display name")
			}
			if !strings.Contains(email.html, "&lt;Admin") || !strings.Contains(email.html, "&amp;") {
				t.Fatal("HTML template lacks escaped display name")
			}
		})
	}
}

func TestInvitationTemplateEscapesOrganizationAndInviter(t *testing.T) {
	email, err := renderInvitation("<script>Org</script>", `A & B`, "https://app.example.test/invitation#token=inv_secret")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(email.html, "<script>") || !strings.Contains(email.html, "&lt;script&gt;") || !strings.Contains(email.html, "A &amp; B") {
		t.Fatal("invitation HTML was not escaped")
	}
}

func TestTemplatesRejectMissingLink(t *testing.T) {
	tests := []struct {
		name   string
		render func() error
	}{
		{name: "verification", render: func() error { _, err := renderVerification("Name", "  "); return err }},
		{name: "password reset", render: func() error { _, err := renderPasswordReset("Name", ""); return err }},
		{name: "invitation", render: func() error { _, err := renderInvitation("Org", "Inviter", "\t"); return err }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.render(); err == nil {
				t.Fatal("expected missing link error")
			}
		})
	}
}
