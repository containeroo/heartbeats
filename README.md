# Heartbeats

![heartbeats.png](./web/static/icons/apple-touch-icon.png)

[![Go Report Card](https://goreportcard.com/badge/github.com/containeroo/heartbeats?style=flat-square)](https://goreportcard.com/report/github.com/containeroo/heartbeats)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/containeroo/heartbeats)
[![Release](https://img.shields.io/github/release/containeroo/heartbeats.svg?style=flat-square)](https://github.com/containeroo/heartbeats/releases/latest)
[![GitHub tag](https://img.shields.io/github/tag/containeroo/heartbeats.svg?style=flat-square)](https://github.com/containeroo/heartbeats/releases/latest)
![Tests](https://github.com/containeroo/heartbeats/actions/workflows/tests.yml/badge.svg)
[![Build](https://github.com/containeroo/heartbeats/actions/workflows/release.yml/badge.svg)](https://github.com/containeroo/heartbeats/actions/workflows/release.yml)
[![license](https://img.shields.io/github/license/containeroo/heartbeats.svg?style=flat-square)](LICENSE)

---

A lightweight HTTP service for monitoring periodic “heartbeat” pings (“bumps”) and notifying configured receivers when a heartbeat goes missing or recovers. Includes an in‐browser read-only dashboard showing current heartbeats, receivers, and historical events.

## Features

- **Heartbeat monitoring** with configurable `interval` & `grace` periods
- **Pluggable notifications** via Slack, Email, or MS Teams
- **In-memory history** of events (received, failed, state changes, notifications, API requests)
- **Dashboard** with:
  - **Heartbeats**: status, URL, last bump, receivers, quick-links
  - **Receivers**: type, destination, last sent, status
  - **History**: timestamped events, filter by heartbeat
  - Text-search filters & copy-to-clipboard URLs
- `/healthz` and `/metrics` endpoints for health checks & Prometheus
- YAML configuration with variable resolution via [containeroo/resolver](https://github.com/containeroo/resolver)

### Flags

| Flag                       | Shorthand | Default                 | Description                                                                 |
| :------------------------- | :-------- | :---------------------- | :-------------------------------------------------------------------------- |
| `--config`, `-c`           | `-c`      | `config.yaml`           | Path to configuration file                                                  |
| `--listen-address`, `-a`   | `-a`      | `:8080`                 | Address to listen on                                                        |
| `--site-root`, `-r`        | `-r`      | `http://localhost:8080` | Base URL for dashboard links                                                |
| `--history-size`, `-s`     | `-s`      | `10000`                 | Maximum history buffer size                                                 |
| `--skip-tls`               | —         | `false`                 | Skip TLS verification for all receivers (can be overridden per receiver)    |
| `--debug`, `-d`            | `-d`      | `false`                 | Enable debug-level logging                                                  |
| `--log-format`, `-l`       | `-l`      | `json`                  | Log format (`json` or `text`)                                               |
| `--retry-count` `int`      | —         | `3`                     | How many times to retry a failed notification. Use -1 for infinite retries. |
| `--retry-delay` `duration` | —         | `5s`                    | Delay between retries. Must be >= 1s.                                       |
| `--help`, `-h`             | `-h`      | —                       | Show help & exit                                                            |
| `--version`                | —         | —                       | Print version & exit                                                        |

#### Proxy Environment Variables

You can set the following environment variables for proxy configuration:

- `HTTP_PROXY`: URL of the proxy server to use for HTTP requests.
- `HTTPS_PROXY`: URL of the proxy server to use for HTTPS requests.

## HTTP Endpoints

## Endpoints

| Path                     | Method        | Description                       |
| :----------------------- | :------------ | :-------------------------------- |
| `/`                      | `GET`         | Dashboard home page               |
| `/api/v1/bump/{id}`      | `POST`, `GET` | Create a new heartbeat            |
| `/api/v1/bump/{id}/fail` | `POST`, `GET` | Manually mark heartbeat as failed |
| `/healthz`               | `GET`         | Liveness probe                    |
| `/metrics`               | `GET`         | Prometheus metrics endpoint       |

## Configuration

`heartbeats` and `receivers` must be defined in your YAML file (default `config.yaml`).

### Examples

```yaml
---
receivers:
  dev-crew-int:
    slack_configs:
      - channel: integration
        token: env:SLACK_TOKEN
        # not title or text specified, will use the default
      - channel: dev-crew
        token: env:SLACK_TOKEN
        # not title or text specified, will use the default
    email_configs:
      - smtp:
          host: smtp.gmail.com
          port: 587
          from: env:MAIL_FROM
          username: env:MAIL_USERNAME
          password: env:MAIL_PASSWORD
          startTLS: true
          skipInsecureVerify: true
        email:
          isHTML: true
          subjectTemplate: "[HEARTBEATS] {{ .Name }} {{ upper .Status }}"
          password: env:MAIL_PASSWORD
  dev-crew-prod:
    msteams_configs:
      webhook_url: file:/secrets/teams/webhooks//production
      # no title nor text specified, will use the default
```

### Heartbeats

A **heartbeat** waits for periodic pings (“bumps”). If no bump arrives within `interval + grace`, notifications are sent.

| Key           | Type       | Description                                                                     |
| :------------ | :--------- | :------------------------------------------------------------------------------ |
| `description` | `string`   | (optional) Human-friendly description                                           |
| `interval`    | `duration` | Required. Go duration (e.g. `30s`, `2m`) for expected interval between pings    |
| `grace`       | `duration` | Required. Go duration after `interval` before marking missing                   |
| `receivers`   | `[]string` | Required. List of receiver IDs (keys under `receivers:`) to notify upon missing |

#### Example

```yaml
heartbeats:
  prometheus-int:
    description: "Prometheus → Alertmanager test"
    interval: 30s
    grace: 10s
    receivers:
      - dev-crew-int
```

### Receivers

Each **receiver** can have multiple notifier configurations. Supported under `receivers:`:

- `slack_configs`
- `email_configs`
- `msteams_configs`

You may use any template variable from the heartbeat (e.g. `{{ .ID }}`, `{{ .Status }}`), and these helper functions:

- **`upper`**: `{{ upper .ID }}`
- **`lower`**: `{{ lower .ID }}`
- **`formatTime`**: `{{ formatTime .LastBump "2006-01-02 15:04:05" }}`
- **`ago`**: `{{ ago .LastBump }}`
- **`join`**: `{{ join .Tags ", " }}`

#### Variable Resolution

`Heartbeats` uses [containeroo/resolver](https://github.com/containeroo/resolver) for variable resolving.

Resolver supports:

- **Plain**: literal value
- **Environment**: `env:VAR_NAME`
- **File**: `file:/path/to/file`
- **Within-file**: `file:/path/to/file//KEY`, also supported `yaml:`,`json:`,`ini:` and `toml:`. For more details see [containeroo/resolver](https://github.com/containeroo/resolver).

#### Slack

_Defaults:_

- SubjectTemplate: `[{{ upper .Status }}] {{ .ID }}"`
- TextTemplate: `{{ .ID }} is {{ .Status }} (last Ping: {{ ago .LastBump }})"`

```yaml
receivers:
  dev-crew-int:
    slack_configs:
      - channel: "#integration"
        token: env:SLACK_TOKEN
        # optional custom templates:
        titleTemplate: "[{{ upper .Status }}] {{ .ID }}"
        textTemplate: "{{ .ID }} status: {{ .Status }}"
        # optional: override global skip TLS
        skipTLS: true
```

#### Email

_Defaults:_

- SubjectTemplate: `"[HEARTBEATS]: {{ .ID }} {{ upper .Status }}"`
- BodyTemplate: `"<b>Description:</b> {{ .Description }}<br>Last bump: {{ ago .LastBump }}"`

```yaml
email_configs:
  - smtp:
      host: smtp.gmail.com
      port: 587
      from: admin@example.com
      username: env:EMAIL_USER
      password: env:EMAIL_PASS
      # optional
      startTLS: true
      # optional: override global skip TLS
      skipInsecureVerify: true
    email:
      isHTML: true
      to: ["ops@example.com"]
      # optional custom templates:
      subjectTemplate: "[HB] {{ .ID }} {{ upper .Status }}"
      bodyTemplate: "Last ping: {{ ago .LastBump }}"
```

#### MS Teams

_Defaults:_

- TitleTemplate: `"[{{ upper .Status }}] {{ .ID }}"`
- TextTemplate: `"{{ .ID }} is {{ .Status }} (last bump: {{ ago .LastBump }})"`

```yaml
msteams_configs:
  - webhookURL: file:/secrets/teams/webhook//prod
    # optional custom templates:
    titleTemplate: "[{{ upper .Status }}] {{ .ID }}"
    textTemplate: "{{ .ID }} status: {{ .Status }}"
    # optional: override global skip TLS
    skipTLS: true
```

## Deployment

Download the binary and update the example [config.yaml](./deploy/config.yaml) according your needs.
If you prefer to run heartbeats in docker, you find a `docker-compose.yaml` & `config.yaml` [here](./deploy/).
For a kubernetes deployment you find the manifests [here](./deploy/kubernetes).

## License

This project is licensed under the Apache 2.0 License. See the [LICENSE](LICENSE) file for details.
