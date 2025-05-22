package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"net/smtp"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockClient is a fake SMTP connection that records method calls and captures written data.
type mockClient struct {
	Calls      []string
	MailFrom   string
	Recipients []string
	DataBuf    *bytes.Buffer
}

func (m *mockClient) Mail(from string) error {
	m.Calls = append(m.Calls, "MAIL")
	m.MailFrom = from
	return nil
}

func (m *mockClient) Rcpt(to string) error {
	m.Calls = append(m.Calls, "RCPT")
	m.Recipients = append(m.Recipients, to)
	return nil
}

func (m *mockClient) Data() (io.WriteCloser, error) {
	m.Calls = append(m.Calls, "DATA")
	m.DataBuf = new(bytes.Buffer)
	return nopWriteCloser{m.DataBuf}, nil
}

func (m *mockClient) Close() error {
	m.Calls = append(m.Calls, "CLOSE")
	return nil
}

func (m *mockClient) StartTLS(*tls.Config) error {
	m.Calls = append(m.Calls, "STARTTLS")
	return nil
}

func (m *mockClient) Auth(_ smtp.Auth) error {
	m.Calls = append(m.Calls, "AUTH")
	return nil
}

type nopWriteCloser struct {
	io.Writer
}

func (nopWriteCloser) Close() error { return nil }

type mockDialer struct {
	client *mockClient
}

func (d mockDialer) Dial(_ string) (Client, error) {
	return d.client, nil
}

func TestMailClient_Send(t *testing.T) {
	t.Parallel()

	mock := &mockClient{}
	client := &MailClient{
		SMTPConfig: SMTPConfig{
			Host:     "smtp.test",
			Port:     25,
			From:     "from@example.com",
			Username: "user",
			Password: "pass",
		},
		Dialer: mockDialer{client: mock},
	}

	msg := Message{
		To:      []string{"to@example.com"},
		Subject: "Test Subject",
		Body:    "This is a test.",
		IsHTML:  false,
	}

	err := client.Send(context.Background(), msg)
	assert.NoError(t, err)
	assert.Contains(t, mock.Calls, "MAIL")
	assert.Contains(t, mock.Calls, "RCPT")
	assert.Contains(t, mock.Calls, "DATA")
	assert.Equal(t, "from@example.com", mock.MailFrom)
	assert.Equal(t, []string{"to@example.com"}, mock.Recipients)
	assert.Contains(t, mock.DataBuf.String(), "This is a test.")
}

func TestBuildMIMEMessage(t *testing.T) {
	t.Parallel()

	msg := Message{
		To:      []string{"to@example.com"},
		Cc:      []string{"cc@example.com"},
		Bcc:     []string{"bcc@example.com"},
		Subject: "Hello",
		Body:    "Hello body",
		IsHTML:  true,
		Attachments: []Attachment{
			{Filename: "doc.txt", Data: []byte("testdata")},
		},
	}

	data := buildMIMEMessage(msg, "sender@example.com")
	out := string(data)

	assert.Contains(t, out, "From: sender@example.com")
	assert.Contains(t, out, "To: to@example.com")
	assert.Contains(t, out, "Subject: Hello")
	assert.Contains(t, out, "Content-Type: text/html")
	assert.Contains(t, out, "Hello body")
	assert.Contains(t, out, "doc.txt")
	assert.Contains(t, out, "testdata")
}
