# Heartbeats

![heartbeats.png](.github/icons/heartbeats.png)

Small helper service to monitor heartbeats (repeating "pings" from other systems).
If a "ping" does not arrive in the given interval & grace period, Heartbeats will send notifications.

## Flags

```yaml
  -c, --config string      Path to Heartbeats config file (default "./config.yaml")
  -d, --debug              Verbose logging.
  -t, --trace              More verbose logging.
  -j, --json-log           Output logging as json.
  -v, --version            Print the current version and exit.
      --host string        Host of Heartbeat service. (default "127.0.0.1")
  -p, --port int           Port to listen on (default 8090)
  -s, --site-root string   Site root for the heartbeat service (default "http://host:port")
  -m, --max-size int       Max Size of History Cache per Heartbeat (default 500)
  -r, --reduce int         Reduce Max Size of History Cache by this value if it exceeds the Max Size (default 100)
  -h, --help               help for heartbeat
```

## Endpoints

| Path                     | Method        | Description                              |
| :----------------------- | :------------ | :--------------------------------------- |
| `/`                      | `GET`         | Show small overview                      |
| `/config`                | `GET`         | Shows current configuration              |
| `/ping/{HEARTBEAT}`      | `GET`, `POST` | Resets timer at configured interval      |
| `/ping/{HEARTBEAT}/fail` | `GET`, `POST` | Mark heartbeat as failed                 |
| `/status`                | `GET`         | Returns current status of all Heartbeats |
| `/status/{HEARTBEAT}`    | `GET`         | Returns current status of Heartbeat      |
| `/metrics`               | `GET`         | Entrypoint for prometheus metrics        |
| `/healthz`               | `GET`         | Show if Heartbeats is healthy            |

## Send a heartbeat

```sh
GET|POST https://heartbeats.example.com/ping/heartbeat1
```

Sends a "alive" message.

### Query parameters

```text
output=json|yaml|yml|text|txt
```

Format response in one of the passed format. If no specific format is passed the response will be `text`.

### Response Codes

| Status        | Description                    |
| :------------ | :----------------------------- |
| 404 Not Found | Given heartbeat not found      |
| 200 OK        | Heartbeat was successully send |

## Send a failed heartbeat

```sh
GET|POST https://heartbeats.example.com/ping/heartbeat1/fail
```

Send a direct failure to not wait until the heartbeat grace period is expired.

### Query parameters

```text
output=json|yaml|yml|text|txt
```

Format response in one of the passed format. If no specific format is passed the response will be `text`.

### Response Codes

| Status        | Description                |
| :------------ | :------------------------- |
| 404 Not Found | Given heartbeat not found  |
| 200 OK        | Fail was successfully send |

## Show heartbeat status

```sh
GET https://heartbeats.example.com/status
```

Shows current status of all heartbeats.

```sh
GET https://heartbeats.example.com/status/heartbeat
```

Shows current status of given heartbeat.
### Query parameters

```text
output=json|yaml|yml|text|txt
```

Format response in one of the passed format. If no specific format is passed the response will be `text`.

### Response Codes

| Status        | Description                |
| :------------ | :------------------------- |
| 404 Not Found | Given heartbeat not found  |
| 200 OK        | Status successful received |

## Show configuration

```sh
GET https://heartbeats.example.com/config
```

Shows current configuration.

### Query parameters

```text
output=json|yaml|yml|text|txt
```

Format response in one of the passed format. If no specific format is passed the response will be `text`.

### Response Codes

| Status                        | Description                            |
| :---------------------------- | :------------------------------------- |
| 500 StatusInternalServerError | Problem with processing current config |
| 200 OK                        | Configuration successful received      |

## Show metrics

```sh
GET https://heartbeats.example.com/metrics
```

Shows metrics for Prometheus.

### Response Codes

| Status | Description                 |
| :----- | :-------------------------- |
| 200 OK | Metrics successful received |

## Show Heartbeats server status

```sh
GET https://heartbeats.example.com/healthz
```

Shows Heartbeats server status.

### Response Codes

| Status | Description                                  |
| :----- | :------------------------------------------- |
| 200 OK | Heartbeats Server status successful received |

## Config file

Heartbeats and notifications must be configured in a file.
Config files can be `yaml`, `json` or `toml`. The config file should be loaded automatically if changed. Please check the log output to control if the automatic config reload works in your environment.
If `interval` and `grace` where changed, they will be reset to the corresponding *new value*!

Avoid using "secrets" directly in your config file by using environment variables. Use regular "bash" variables like `${MY_VAR}` or `$MY_VAR`.

Examples:

`./config.yaml`

```yaml
---
heartbeats:
  - name: watchdog-prometheus-prd
    uuid: 9e22b12b-a9c0-4820-8e54-1b9e226ff45f
    description: test prometheus -> alertmanager workflow
    interval: 5m
    grace: 30s
    notifications: # must match with notifications.services[*].name
      - slack
  - name: watchdog-prometheus-int
    description: test prometheus -> alertmanager workflow
    interval: 60m
    grace: 5m
    notifications:
      - gmail
notifications:
  defaults:
    sendResolved: true
    message: Heartbeat is «{{ .Status }}». Last Ping was «{{ .TimeAgo .LastPing }}»
  services:
  - name: slack
    enabled: true
    shoutrrr: |
      slack://$SLACK_TOKEN@test?color={{ if eq .Status "OK" }}good{{ else }}danger{{ end }}&title=Heartbeat {{ .Name }} «{{ .Status }}»&botname=heartbeats
  - name: gmail
    enabled: true
    shoutrrr: |
      smtp://USERNAME:${MAIL_PASSWORD}@smtp.gmail.com:587?from=example@gmail.com&to=example@gmail.com&subject=Heartbeat {{ .Name }} «{{ .Status }}»
```

### Heartbeat

Each Heartbeat must have following parameters:

| Key             | Description                                                                                    | Example                   |
| :-------------- | :--------------------------------------------------------------------------------------------- | :------------------------ |
| `name`          | Name for heartbeat                                                                             | `watchdog-prometheus-prd` |
| `uuid` | uuid as alternative identifier                                                             | `9e22b12b-a9c0-4820-8e54-1b9e226ff45f` |
| `description`   | Description for heartbeat                                           | `test workflow prometheus -> alertmanager workflow`  |
| `interval`      | Interval in which ping should arrive                                                           |`5m`                       |
| `grace`         | Grace period which starts after `interval` expired                                             | `30`                      |
| `notifications` | List of notification to use if grace period is expired. Must match with `Notifications[*].name` | - `slack` <br> `- gmail`  |

### Notifications

Heartbeat uses the library [https://github.com/containrrr/shoutrrr/](https://github.com/containrrr/shoutrrr/) to send notifications.

`Defaults` (`notification.defaults`) set the general `message` and/or `sendResolved` for each service.
Each service can override these settings by adding the corresponding key (`message` and/or `sendResolved`).

You can use all properties from `heartbeats` in `shoutrrr` and/or `message`. The variables must start with a dot, a capital letter and be surrounded by double curly brackets. Example: `{{ .Status }}`

There is a function (`TimeAgo`) that calculates the time of the last ping to now. (borrowed from [here](https://github.com/xeonx/timeago/))

Example:

```yaml
message: "Last ping was: {{ .TimeAgo .LastPing }}"
```

### Service

| Key            | Description                                                            | Example                                                                                                                                                |
| :------------- | :--------------------------------------------------------------------- | :----------------------------------------------------------------------------------------------------------------------------------------------------- |
| `name`         | Name for this service                                                  | `slack-events`                                                                                                                                         |
| `enabled`      | If enabled, Heartbeat will use this service to send notification       | `true` or `false`                                                                                                                                      |
| `sendResolved` | Send notification if heartbeat changes back to «OK»                    | `true`                                                                                                                                                 |
| `message`      | Message for Notification. If not set, `defaults.message` will be used. | `"Heartbeat is missing.\n\n{{.Description}}"`                                                                                                          |
| `shoutrrr`     | Shoutrrr URL, see [here](https://containrrr.dev/shoutrrr/)             | `slack://$SLACK_TOKEN@prod?color={{ if eq .Status "OK" }}good{{ else }}danger{{ end }}&title=Heartbeat {{ .Name }} «{{ .Status }}»&botname=heartbeats` |

You can use environment variables in `message` and `shoutrrr`, like `$MY_VAR` or `${MY_VAR}`.
Heartbeats will also try to parse `message` and `shoutrrr` as a go-template with the content of the corresponding `heartbeat`.

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
