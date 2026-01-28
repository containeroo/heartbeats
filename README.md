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

A lightweight HTTP service for monitoring periodic "heartbeat" pings ("bumps") and notifying configured receivers when a heartbeat goes missing or recovers.
It ships with a tiny React dashboard (websocket updates, heartbeat/receiver/history panels) and configurable webhook notifications.

## Run

```bash
go run . \
  --listen-address :8080 \
  --config ./config.yaml
```

Environment variables use the `HEARTBEATS__` prefix, e.g.:

```bash
export HEARTBEATS__CONFIG=./config.yaml
export HEARTBEATS__LISTEN_ADDRESS=:8080
export HEARTBEATS__STRICT_ENV=true
```

## Configuration

Configuration is defined in a YAML file with receivers and heartbeats (Alertmanager-style). See `./deploy/config.yaml` for a full example.

## Migration from deploy/config.yaml

If you previously used the old `deploy/config.yaml`, most of the heartbeat and receiver IDs stay the same. The new schema reorganizes receivers under `receivers.<name>.webhook`/`email` instead of `slack_configs`, and renames `grace` to `late_after`. The older Slack/email templates translate into the new `template`/`subject_override_tmpl` fields (or use the built-in `template: slack` shortcut). Receiver-level `vars` now hold shared values such as channels/tokens, and webhook headers render via the writer-friendly Go templates.

The example under `deploy/` remains a useful reference during migration—translate each Slack/email block into the new structure (move `token`, `channel`, and helper templates into the webhook/email blocks) and keep the heartbeat intervals/receivers as before. Once converted, you can hot-reload the YAML via `SIGHUP` or `POST /-/reload` without restarting the process.

### Env expansion

YAML supports `${VAR}` placeholders which are expanded from the environment before parsing. By default, unresolved placeholders are left intact. Use `--strict-env` (or `HEARTBEATS_STRICT_ENV=true`) to fail on missing or malformed placeholders.

> NOTE: The previous `file:`/`env:file:` helpers have been removed; only direct environment lookups `${VAR}` substitution remain supported now.

### Templates

Template resolution works like this:

- `heartbeats.<id>.subject_tmpl` sets the default subject for that heartbeat.
- `receivers.<name>.webhook.subject_override_tmpl` or `receivers.<name>.email.subject_override_tmpl` overrides the default for that receiver/target.
- The subject is exposed to templates as `.Subject` and can be used inside webhook/email templates.
- `receivers.<name>.vars` is a free-form map exposed as `.Vars` inside templates for custom fields (e.g., Slack channel).
- Template shortcuts are available for built-ins: `template: slack`, `template: default`, and `template: email` (otherwise treated as a file path).

```yaml
receivers:
  ops:
    webhook:
      url: https://example.com/webhook
      headers:
        Authorization: "Bearer YOUR_TOKEN"
      template: templates/default.tmpl
      subject_override_tmpl: "{{ .Title }} is {{ toUpper .Status }}"
    email:
      host: smtp.example.com
      port: 587
      user: smtp-user
      pass: smtp-pass
      from: heartbeat@example.com
      to:
        - ops@example.com
      starttls: true
      ssl: false
      insecure_skip_verify: false
      template: templates/email.tmpl
      subject_override_tmpl: "[{{ .Title }}] {{ .Status }}"
    retry:
      count: 3
      delay: 2s

heartbeats:
  api:
    title: "API"
    interval: 30s
    late_after: 10s
    subject_tmpl: "[{{ .Title }}] {{ .Status }}"
    receivers: ["ops"]
```

## Features

- **Heartbeat monitoring** with configurable `interval`/`late_after` windows, late & missing alerts, and optional recovery notifications.
- **Pluggable receivers** (multiple webhook targets, email) with retry policies, headers, and Go template rendering for both payload and title.
- **Dashboard** served from the built-in SPA with heartbeat, receiver, and history views; WebSocket push keeps the UI in sync without manual refresh.
- **History store**: keeps the last 10,000 events (default) in memory with metrics for bytes used and exposes `/api/history` + `/api/history/{id}`.
- **Metrics**: `/metrics` exposes Prometheus-friendly counters & gauges such as `heartbeats_heartbeat_last_status`, `heartbeats_heartbeat_received_total`, and `heartbeats_receiver_last_status`.
- **Hot reloads**: send `SIGHUP` or `POST /-/reload` to apply a new config without downtime.
- **Debug helpers**: enable `--debug` to hit `/internal/receiver/{id}` or `/internal/heartbeat/{id}` for local testing.

## Endpoints

- `POST /api/heartbeat/{id}` — records a heartbeat bump (accepts any payload/body).
- `GET /api/status` — JSON snapshot of all heartbeat stages.
- `GET /api/history` and `/api/history/{id}` — view the in-memory history for all heartbeats or a specific one.
- `GET /healthz` and `POST /healthz` — liveness probe.
- `/metrics` — Prometheus metrics endpoint.
- `POST /-/reload` — reload configuration (supports both HTTP and SIGHUP).

## Notes

- Any heartbeat payload is accepted; the body is stored as the last payload for heartbeat context.
- Alerts fire when late and again when missing after the late window.
- Recovery alerts are enabled by default (`HEARTBEATS_ALERT_ON_RECOVERY=true`).

## License

This project is licensed under the Apache 2.0 License. See the [LICENSE](LICENSE) file for details.
