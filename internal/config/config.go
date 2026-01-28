package config

import "time"

// Config defines the YAML configuration structure.
type Config struct {
	Receivers  map[string]ReceiverConfig  `yaml:"receivers"`  // Receiver definitions.
	Heartbeats map[string]HeartbeatConfig `yaml:"heartbeats"` // Heartbeat definitions.
	History    HistoryConfig              `yaml:"history"`    // History configuration.
}

// ReceiverConfig describes where notifications are delivered.
type ReceiverConfig struct {
	Webhooks []WebhookConfig `yaml:"webhooks,omitempty"` // Webhook delivery settings.
	Emails   []EmailConfig   `yaml:"emails,omitempty"`   // Email delivery settings.
	Retry    RetryConfig     `yaml:"retry,omitempty"`    // Per-receiver retry policy.
	Vars     map[string]any  `yaml:"vars,omitempty"`     // Additional template variables.
}

// WebhookConfig configures webhook delivery.
type WebhookConfig struct {
	URL         string            `yaml:"url"`                             // Destination URL.
	Headers     map[string]string `yaml:"headers,omitempty"`               // Optional headers.
	Template    string            `yaml:"template,omitempty"`              // Webhook payload template path.
	SubjectTmpl string            `yaml:"subject_override_tmpl,omitempty"` // Subject template override.
}

// EmailConfig configures SMTP delivery.
type EmailConfig struct {
	Host               string   `yaml:"host"`                            // SMTP host.
	Port               int      `yaml:"port"`                            // SMTP port.
	User               string   `yaml:"user,omitempty"`                  // SMTP user.
	Pass               string   `yaml:"pass,omitempty"`                  // SMTP password.
	From               string   `yaml:"from"`                            // Sender address.
	To                 []string `yaml:"to"`                              // Recipient list.
	StartTLS           bool     `yaml:"starttls"`                        // Enable STARTTLS.
	SSL                bool     `yaml:"ssl"`                             // Use implicit TLS.
	InsecureSkipVerify bool     `yaml:"insecure_skip_verify"`            // Skip TLS verification.
	Template           string   `yaml:"template,omitempty"`              // Email body template path.
	SubjectTmpl        string   `yaml:"subject_override_tmpl,omitempty"` // Email subject template override.
}

// RetryConfig defines retry behavior for a receiver.
type RetryConfig struct {
	Count int           `yaml:"count"` // Number of retry attempts.
	Delay time.Duration `yaml:"delay"` // Delay between retries.
}

// HistoryConfig defines in-memory history settings.
type HistoryConfig struct {
	Size   int `yaml:"size"`   // Number of events to keep.
	Buffer int `yaml:"buffer"` // Async buffer size for history events.
}

// HeartbeatConfig defines a monitored heartbeat and its receivers.
type HeartbeatConfig struct {
	Title           string        `yaml:"title,omitempty"`             // Human-friendly title.
	Interval        time.Duration `yaml:"interval"`                    // Expected interval between heartbeats.
	LateAfter       time.Duration `yaml:"late_after"`                  // Late window duration.
	AlertOnRecovery *bool         `yaml:"alert_on_recovery,omitempty"` // Enable recovery alerts.
	AlertOnLate     *bool         `yaml:"alert_on_late,omitempty"`     // Enable late alerts.
	SubjectTmpl     string        `yaml:"subject_tmpl,omitempty"`      // Default subject template.
	WebhookTemplate string        `yaml:"webhook_template,omitempty"`  // Default webhook template path.
	EmailTemplate   string        `yaml:"email_template,omitempty"`    // Default email template path.
	Receivers       []string      `yaml:"receivers"`                   // Receiver names for this heartbeat.
}
