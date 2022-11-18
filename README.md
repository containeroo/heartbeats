# Heartbeat

![heartbeats.pn](.github/icons/heartbeats.png)

Small helper service to monitor heartbeats.

## Flags

```yaml
-c, --config string      Path to notifications config file (default "./config.yaml")
-d, --debug              Verbose logging.
-t, --trace              More verbose logging.
-h, --help               Help for heartbeat
    --host string        Host of Heartbeat service. (default "127.0.0.1")
-p, --port int           Port to listen on (default 8090)
-s, --site-root string   Site root for the heartbeat service (default "http://host:port")
-v, --version            Print the current version and exit.
```

## Endpoints

| Path                  | Method          | Description                              |
| :-------------------- | :-------------- | :--------------------------------------- |
| `/`                   | `GET`           | show current version                     |
| `/healthz`            | `GET`           | show if heartbeats is healthy            |
| `/ping/{HEARTBEAT}`   | `GET`, `POST`   | reset timer at configured interval       |
| `/status/{HEARTBEAT}` | `GET`           | returns current status of Heartbeat      |
| `/status/`            | `GET`           | returns current status of all Heartbeats |

## Parameters

| Query                   | Description                               |
| :---------------------- | :----------------------------------------- |
| `output=txt|json|yaml`  | return server response in selected format |

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

## Config files

Heartbeats and notifications must be configured in a file.
Config files can be `yaml`, `json` or `toml`. The config file will be loaded automatically if changed.
If `interval` and `grace` where changed, they will be reset to the corresponding *new value*!

To avoid using "secrets" directly in your config file, you can use the prefix `env:` followed by the environment variable.

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
  defaults:
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

## Notifications

Heartbeat uses the library [https://github.com/nikoksr/notify](https://github.com/nikoksr/notify) for notification.

For the moment only `mail`, `slack` and `msteams` are implemented. Feel free to create a Pull Request.

`Defaults` (`notification.defaults`) set the general subject & message for each service.
Each service can override these settings by adding the corresponding key (`subject` and/or `message`)

You can use all properties from `heartbeats` in `subject` and/or `message`. They must start with a capital letter and be surrounded by double curly brackets.

There is a function (`TimeAgo`) that calculates the time of the last ping to now. (borrowed from [here](https://github.com/xeonx/timeago/))

Example:

```yaml
message: "Last ping was: {{ .TimeAgo .LastPing }}"
```
