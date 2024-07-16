package Utilities

import (
	"net/smtp"
	"strings"
)

// uses smtp to send emails
type Mailer struct {
	smtpHost       string
	smtpPort       uint16
	senderEmail    string
	senderPassword string
	recipients     []string
}
type Mail struct {
	Cc      []string
	Subject string
	Body    string
}

func NewMailer(smtpHost string, smtpPort uint16, senderEmail string, senderPassword string, recipients []string) *Mailer {
	return &Mailer{
		smtpHost:       smtpHost,
		smtpPort:       smtpPort,
		senderEmail:    senderEmail,
		senderPassword: senderPassword,
	}
}

func NewMail(cc []string, subject string, body string) *Mail {
	return &Mail{
		Cc:      cc,
		Subject: subject,
		Body:    body,
	}
}

func (mailer *Mailer) Send(mail *Mail) error {
	auth := smtp.PlainAuth("", mailer.senderEmail, mailer.senderPassword, mailer.smtpHost)
	err := smtp.SendMail(mailer.smtpHost+":"+IntToString(int(mailer.smtpPort)), auth, mailer.senderEmail, mail.Cc, mailer.constructEmail(mail))
	if err != nil {
		return err
	}
	return nil
}

func (mailer *Mailer) constructEmail(mail *Mail) []byte {
	header := make(map[string]string)
	header["From"] = mailer.senderEmail
	header["To"] = strings.Join(mailer.recipients, ",")
	header["Subject"] = mail.Subject
	header["MIME-version"] = "1.0"
	header["Content-Type"] = "text/html"

	message := ""
	for key, value := range header {
		message += key + ": " + value + "\r\n"
	}
	message += "\r\n" + mail.Body
	return []byte(message)
}
