package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/smtp"
	"strings"

	"github.com/containeroo/heartbeats/pkg/notify/utils"
)

// Sender defines an interface for sending email messages.
type Sender interface {
	// Send sends the given message using the SMTP protocol.
	//
	// Parameters:
	//   - ctx: The context used to control timeout or cancellation.
	//   - msg: The Message to be sent.
	//
	// Returns:
	//   - error: If sending the message fails.
	Send(ctx context.Context, msg Message) error
}

// Dialer establishes SMTP connections for sending messages.
type Dialer interface {
	// Dial connects to the SMTP server at the specified address.
	//
	// Parameters:
	//   - addr: The address of the SMTP server (e.g. "smtp.example.com:587").
	//
	// Returns:
	//   - Client: An established SMTP client.
	//   - error: If the connection cannot be established.
	Dial(addr string) (Client, error)
}

// Client represents a low-level connection to an SMTP server.
type Client interface {
	Mail(from string) error
	Rcpt(to string) error
	Data() (io.WriteCloser, error)
	Close() error
	StartTLS(config *tls.Config) error
	Auth(auth smtp.Auth) error
}

// DefaultDialer implements Dialer using the standard net/smtp package.
type DefaultDialer struct{}

// Dial connects to the SMTP server using the standard library's smtp.Dial.
//
// Parameters:
//   - addr: SMTP server address in host:port format.
//
// Returns:
//   - Client: An active SMTP connection.
//   - error: If the connection fails.
func (DefaultDialer) Dial(addr string) (Client, error) {
	return smtp.Dial(addr)
}

// MailClient sends email messages using SMTP.
type MailClient struct {
	SMTPConfig SMTPConfig // SMTPConfig holds connection and authentication settings.
	Dialer     Dialer     // Dialer provides SMTP connections.
}

// New returns a new MailClient with the given SMTP settings.
//
// Parameters:
//   - cfg: The SMTPConfig specifying connection and authentication details.
//
// Returns:
//   - *MailClient: A configured MailClient ready to send messages.
func New(cfg SMTPConfig) *MailClient {
	return &MailClient{
		SMTPConfig: cfg,
		Dialer:     DefaultDialer{},
	}
}

// Send sends the given email message using the configured SMTP settings.
//
// Parameters:
//   - ctx: A context to manage cancellation and timeout.
//   - msg: The email Message to send.
//
// Returns:
//   - error: If sending the message fails at any stage.
func (c *MailClient) Send(ctx context.Context, msg Message) error {
	raw := buildMIMEMessage(msg, c.SMTPConfig.From)
	addr := fmt.Sprintf("%s:%d", c.SMTPConfig.Host, c.SMTPConfig.Port)

	conn, err := c.Dialer.Dial(addr)
	if err != nil {
		return utils.Wrap(utils.ErrorTransient, "smtp connect", err)
	}
	defer conn.Close() // nolint:errcheck

	if c.SMTPConfig.StartTLS != nil && *c.SMTPConfig.StartTLS {
		tlsConfig := &tls.Config{
			ServerName:         c.SMTPConfig.Host,
			InsecureSkipVerify: c.SMTPConfig.SkipInsecureVerify != nil && *c.SMTPConfig.SkipInsecureVerify,
		}
		if err := conn.StartTLS(tlsConfig); err != nil {
			return utils.Wrap(utils.ErrorTransient, "smtp starttls", err)
		}
	}

	if c.SMTPConfig.Username != "" {
		auth := smtp.PlainAuth("", c.SMTPConfig.Username, c.SMTPConfig.Password, c.SMTPConfig.Host)
		if err := conn.Auth(auth); err != nil {
			return utils.Wrap(utils.ErrorPermanent, "smtp auth", err)
		}
	}

	if err := conn.Mail(c.SMTPConfig.From); err != nil {
		return utils.Wrap(utils.ErrorPermanent, "smtp mail from", err)
	}
	for _, to := range msg.To {
		if err := conn.Rcpt(to); err != nil {
			return utils.Wrap(utils.ErrorPermanent, "smtp rcpt", fmt.Errorf("%s: %w", to, err))
		}
	}

	wc, err := conn.Data()
	if err != nil {
		return utils.Wrap(utils.ErrorTransient, "smtp data", err)
	}
	defer wc.Close() // nolint:errcheck

	if _, err := wc.Write(raw); err != nil {
		return utils.Wrap(utils.ErrorTransient, "smtp write", err)
	}

	return nil
}

// SMTPConfig contains connection and authentication settings for an SMTP server.
type SMTPConfig struct {
	Host               string `yaml:"host,omitempty"`                 // SMTP server hostname.
	Port               int    `yaml:"port,omitempty"`                 // SMTP server port.
	From               string `yaml:"from,omitempty"`                 // Sender email address.
	Username           string `yaml:"username,omitempty"`             // SMTP username.
	Password           string `yaml:"password,omitempty"`             // SMTP password.
	StartTLS           *bool  `yaml:"start_tls,omitempty"`            // Whether to initiate STARTTLS.
	SkipInsecureVerify *bool  `yaml:"skip_insecure_verify,omitempty"` // Whether to skip TLS certificate verification.
}

// Message represents an email to be sent.
type Message struct {
	To          []string     `yaml:"to,omitempty"`          // Recipient email addresses.
	Cc          []string     `yaml:"cc,omitempty"`          // CC recipient email addresses.
	Bcc         []string     `yaml:"bcc,omitempty"`         // BCC recipient email addresses.
	IsHTML      bool         `yaml:"is_html,omitempty"`     // Whether the body is HTML formatted.
	Subject     string       `yaml:"subject,omitempty"`     // Subject of the email.
	Body        string       `yaml:"body,omitempty"`        // Body content of the email.
	Attachments []Attachment `yaml:"attachments,omitempty"` // Optional file attachments.
}

// Attachment represents a single email attachment.
type Attachment struct {
	Filename string // Filename to display for the attachment.
	Data     []byte // Raw file data.
}

// buildMIMEMessage encodes the message into MIME format.
//
// Parameters:
//   - msg: The email message content and metadata.
//   - from: The sender's email address for headers.
//
// Returns:
//   - []byte: MIME-encoded message ready for transmission.
func buildMIMEMessage(msg Message, from string) []byte {
	boundary := "mimeboundary"
	headers := map[string]string{
		"From":         from,
		"To":           strings.Join(msg.To, ","),
		"Cc":           strings.Join(msg.Cc, ","),
		"Bcc":          strings.Join(msg.Bcc, ","),
		"Subject":      msg.Subject,
		"MIME-Version": "1.0",
		"Content-Type": fmt.Sprintf("multipart/mixed; boundary=%s", boundary),
	}

	var buf strings.Builder
	for k, v := range headers {
		fmt.Fprintf(&buf, "%s: %s\r\n", k, v)
	}

	fmt.Fprintf(&buf, "\r\n--%s\r\n", boundary)
	if msg.IsHTML {
		fmt.Fprintf(&buf, "Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n")
	} else {
		fmt.Fprintf(&buf, "Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n")
	}
	buf.WriteString(msg.Body + "\r\n")

	for _, att := range msg.Attachments {
		fmt.Fprintf(&buf, "\r\n--%s\r\n", boundary)
		fmt.Fprintf(&buf, "Content-Type: application/octet-stream; name=\"%s\"\r\n", att.Filename)
		fmt.Fprintf(&buf, "Content-Transfer-Encoding: base64\r\n\r\n")
		fmt.Fprintf(&buf, "%s\r\n", string(att.Data))
	}

	fmt.Fprintf(&buf, "\r\n--%s--\r\n", boundary)
	return []byte(buf.String())
}
