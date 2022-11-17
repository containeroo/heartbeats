# Heartbeat

Small helper service to monitor heartbeats.

## Flags

```yaml
  -c, --config string   path to notifications config file (default is ./config.yaml)
  -d, --debug           Verbose logging.
  -h, --help            help for heartbeat
      --host string     Host of Heartbeat service. (default "127.0.0.1")
  -p, --port int        Port to listen on (default 8090)
  -v, --version         Print the current version and exit.
```

## Endpoints

| flag                  | description                              |
| :-------------------- | :--------------------------------------- |
| `/`                   | show current version                     |
| `/healthz`            | show if heartbeats is healthy            |
| `/ping/{HEARTBEAT}`   | reset timer at configured interval       |
| `/status/{HEARTBEAT}` | returns current status of Heartbeat      |
| `/status/`            | returns current status of all Heartbeats |

## Parameters

| query       | description                                           |
| :---------- | :---------------------------------------------------- |
| `output=txt|json| yaml` | return server response in selected format |

## Notifications

Heartbeat uses the library [https://github.com/nikoksr/notify](https://github.com/nikoksr/notify) for notification.

For the moment only `mail`, `slack` and `msteams` are implemented. Feel free to create a Pull Request.

## Config files

Heartbeats and notifications must be configured in a file.
Config files can be `yaml`, `json` or `toml`. The file will be loaded automatically if changed.

To avoid using "secrets" directly in your config file, you can use the prefix `env:` followed by the environment variable.

Examples:

`./config.yaml`

```yaml
---
heartbeats:
  - name: watchdog-prometheus-prd
    description: test prometheus -> alertmanager
    interval: 5m
    grace: 30s
    notifications: # must match with notifications.services[*].name
      - slack
      - mail
  - name: watchdog-prometheus-int
    description: test prometheus -> alertmanager
    interval: 60m
    grace: 5m
    notifications:
      - msteams
notifications:
  defaults:
    subject: Heartbeat {{ .Name }} «{{ .Status }}»
    message: "*Description:*\n{{.Description}}.\n\nLast ping: {{ .GetAgo .LastPing }}"
  services:
    - name: slack
      enabled: false
      type: slack
      oauthToken: env:ENV_VARIABLE_YOU_DEFINE
      channels:
      - test
    - name: mail
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
      subject: "[Heartbeat]: {{ .Name }}"
      message: "Heartbeat is missing.\n\n{{.Description}}\n interval: {{.Interval}}, grace: {{.Grace}}\nPlease check your sending service!"
      webhooks:
        - <YOUR WEBHOOK URL>/teams1
        - <YOUR WEBHOOK URL>/teams2
        - env:WHY_NOT_A_SECRET_WEBHOOK
```

## Notifications

`Defaults` (`notification.defaults`) set the general subject & message for each service.
Each service can override these settings by adding the corresponding key (`subject` and/or `message`)

You can use all properties from `heartbeats` in `subject` and/or `message`. Puth them double curley braces.

There is a function (`GetAgo`) that calculates the time of the last ping to now. (borrowed from [here](https://github.com/xeonx/timeago/))

Example:

```yaml
message: "Last ping was: {{ .GetAgo .LastPing }}"
```
