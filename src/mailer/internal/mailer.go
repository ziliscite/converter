package internal

import (
	"bytes"
	"embed"
	"errors"
	"github.com/go-mail/mail/v2"
	"html/template"
	"time"
)

var ErrConnection = errors.New("connection error")

//go:embed "templates"
var templateFS embed.FS

type Mailer struct {
	dialer *mail.Dialer
	sender string
}

func New(host string, port int, username, password, sender string) *Mailer {
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 5 * time.Second

	return &Mailer{
		dialer: dialer,
		sender: sender,
	}
}

// Send takes the recipient email, file name containing the templates, and any
// dynamic data for the templates.
func (m Mailer) Send(recipient, templateFile string, data interface{}) error {
	// parse the required template file from the embedded file.
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	// Execute the template "subject", passing in the dynamic data and storing the
	// result in a bytes.Buffer.
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	// Follow the same pattern to execute the "plainBody".
	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	// And likewise with the "htmlBody" template.
	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	// Initialize a new mail.Message instance.
	// AddAlternative() should always be called *after* SetBody().
	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())

	// Try sending the email up to three times before aborting and returning the final error.
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for i := 1; i <= 3; i++ {
		// Opens a connection to the SMTP server, sends the message, then closes the
		// connection. If there is a timeout, it will return a "dial tcp: i/o timeout"
		// If everything worked, return nil.
		if err = m.dialer.DialAndSend(msg); nil == err {
			return nil
		}

		// If it didn't work, wait for the next tick and retry.
		<-ticker.C
	}

	// If it failed after three attempts, return the error.
	return ErrConnection
}
