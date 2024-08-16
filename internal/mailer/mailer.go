package mailer

import (
	"bytes"
	"embed"
	"html/template"
	"time"

	"github.com/wneessen/go-mail"
)

// The follow indicates to Go that we want to store the contents of The
// ./templates directory in the templateFS embedded file system variable.
// ↓↓↓

//go:embed "templates"
var templateFS embed.FS

type Mailer struct {
	dialer *mail.Client
	sender string
}

func New(host string, port int, username, password, sender string) Mailer {
	dialer, _ := mail.NewClient(
		"sandbox.smtp.mailtrap.io",
		mail.WithPort(port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(username),
		mail.WithPassword(password),
	)

	return Mailer{
		dialer: dialer,
		sender: sender,
	}
}

func (m Mailer) Send(recipient, templateFile string, data interface{}) error {
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	msg := mail.NewMsg()
	msg.From(m.sender)
	msg.To(recipient)
	msg.Subject(subject.String())
	msg.SetBodyHTMLTemplate(tmpl, data)
	msg.SetBodyString("text/plain", plainBody.String())
	msg.AddAlternativeString("text/html", htmlBody.String())

	for i := 1; i <= 3; i++ {
		err = m.dialer.DialAndSend(msg)
		// message sent, return nil
		if nil == err {
			return nil
		}

		// failed, sleep and try again
		time.Sleep(500 * time.Millisecond)
	}

	return nil
}
