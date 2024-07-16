package Utilities

import (
	"fmt"
	"net/smtp"
	"strings"
)

type Mailer struct {
	smtpHost       string
	smtpPort       uint16
	senderEmail    string
	senderPassword string
	recipients     []string
	mailQueue      chan *Mail
	logger         *Logger
	stopChan       chan bool
}

type Mail struct {
	Cc      []string
	Subject string
	Body    string
}

func NewMailer(smtpHost string, smtpPort uint16, senderEmail string, senderPassword string, recipients []string, logger *Logger) *Mailer {
	mailer := &Mailer{
		smtpHost:       smtpHost,
		smtpPort:       smtpPort,
		senderEmail:    senderEmail,
		senderPassword: senderPassword,
		recipients:     recipients,
		logger:         logger,
		mailQueue:      make(chan *Mail, 1000),
	}
	go mailer.sendRoutine()
	return mailer
}

func (mailer *Mailer) SetRecipients(recipients []string) {
	mailer.recipients = recipients
}

func (mailer *Mailer) GetRecipients() []string {
	return mailer.recipients
}

func (mailer *Mailer) Close() {
	close(mailer.stopChan)
	close(mailer.mailQueue)
}

func NewMail(cc []string, subject string, body string) *Mail {
	return &Mail{
		Cc:      cc,
		Subject: subject,
		Body:    body,
	}
}

func (mailer *Mailer) Send(mail *Mail) {
	mailer.mailQueue <- mail
}

func (mailer *Mailer) sendRoutine() {
	for {
		select {
		case <-mailer.stopChan:
			return
		case mail := <-mailer.mailQueue:
			err := mailer.sendMail(mail)
			if err != nil {
				mailer.logger.Error("failed sending email: " + err.Error())
			}
		}
	}
}

func (mailer *Mailer) sendMail(mail *Mail) error {
	auth := smtp.PlainAuth("", mailer.senderEmail, mailer.senderPassword, mailer.smtpHost)
	to := append(mailer.recipients, mail.Cc...)
	msg := mailer.constructEmail(mail)
	err := smtp.SendMail(fmt.Sprintf("%s:%d", mailer.smtpHost, mailer.smtpPort), auth, mailer.senderEmail, to, msg)
	return err
}

func (mailer *Mailer) constructEmail(mail *Mail) []byte {
	header := make(map[string]string)
	header["From"] = mailer.senderEmail
	header["To"] = strings.Join(mailer.recipients, ",")
	header["Cc"] = strings.Join(mail.Cc, ",")
	header["Subject"] = mail.Subject
	header["MIME-version"] = "1.0"
	header["Content-Type"] = "text/html; charset=\"UTF-8\""

	message := ""
	for key, value := range header {
		message += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	message += "\r\n" + mail.Body
	return []byte(message)
}
