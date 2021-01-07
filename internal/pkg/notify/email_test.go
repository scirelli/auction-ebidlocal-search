package notify

import (
	"fmt"
	"os"
	"testing"
	"text/template"
)

func TestNewEmail(t *testing.T) {
	var email *Email = NewEmail()
	var expected string = "scirelli+ebidlocal@gmail.com"

	if email.From != expected {
		t.Errorf("expected '%s' got '%s'.", expected, email.From)
	}

	expected = "smtp.gmail.com"
	if email.SmtpHost != expected {
		t.Errorf("expected '%s' got '%s'.", expected, email.SmtpHost)
	}

	expected = "587"
	if email.SmtpPort != expected {
		t.Errorf("expected '%s' got '%s'.", expected, email.SmtpPort)
	}

	expected = ""
	if email.Password != expected {
		t.Errorf("expected '%s' got '%s'.", expected, email.Password)
	}

	expected = "asdfasd"
	os.Setenv("EMAIL_PASSWORD", expected)
	email = NewEmail()
	if email.Password != expected {
		t.Errorf("expected '%s' got '%s'.", expected, email.Password)
	}
	os.Setenv("EMAIL_PASSWORD", "")
}

func TestBuildEmail(t *testing.T) {
	var email *Email = NewEmail()

	const templateText = `<tbody>{{range $index, $element := .}}{{.}}{{end}}</tbody>`
	tmpl, err := template.New("test").Parse(templateText)
	if err != nil {
		t.Errorf("parsing: %s", err)
	}
	email.BodyTmpl = tmpl

	email.TableRows = make(chan string, 1)
	email.TableRows <- "Hi"
	close(email.TableRows)

	var expected = `Watch List ALERT: '' 
MIME-version: 1.0;
Content-Type: text/html; charset="UTF-8";



<tbody>Hi</tbody>`
	var got = fmt.Sprintf("%s", email.BuildEmail())

	if got != expected {
		t.Errorf("expected '%s' got '%s'.", expected, got)
	}
}
