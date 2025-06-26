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

| Flag                  | Shorthand | Default                 | Environment Variable           | Description                                                              |
| :-------------------- | :-------- | :---------------------- | :----------------------------- | :----------------------------------------------------------------------- | ------ |
| `--config`            | `-c`      | `heartbeats.yml`        | `HEARTBEATS_CONFIG`            | Path to configuration file                                               |
| `--listen-address`    | `-a`      | `:8080`                 | `HEARTBEATS_LISTEN_ADDRESS`    | Address to listen on (host:port)                                         |
| `--site-root`         | `-r`      | `http://localhost:8080` | `HEARTBEATS_SITE_ROOT`         | Base URL for dashboard and link rendering                                |
| `--skip-tls`          | -         | `false`                 | `HEARTBEATS_SKIP_TLS`          | Skip TLS verification for all receivers (can be overridden per receiver) |
| `--debug`             | `-d`      | `false`                 | `HEARTBEATS_DEBUG`             | Enable debug-level logging                                               |
| `--debug-server-port` | `-p`      | `8081`                  | `HEARTBEATS_DEBUG_SERVER_PORT` | Port for the debug server                                                |
| `--log-format`        | `-l`      | `text`                  | `HEARTBEATS_LOG_FORMAT`        | Log format (`json` or `text`)                                            |
| `--retry-count`       | -         | `3`                     | `HEARTBEATS_RETRY_COUNT`       | Number of times to retry a failed notification. Use `-1` for infinite.   |
| `--retry-delay`       | -         | `2s`                    | `HEARTBEATS_RETRY_DELAY`       | Delay between retries. Must be ≥ 1s.                                     |
| `--history-backend`   | -         | `ring`                  | `HEARTBEATS_HISTORY_BACKEND`   | Backend for history: ring                                                | badger |
| `--ring-size`         | -         | `10000`                 | `HEARTBEATS_RING_SIZE`         | Maximum number of historical heartbeat events to retain                  |
| `--badger-path`       | -         | `db`                    | `HEARTBEATS_BADGR_PATH`        | Path to the badger directory                                             |
| `--help`              | `-h`      | -                       | -                              | Show help and exit                                                       |
| `--version`           | -         | -                       | -                              | Print version and exit                                                   |

#### Proxy Environment Variables

You can set the following environment variables for proxy configuration:

- `HTTP_PROXY`: URL of the proxy server to use for HTTP requests.
- `HTTPS_PROXY`: URL of the proxy server to use for HTTPS requests.

## HTTP Endpoints

## Endpoints

| Path              | Method        | Description                       |
| :---------------- | :------------ | :-------------------------------- |
| `/`               | `GET`         | Dashboard home page               |
| `/bump/{id}`      | `POST`, `GET` | Create a new heartbeat            |
| `/bump/{id}/fail` | `POST`, `GET` | Manually mark heartbeat as failed |
| `/healthz`        | `GET`         | Liveness probe                    |
| `/metrics`        | `GET`         | Prometheus metrics endpoint       |

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
  dev-crew-prod:
    msteams_configs:
      - webhook_url: file:/secrets/teams/webhooks//production
        # no title nor text specified, will use the default
    msteamsgraph_configs:
      - token: env:MSTEAMSGRAPH_TOKEN
        teamID: env:MSTEAMSGRAPH_TEAM_ID
        channelID: env:MSTEAMSGRAPH_CHANNEL_ID
        # no title nor text specified, will use the default
```

### Heartbeats

A **heartbeat** waits for periodic pings (“bumps”). If no bump arrives within `interval + grace`, notifications are sent.

To reduce noise from race conditions (e.g. pings arriving milliseconds after grace timeout), Heartbeats adds a short internal delay before transitioning to `grace` or `missing`. This ensures smoother handling of near-expiry bumps without affecting responsiveness.

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
- **`isRecent`**: `{{ isRecent .LastBump }}` // isRecent returns true if the last bump was less than 2 seconds ago
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
- TextTemplate: `{{ .ID }} is {{ .Status }} (last bump: {{ ago .LastBump }})"`

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

> `Heartbeats` adds a custom `User-Agent: Heartbeats/<version>` header to all outbound HTTP requests.
> The `Content-Type` header is also set to `application/json`.

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
      bodyTemplate: "Last bump: {{ ago .LastBump }}"
```

#### MS Teams (incomming webhook)

_Defaults:_

- TitleTemplate: `"[{{ upper .Status }}] {{ .ID }}"`
- TextTemplate: `"{{ .ID }} is {{ .Status }} (last bump: {{ ago .LastBump }})"`

```yaml
msteams_configs:
  - webhook_url: file:/secrets/teams/webhook//prod
    # optional custom templates:
    titleTemplate: "[{{ upper .Status }}] {{ .ID }}"
    textTemplate: "{{ .ID }} status: {{ .Status }}"
    # optional: override global skip TLS
    skipTLS: true
```

> `Heartbeats` adds a custom `User-Agent: Heartbeats/<version>` header to all outbound HTTP requests.
> The `Content-Type` header is also set to `application/json`.

#### MS Teams (Graph API) (NOT TESTED)

_Defaults:_

- TitleTemplate: `"[{{ upper .Status }}] {{ .ID }}"`
- TextTemplate: `"{{ .ID }} is {{ .Status }} (last bump: {{ ago .LastBump }})"`

```yaml
msteamsgraph_configs:
  - webhook_url: file:/secrets/teams/webhook//graph
    # optional custom templates:
    titleTemplate: "[{{ upper .Status }}] {{ .ID }}"
    textTemplate: "{{ .ID }} status: {{ .Status }}"
    # optional: override global skip TLS
    skipTLS: true
```

> `Heartbeats` adds a custom `User-Agent: Heartbeats/<version>` header to all outbound HTTP requests.
> The `Content-Type` header is also set to `application/json`.

## Deployment

Download the binary and update the example [config.yaml](./deploy/config.yaml) according your needs.
If you prefer to run heartbeats in docker, you find a `docker-compose.yaml` & `config.yaml` [here](./deploy/).
For a kubernetes deployment you find the manifests [here](./deploy/kubernetes).

## Development & Debugging

Heartbeats includes optional internal endpoints for testing receiver notifications and simulating heartbeats. These are only enabled when the `--debug` flag is set.

### Internal Endpoints

| Path                       | Method | Description                               |
| :------------------------- | :----- | :---------------------------------------- |
| `/internal/receiver/{id}`  | `GET`  | Sends a test notification to the receiver |
| `/internal/heartbeat/{id}` | `GET`  | Simulates a bump for the given heartbeat  |

These endpoints listen only on `127.0.0.1`.

#### Kubernetes

Forward the debug port to your local machine:

```bash
kubectl port-forward deploy/heartbeats 8081:8081

curl http://localhost:8081/internal/receiver/{id}
curl http://localhost:8081/internal/heartbeat/{id}
```

#### Docker

Bind the debug port only to your host’s loopback interface:

```bash
docker run \
  -p 127.0.0.1:8081:8081 \
  -v ./config.yaml:/config.yaml \
  containeroo/heartbeats

curl http://localhost:8081/internal/receiver/{id}
curl http://localhost:8081/internal/heartbeat/{id}
```

> ✅ This ensures `/internal/*` endpoints are **only reachable from your local machine**, not from other containers or the network.

> ⚠️ **Warning:** These endpoints are meant for local testing and debugging only. Never expose them in production.

## License

This project is licensed under the Apache 2.0 License. See the [LICENSE](LICENSE) file for details.
