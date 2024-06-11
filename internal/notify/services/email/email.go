package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
)

// SMTPConfig holds the configuration settings for an SMTP server.
type SMTPConfig struct {
	Host               string `mapstructure:"host,omitempty" yaml:"host,omitempty"`
	Port               int    `mapstructure:"port,omitempty" yaml:"port,omitempty"`
	From               string `mapstructure:"from,omitempty" yaml:"from,omitempty"`
	Username           string `mapstructure:"username,omitempty" yaml:"username,omitempty"`
	Password           string `mapstructure:"password,omitempty" yaml:"password,omitempty"`
	StartTLS           *bool  `mapstructure:"startTLS,omitempty" yaml:"startTLS,omitempty"`
	SkipInsecureVerify *bool  `mapstructure:"skipInsecureVerify,omitempty" yaml:"skipInsecureVerify,omitempty"`
}

// Email represents the structure of an email message.
type Email struct {
	To          []string     `mapstructure:"to,omitempty" yaml:"to,omitempty"`
	Cc          []string     `mapstructure:"cc,omitempty" yaml:"cc,omitempty"`
	Bcc         []string     `mapstructure:"bcc,omitempty" yaml:"bcc,omitempty"`
	IsHTML      bool         `mapstructure:"isHTML,omitempty" yaml:"isHTML,omitempty"`
	Subject     string       `mapstructure:"subject,omitempty" yaml:"subject,omitempty"`
	Body        string       `mapstructure:"body,omitempty" yaml:"body,omitempty"`
	Attachments []Attachment `mapstructure:"attachments,omitempty" yaml:"attachments,omitempty"`
}

// Attachment represents an email attachment.
type Attachment struct {
	Filename string `mapstructure:"filename"`
	Data     []byte `mapstructure:"-"`
}

// MailClient handles the connection and sending of emails.
type MailClient struct {
	SMTPConfig SMTPConfig
}

// New creates a new MailClient with the given SMTP configuration.
//
// Parameters:
//   - smtpConfig: The SMTP configuration settings.
//
// Returns:
//   - *MailClient: A new instance of MailClient.
func New(smtpConfig SMTPConfig) *MailClient {
	return &MailClient{smtpConfig}
}

// Send sends an email using the MailClient's SMTP configuration.
//
// Parameters:
//   - ctx: Context for controlling the lifecycle of the email sending.
//   - email: The email message to be sent.
//
// Returns:
//   - error: An error if sending the email fails.
func (c *MailClient) Send(ctx context.Context, email Email) error {
	msg := createMIMEMessage(email, c.SMTPConfig.From)

	serverAddr := fmt.Sprintf("%s:%d", c.SMTPConfig.Host, c.SMTPConfig.Port)
	conn, err := smtp.Dial(serverAddr)
	if err != nil {
		return fmt.Errorf("error connecting to SMTP server: %v", err)
	}
	defer conn.Close()

	if c.SMTPConfig.StartTLS != nil && *c.SMTPConfig.StartTLS {
		tlsConfig := &tls.Config{
			ServerName:         c.SMTPConfig.Host,
			InsecureSkipVerify: c.SMTPConfig.SkipInsecureVerify != nil && *c.SMTPConfig.SkipInsecureVerify,
		}
		if err := conn.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("error starting TLS. %w", err)
		}
	}

	if c.SMTPConfig.Username != "" {
		auth := smtp.PlainAuth("", c.SMTPConfig.Username, c.SMTPConfig.Password, c.SMTPConfig.Host)
		if err := conn.Auth(auth); err != nil {
			return fmt.Errorf("error authenticating to SMTP server. %w", err)
		}
	}

	if err := conn.Mail(c.SMTPConfig.From); err != nil {
		return fmt.Errorf("error setting sender. %w", err)
	}
	for _, recipient := range email.To {
		if err := conn.Rcpt(recipient); err != nil {
			return fmt.Errorf("error setting recipient (%s). %w", recipient, err)
		}
	}

	wc, err := conn.Data()
	if err != nil {
		return fmt.Errorf("error opening data connection. %w", err)
	}
	defer wc.Close()

	if _, err = wc.Write(msg); err != nil {
		return fmt.Errorf("error writing email body. %w", err)
	}

	return nil
}

// createMIMEMessage constructs a MIME message from the given Email struct and sender address.
//
// Parameters:
//   - email: The email message to be converted into MIME format.
//   - fromAddress: The email address of the sender.
//
// Returns:
//   - []byte: The constructed MIME message as a byte slice.
func createMIMEMessage(email Email, fromAddress string) []byte {
	boundary := "mimeboundary"
	header := make(map[string]string)
	header["From"] = fromAddress
	header["To"] = strings.Join(email.To, ",")
	header["Cc"] = strings.Join(email.Cc, ",")
	header["Bcc"] = strings.Join(email.Bcc, ",")
	header["Subject"] = email.Subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = fmt.Sprintf("multipart/mixed; boundary=%s", boundary)

	var message strings.Builder
	for k, v := range header {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}

	// Email body part
	message.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))
	if email.IsHTML {
		message.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n")
	} else {
		message.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n")
	}
	message.WriteString(email.Body + "\r\n")

	// Attachments
	for _, attachment := range email.Attachments {
		message.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))
		message.WriteString(fmt.Sprintf("Content-Type: application/octet-stream; name=\"%s\"\r\n", attachment.Filename))
		message.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
		message.WriteString(string(attachment.Data) + "\r\n")
	}

	// Final boundary
	message.WriteString(fmt.Sprintf("\r\n--%s--\r\n", boundary))

	return []byte(message.String())
}
