package email

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
)

const FROM_EMAIL = "b6e051d671451ce4c4db+ebidlocal@gmail.com"

type Mailer interface {
	Mail(addr string, a smtp.Auth, from string, to []string, msg []byte) error
}

type MailerFunc func(addr string, a smtp.Auth, from string, to []string, msg []byte) error

func (m MailerFunc) Mail(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	return m(addr, a, from, to, msg)
}

var SendMail MailerFunc = MailerFunc(smtp.SendMail)

type Email struct {
	subject  string
	body     string
	from     string
	to       []string
	password string

	// smtp server configuration.
	smtpHost string
	smtpPort string
}

//NewEmail create a new Email with some gmail defaults
func NewEmail(to []string, subject, body string) *Email {
	return &Email{
		to:       to,
		subject:  subject,
		body:     body,
		from:     FROM_EMAIL,
		password: os.Getenv("EMAIL_PASSWORD"),
		smtpHost: "smtp.gmail.com",
		smtpPort: "587",
	}
}

//buildEmail send an email alerting items of the watch list were found
func (e *Email) buildEmail() *bytes.Buffer {
	var (
		bodyBuffer  bytes.Buffer
		mimeHeaders string = "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";"
	)

	bodyBuffer.Write([]byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n%s\r\n\r\n%s", strings.Join(e.to, ","), e.subject, mimeHeaders, e.body)))

	return &bodyBuffer
}

func (e *Email) Send() error {
	var auth smtp.Auth = smtp.PlainAuth("", e.from, e.password, e.smtpHost)

	if len(e.to) <= 0 {
		return errors.New("To email address is required.")
	}

	err := SendMail.Mail(e.smtpHost+":"+e.smtpPort, auth, e.from, e.to, e.buildEmail().Bytes())
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
