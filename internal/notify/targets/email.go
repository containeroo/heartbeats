package targets

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"

	"github.com/containeroo/heartbeats/internal/notify/render"
	"github.com/containeroo/heartbeats/internal/notify/types"
	"github.com/containeroo/heartbeats/internal/templates"
)

// EmailTarget delivers notifications via SMTP.
type EmailTarget struct {
	Host               string
	Port               int
	User               string
	Pass               string
	From               string
	To                 []string
	StartTLS           bool
	SSL                bool
	InsecureSkipVerify bool
	Template           *templates.StringTemplate
	TitleTmpl          *templates.StringTemplate
	Receiver           string
	Vars               map[string]any
}

// NewEmailTarget constructs an email target.
func NewEmailTarget(cfg EmailTarget) *EmailTarget {
	return &cfg
}

// Type returns the target type name.
func (e *EmailTarget) Type() string { return types.TargetEmail.String() }

// Send renders and sends an email notification.
func (e *EmailTarget) Send(n types.Payload) error {
	if e == nil {
		return errors.New("email target is nil")
	}
	data := render.NewRenderData(n, e.Receiver, e.Vars, "")
	subject, err := e.TitleTmpl.Render(data)
	if err != nil {
		return err
	}
	data.Subject = subject
	body, err := e.Template.Render(data)
	if err != nil {
		return fmt.Errorf("render email template: %w", err)
	}
	return sendEmailSMTP(*e, subject, body)
}

// EmailMessage represents a single email payload.
type EmailMessage struct {
	From    string
	To      []string
	Cc      []string
	Bcc     []string
	Subject string
	Body    string
	IsHTML  bool
}

// sendEmailSMTP sends the email via SMTP with the provided subject/body.
func sendEmailSMTP(cfg EmailTarget, subject, body string) error {
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	msg := buildEmail(EmailMessage{
		From:    cfg.From,
		To:      cfg.To,
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	}, cfg.From)

	if cfg.SSL {
		c, err := smtpClientTLS(addr, cfg)
		if err != nil {
			return err
		}
		defer c.Close() //nolint:errcheck

		if err := smtpAuth(c, cfg); err != nil {
			return err
		}
		return smtpSend(c, cfg.From, cfg.To, msg)
	}

	c, err := smtpClientPlain(addr, cfg)
	if err != nil {
		return err
	}
	defer c.Close() //nolint:errcheck

	if err := smtpAuth(c, cfg); err != nil {
		return err
	}
	return smtpSend(c, cfg.From, cfg.To, msg)
}

func smtpClientTLS(addr string, cfg EmailTarget) (*smtp.Client, error) {
	conn, err := tls.Dial("tcp", addr, smtpTLSConfig(cfg))
	if err != nil {
		return nil, err
	}
	return smtp.NewClient(conn, cfg.Host)
}

func smtpClientPlain(addr string, cfg EmailTarget) (*smtp.Client, error) {
	c, err := smtp.Dial(addr)
	if err != nil {
		return nil, err
	}
	if cfg.StartTLS {
		if err := c.StartTLS(smtpTLSConfig(cfg)); err != nil {
			_ = c.Close()
			return nil, err
		}
	}
	return c, nil
}

func smtpTLSConfig(cfg EmailTarget) *tls.Config {
	return &tls.Config{
		ServerName:         cfg.Host,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
	}
}

func smtpAuth(c *smtp.Client, cfg EmailTarget) error {
	if cfg.User == "" {
		return nil
	}
	return c.Auth(smtp.PlainAuth("", cfg.User, cfg.Pass, cfg.Host))
}

// smtpSend writes the message to the SMTP client after MAIL/RCPT.
func smtpSend(c *smtp.Client, from string, to []string, msg []byte) error {
	if err := c.Mail(from); err != nil {
		return err
	}
	for _, rcpt := range to {
		if err := c.Rcpt(rcpt); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		_ = w.Close()
		return err
	}
	return w.Close()
}

// buildEmail constructs a basic RFC 5322 email message.
func buildEmail(msg EmailMessage, from string) []byte {
	contentType := "text/plain"
	if msg.IsHTML {
		contentType = "text/html"
	}
	headers := map[string]string{
		"From":         from,
		"To":           strings.Join(msg.To, ","),
		"Cc":           strings.Join(msg.Cc, ","),
		"Bcc":          strings.Join(msg.Bcc, ","),
		"Subject":      msg.Subject,
		"MIME-Version": "1.0",
		"Content-Type": fmt.Sprintf("%s; charset=\"UTF-8\"", contentType),
	}

	var buf strings.Builder
	for k, v := range headers {
		if v == "" {
			continue
		}
		fmt.Fprintf(&buf, "%s: %s\r\n", k, v)
	}
	buf.WriteString("\r\n")
	buf.WriteString(msg.Body)
	buf.WriteString("\r\n")
	return []byte(buf.String())
}
