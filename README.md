# Heartbeats

![heartbeats.png](.github/icons/heartbeats.png)

Small helper service to monitor heartbeats (repeating "pings" from other systems).
If a "ping" does not arrive in the given interval & grace period, Heartbeats will send notifications.

## Flags

```yaml
-c, --config string      Path to notifications config file (default "./config.yaml")
-d, --debug              Verbose logging.
-t, --trace              More verbose logging.
-j, --json               Output logging as json.
-v, --version            Print the current version and exit.
    --host string        Host of Heartbeat service. (default "127.0.0.1")
-p, --port int           Port to listen on (default 8090)
-s, --site-root string   Site root for the heartbeat service (default "http://host:port")
-h, --help               help for heartbeat
```

## Endpoints

| Path                  | Method        | Output             | Description                              |
| :-------------------- | :------------ | :----------------- | :--------------------------------------- |
| `/`                   | `GET`         | `html`             | Show small overview                      |
| `/healthz`            | `GET`         | `json\|yaml\|text` | Show if Heartbeats is healthy            |
| `/ping/{HEARTBEAT}`   | `GET`, `POST` | `json\|yaml\|text` | Resets timer at configured interval      |
| `/status/{HEARTBEAT}` | `GET`         | `json\|yaml\|text` | Returns current status of Heartbeat      |
| `/status`             | `GET`         | `json\|yaml\|text` | Returns current status of all Heartbeats |
| `/metrics`            | `GET`         | `text`             | Entrypoint for prometheus metrics        |
| `/config`             | `GET`         | `json\|yaml\|text` | Shows current configuration              |

To get the response in `json` or `yaml`, add the query `output=json` or `output=yaml|yml` (Default is `output=text|txt`). Not possible for `/` and `/metric` endpoint.

*Example:*

Execute ping:

```sh
curl "http://localhost:8090/status?output=json"
```

Result:

```json
[
  {
    "name": "watchdog-prometheus-prd",
    "status": "",
    "lastPing": "never"
  },
  {
    "name": "watchdog-prometheus-int",
    "status": "OK",
    "lastPing": "30 seconds ago"
  }
]
```

## Config file

Heartbeats and notifications must be configured in a file.
Config files can be `yaml`, `json` or `toml`. The config file should be loaded automatically if changed. Please check the log output to control if the automatic config reload works in your environment.
If `interval` and `grace` where changed, they will be reset to the corresponding *new value*!

Avoid using "secrets" directly in your config file by using environment variables. Set the prefix `env:` followed by the environment variable to load the corresponding environment variable.

Examples:

`./config.yaml`

```yaml
---
heartbeats:
  - name: watchdog-prometheus-prd
    description: test prometheus -> alertmanager workflow
    interval: 5m
    grace: 30s
    notifications: # must match with notifications.services[*].name
      - slack
      - mail_provider_x
  - name: watchdog-prometheus-int
    description: test prometheus -> alertmanager workflow
    interval: 60m
    grace: 5m
    notifications:
      - msTeams
notifications:
  defaults: # uses this subject & message if not overwritten in a service
    subject: Heartbeat {{ .Name }} «{{ .Status }}»
    message: "*Description:*\n{{.Description}}.\n\nLast ping: {{ .TimeAgo .LastPing }}"
  services:
    - name: slack
      enabled: false
      type: slack
      oauthToken: env:ENV_VARIABLE_YOU_DEFINE
      channels:
      - test
    - name: mail_provider_x
      enabled: true
      type: mail
      subject: "[Heartbeat]: {{ .Name }}"
      message: "Heartbeat is missing.\n\n{{.Description}}\n interval: {{.Interval}}, grace: {{.Grace}}\nPlease check your sending service!"
      senderAddress: heartbeat@example.com
      smtpHostAddr: smtp.example.com
      smtpHostPort: 587
      smtpAuthUser: heartbeat@example.com
      smtpAuthPassword: env:ENV_VARIABLE_YOU_DEFINE
      receiverAddresses:
        - heartbeat@example.com
    - name: msTeams
      enabled: true
      type: msteams
      message: "Heartbeat is missing.\n\n{{.Description}}\n interval: {{.Interval}}, grace: {{.Grace}}\nPlease check your sending service!"
      webhooks:
        - http://example.webhook.office.com/webhook2/...
        - env:WHY_NOT_A_SECRET_WEBHOOK
```

### Heartbeat

Each Heartbeat must have following parameters:

| Key             | Description                                                                                    | Wxample                                    |
| :-------------- | :--------------------------------------------------------------------------------------------- | :----------------------------------------- |
| `name`          | Name for heartbeat                                                                             | `watchdog-prometheus-prd`                  |
| `description`   | Description for heartbeat                                                                      | `test prometheus -> alertmanager workflow` |
| `interval`      | Interval in which ping should arrive                                                           | `5m`                                       |
| `grace`         | Grace period which starts after `interval` expired                                             | `30`                                       |
| `notifications` | List of notification to use if grace period is expired. Must match with `Notification[*].name` | - `slack-events` <br> `- gmail`            |

### Notifications

Heartbeat uses the library [https://github.com/nikoksr/notify](https://github.com/nikoksr/notify) for notification.

For the moment only `mail`, `slack` and `msteams` are implemented. Feel free to create a Pull Request.

`Defaults` (`notification.defaults`) set the general subject & message for each service.
Each service can override these settings by adding the corresponding key (`subject` and/or `message`).

You can use all properties from `heartbeats` in `subject` and/or `message`. The variables must start with a dot, a capital letter and be surrounded by double curly brackets. Example: `{{ .Status }}`

There is a function (`TimeAgo`) that calculates the time of the last ping to now. (borrowed from [here](https://github.com/xeonx/timeago/))

Example:

```yaml
message: "Last ping was: {{ .TimeAgo .LastPing }}"
```

#### Slack

| Key          | Description                                                            | Example                                       |
| :----------- | :--------------------------------------------------------------------- | :-------------------------------------------- |
| `name`       | Name for this service                                                  | `slack-events`                                |
| `enabled`    | If enabled, Heartbeat will use this service to send notification       | `true` or `false`                             |
| `type`       | type of notification                                                   | `slack`                                       |
| `subject`    | Subject for Notification. If not set, `defaults.subject` will be used. | `"[Heartbeat]: {{ .Name }}"`                  |
| `message`    | Message for Notification. If not set, `defaults.message` will be used. | `"Heartbeat is missing.\n\n{{.Description}}"` |
| `oauthToken` | Slack oAuth Token (Redacted in endpoint `/config`)                     | xoxb-1234...                                  |
| `Channels`   | List of Channels to send Slack notification                            | `- int ` <br> `- prod`                        |

#### Mail

| Key                 | Description                                                            | Example                                       |
| :------------------ | :--------------------------------------------------------------------- | :-------------------------------------------- |
| `name`              | Name for this service                                                  | `mail_provicer_x`                             |
| `enabled`           | If enabled, Heartbeat will use this service to send notification       | `true` or `false`                             |
| `type`              | type of notification                                                   | `mail`                                        |
| `subject`           | Subject for Notification. If not set, `defaults.subject` will be used. | `"[Heartbeat]: {{ .Name }}"`                  |
| `message`           | Message for Notification. If not set, `defaults.message` will be used. | `"Heartbeat is missing.\n\n{{.Description}}"` |
| `senderAddress`     | SMTP address                                                           | `sender@gmail.com`                            |
| `smtpHostAddr`      | SMTP Host Address                                                      | `smtp.google.com`                             |
| `smtpHostPort`      | SMTP Host Port                                                         | `587`                                         |
| `smtpAuthUser`      | SMTP User (Optional)                                                   | `sender@gmail.com`                            |
| `smtpAuthPassword`  | SMTP Password (Optional) (Redacted in endpoint `/config`)              | `Super Secret!`                               |
| `receiverAddresses` | List of receivers                                                      | `- int@example.com` <br> `- prod@example.com` |

#### MS Teams

| Key       | Description                                                            | Example                                            |
| :-------- | :--------------------------------------------------------------------- | :------------------------------------------------- |
| `name`    | Name for this service                                                  | `msTeams`                                          |
| `enabled` | If enabled, Heartbeat will use this service to send notification       | `true` or `false`                                  |
| `type`    | type of notification                                                   | `slack`                                            |
| `subject` | Subject for Notification. If not set, `defaults.subject` will be used. | `"[Heartbeat]: {{ .Name }}"`                       |
| `message` | Message for Notification. If not set, `defaults.message` will be used. | `"Heartbeat is missing.\n\n{{.Description}}"`      |
| webHooks  | List of Webhooks to send Notification (Redacted in endpoint `/config`) | `- http://example.webhook.office.com/webhook2/...` |

## Metrics

Prometheus metrics can be scraped at the endpoint `/metrics`.
Prometheus metrics starts with `heartbeats_`.

Do not forget to update your Prometheus scrape config.

Example:

```yaml
scrape_configs:
- job_name: heartbeats
  scrape_interval: 30s
  static_configs:
  - targets:
    - heardbeats:8090
```
