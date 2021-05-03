package email

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEmail(t *testing.T) {
	var email *Email = NewEmail([]string{}, "", "")
	var expected string = FROM_EMAIL

	if email.from != expected {
		t.Errorf("expected '%s' got '%s'.", expected, email.from)
	}

	expected = "smtp.gmail.com"
	if email.smtpHost != expected {
		t.Errorf("expected '%s' got '%s'.", expected, email.smtpHost)
	}

	expected = "587"
	if email.smtpPort != expected {
		t.Errorf("expected '%s' got '%s'.", expected, email.smtpPort)
	}

	expected = os.Getenv("EMAIL_PASSWORD")
	if email.password != expected {
		t.Errorf("expected '%s' got '%s'.", expected, email.password)
	}

	tmp := os.Getenv("EMAIL_PASSWORD")
	expected = "asdfasd"
	os.Setenv("EMAIL_PASSWORD", expected)
	email = NewEmail([]string{}, "", "")
	if email.password != expected {
		t.Errorf("expected '%s' got '%s'.", expected, email.password)
	}
	os.Setenv("EMAIL_PASSWORD", tmp)
}

func TestBuildEmail(t *testing.T) {
	var email *Email = NewEmail([]string{"1"}, "sub", "hi")
	var expected = "To: 1\r\nSubject: sub\r\nMIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\nhi"
	var got = fmt.Sprintf("%s", email.buildEmail())

	assert.Equalf(t, expected, got, "Messages must be same length.")
}

func Skip_TestSend(t *testing.T) {
	var e *Email = NewEmail([]string{"scirelli@gmail.com"}, "Test 1", "<b>This is a test. This is only a test.</b>")
	assert.Nilf(t, e.Send(), "Should send email without error.")
}
