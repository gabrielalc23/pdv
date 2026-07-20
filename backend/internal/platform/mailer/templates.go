package mailer

import (
	"bytes"
	"fmt"
	htmltemplate "html/template"
	"strings"
	texttemplate "text/template"
)

type renderedEmail struct {
	subject string
	text    string
	html    string
}

type templatePair struct {
	subject string
	text    *texttemplate.Template
	html    *htmltemplate.Template
}

type templateData struct {
	DisplayName      string
	OrganizationName string
	InviterName      string
	Link             string
}

var verificationTemplates = templatePair{
	subject: "Verifique seu e-mail",
	text: texttemplate.Must(texttemplate.New("verification-text").Parse(`{{if .DisplayName}}Olá, {{.DisplayName}}.{{else}}Olá.{{end}}

Confirme seu e-mail acessando o link abaixo:
{{.Link}}

Este link expira em 24 horas.
Se você não criou esta conta, ignore esta mensagem.
`)),
	html: htmltemplate.Must(htmltemplate.New("verification-html").Parse(`<!doctype html>
<html lang="pt-BR">
<body>
  <p>{{if .DisplayName}}Olá, {{.DisplayName}}.{{else}}Olá.{{end}}</p>
  <p>Confirme seu e-mail usando o link abaixo:</p>
  <p><a href="{{.Link}}">Verificar e-mail</a></p>
  <p>Este link expira em 24 horas.</p>
  <p>Se você não criou esta conta, ignore esta mensagem.</p>
</body>
</html>
`)),
}

var passwordResetTemplates = templatePair{
	subject: "Redefina sua senha",
	text: texttemplate.Must(texttemplate.New("password-reset-text").Parse(`{{if .DisplayName}}Olá, {{.DisplayName}}.{{else}}Olá.{{end}}

Recebemos uma solicitação para redefinir sua senha. Acesse:
{{.Link}}

Este link expira em 30 minutos.
Se você não solicitou a redefinição, ignore esta mensagem.
`)),
	html: htmltemplate.Must(htmltemplate.New("password-reset-html").Parse(`<!doctype html>
<html lang="pt-BR">
<body>
  <p>{{if .DisplayName}}Olá, {{.DisplayName}}.{{else}}Olá.{{end}}</p>
  <p>Recebemos uma solicitação para redefinir sua senha.</p>
  <p><a href="{{.Link}}">Redefinir senha</a></p>
  <p>Este link expira em 30 minutos.</p>
  <p>Se você não solicitou a redefinição, ignore esta mensagem.</p>
</body>
</html>
`)),
}

var invitationTemplates = templatePair{
	subject: "Convite para sua organização",
	text: texttemplate.Must(texttemplate.New("invitation-text").Parse(`Você recebeu um convite{{if .InviterName}} de {{.InviterName}}{{end}} para participar de {{.OrganizationName}}.

Aceite o convite acessando:
{{.Link}}

Se você não esperava este convite, ignore esta mensagem.
`)),
	html: htmltemplate.Must(htmltemplate.New("invitation-html").Parse(`<!doctype html>
<html lang="pt-BR">
<body>
  <p>Você recebeu um convite{{if .InviterName}} de {{.InviterName}}{{end}} para participar de {{.OrganizationName}}.</p>
  <p><a href="{{.Link}}">Aceitar convite</a></p>
  <p>Se você não esperava este convite, ignore esta mensagem.</p>
</body>
</html>
`)),
}

func renderVerification(displayName, link string) (renderedEmail, error) {
	return renderTemplates(verificationTemplates, templateData{DisplayName: displayName, Link: link})
}

func renderPasswordReset(displayName, link string) (renderedEmail, error) {
	return renderTemplates(passwordResetTemplates, templateData{DisplayName: displayName, Link: link})
}

func renderInvitation(organizationName, inviterName, link string) (renderedEmail, error) {
	return renderTemplates(invitationTemplates, templateData{OrganizationName: organizationName, InviterName: inviterName, Link: link})
}

func renderTemplates(templates templatePair, data templateData) (renderedEmail, error) {
	if strings.TrimSpace(data.Link) == "" {
		return renderedEmail{}, fmt.Errorf("mail link is required")
	}

	var textBody bytes.Buffer
	if err := templates.text.Execute(&textBody, data); err != nil {
		return renderedEmail{}, fmt.Errorf("render text mail template: %w", err)
	}
	var htmlBody bytes.Buffer
	if err := templates.html.Execute(&htmlBody, data); err != nil {
		return renderedEmail{}, fmt.Errorf("render HTML mail template: %w", err)
	}
	return renderedEmail{subject: templates.subject, text: textBody.String(), html: htmlBody.String()}, nil
}
