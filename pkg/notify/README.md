# notify

The `notify` package provides pluggable implementations for sending notifications via **Email (SMTP)**, **Slack**, and **Microsoft Teams** (both **Incoming Webhook** and **Graph API**).

## Features

- Send email with plain text, HTML, and attachments
- Post Slack messages with colored attachments
- Send Microsoft Teams messages via:
  - Incoming webhook
  - Graph API (channel messages)
- Configurable headers, TLS settings, and authentication
- Typed error semantics for transient vs permanent failures
- Shared HTTP client defaults and per-request timeouts

## Package Overview

### 🔌 Interfaces

| Interface  | Description                        |
| ---------- | ---------------------------------- |
| `Sender`   | Common interface for all notifiers |
| `Dialer`   | (Email) For abstracting SMTP setup |
| `HTTPDoer` | For mocking HTTP clients           |

### 📦 Subpackages

- [`email`](./email.go) – SMTP email sender with STARTTLS, auth, attachments
- [`slack`](./slack.go) – Sends messages to Slack using `chat.postMessage`
- [`msteams`](./msteams.go) – Sends cards via Microsoft Teams webhook
- [`msteamsgraph`](./msteamspgrahapi.go) – Sends rich messages using Graph API
- [`utils`](./utils.go) – Shared HTTP client abstraction
- [`utils/errors.go`](./utils/errors.go) – Typed error helpers (`transient` vs `permanent`)

## Error Semantics

All HTTP-based notifiers return typed errors via `utils.Wrap`:

- `transient`: safe to retry (timeouts, 5xx, 429, 408, network errors)
- `permanent`: do not retry (4xx, invalid payloads, auth errors)

Use `utils.IsTransient(err)` to decide whether to retry.

## Defaults & Timeouts

HTTP clients share a default per-request timeout of 10s. You can override it via:

- `slack.WithTimeout(...)`
- `msteams.WithTimeout(...)`
- `msteamsgraph.WithTimeout(...)`

## Usage

### Email

```go
cfg := email.SMTPConfig{
    Host:     "smtp.example.com",
    Port:     587,
    From:     "noreply@example.com",
    Username: "user",
    Password: "pass",
}

client := email.New(cfg)
msg := email.Message{
    To:      []string{"recipient@example.com"},
    Subject: "Hello",
    Body:    "This is a test email",
}

_ = client.Send(context.Background(), msg)
```

### Slack

```go
client := slack.NewWithToken("xoxb-your-token")
msg := slack.Slack{
    Channel: "#alerts",
    Attachments: []slack.Attachment{{
        Color: "danger",
        Title: "High Load",
        Text:  "CPU usage exceeds 90%",
    }},
}
_, _ = client.Send(context.Background(), msg)
```

### MS Teams (Webhook)

```go
client := msteams.New()
msg := msteams.MSTeams{
    Title: "Deployment Success",
    Text:  "The app was deployed successfully.",
}
_, _ = client.Send(context.Background(), msg, "https://outlook.office.com/webhook/...")
```

### MS Teams (Graph API)

```go
client := msteamsgraph.NewWithToken("BearerToken")
msg := msteamsgraph.Message{
    Body: msteamsgraph.ItemBody{
        ContentType: "html",
        Content:     "✅ All systems operational.",
    },
}
_, _ = client.SendChannel(context.Background(), "team-id", "channel-id", msg)
```

- All notifiers accept custom `utils.HTTPDoer` for dependency injection and testability.
- Use mocks or replace `HttpClient` to simulate behavior during unit tests.
