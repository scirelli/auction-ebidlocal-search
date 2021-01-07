package notify

import (
	"bytes"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
	"text/template"
)

type Email struct {
	BodyTmpl       *template.Template
	TableRows      chan string
	SearchKeyWords []string

	From     string
	To       []string
	Password string

	// smtp server configuration.
	SmtpHost string
	SmtpPort string
}

//NewEmail create a new Email with some gmail defaults
func NewEmail() *Email {
	return &Email{
		From:     "scirelli+ebidlocal@gmail.com",
		Password: os.Getenv("EMAIL_PASSWORD"),
		SmtpHost: "smtp.gmail.com",
		SmtpPort: "587",
	}
}

//BuildEmail send an email alerting items of the watch list were found
func (e *Email) BuildEmail() *bytes.Buffer {
	var (
		body        bytes.Buffer
		mimeHeaders string = "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
		subject     string = fmt.Sprintf("Watch List ALERT: '%s'", strings.Join(e.SearchKeyWords, " "))
	)
	//t, _ := template.ParseFiles("template.html")

	body.Write([]byte(fmt.Sprintf("%s \n%s\n\n", subject, mimeHeaders)))
	e.BodyTmpl.Execute(&body, e.TableRows)

	return &body
}

func Send(e *Email) error {
	// Authentication.
	var auth smtp.Auth = smtp.PlainAuth("", e.From, e.Password, e.SmtpHost)

	err := smtp.SendMail(e.SmtpHost+":"+e.SmtpPort, auth, e.From, e.To, e.BuildEmail().Bytes())
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("Email Sent!")
	return nil
}
