# Heartbeats

![heartbeats.png](.github/icons/heartbeats.png)

Small helper service to monitor heartbeats (repeating "pings" from other systems).
If a "ping" does not arrive in the given interval & grace period, Heartbeats will send notifications.

## Flags

```yaml
-c, --config string           Path to the configuration file (default "./deploy/config.yaml")
-l, --listen-address string   Address to listen on (default "localhost:8080")
-s, --site-root string         Site root for the heartbeat service (default "http://<listenAddress>")
-m, --max-size int             Maximum size of the cache (default 100)
-r, --reduce int              Amount to reduce when max size is exceeded (default 10)
-v, --verbose                 Enable verbose logging
--version                     Show version and exit
-h, --help                    Show help and exit
```

## Environment Variables

You can also set configuration values using environment variables. The command-line flags will take precedence over environment variables. The following environment variables are supported:

- `HEARTBEATS_CONFIG`: Path to the configuration file (corresponds to `--config` or `-c`).
- `HEARTBEATS_LISTEN_ADDRESS`: Address to listen on (corresponds to `--listen-address` or `-l`).
- `HEARTBEATS_SITE_ROOT`: Site root for the heartbeat service (corresponds to `--site-root` or `-s`).
- `HEARTBEATS_MAX_SIZE`: Maximum size of the cache (corresponds to `--max-size` or `-m`).
- `HEARTBEATS_REDUCE`: Amount to reduce when max size is exceeded (corresponds to `--reduce` or `-r`).
- `HEARTBEATS_VERBOSE`: Enable verbose logging (corresponds to `--verbose` or `-v`).

## Endpoints

| Path                   | Method        | Description                         |
| :--------------------- | :------------ | :---------------------------------- |
| `/`                    | `GET`         | Home Page                           |
| `/ping/{HEARTBEAT}`    | `GET`, `POST` | Resets timer at configured interval |
| `/history/{HEARTBEAT}` | `GET`         | Shows the history of a Heartbeat    |
| `/config`              | `GET`         | Returns current config              |
| `/metrics`             | `GET`         | Entrypoint for prometheus metrics   |
| `/healthz`             | `GET`, `POST` | Show if Heartbeats is healthy       |

## Configuration

Heartbeats and notifications must be configured in a file. The config file can be `yaml`, `json` or `toml`.

## Heartbeats

A Heartbeat waits for a ping from another service. If it does not come in the given `interval` and `grace`, it will send a notification to your configured notification services.

| Key           | Description                                                                                                   | Example                                    |
| :------------ | :------------------------------------------------------------------------------------------------------------ | ------------------------------------------ |
| description   | Description of the heartbeat (optional).                                                                      | `test workflow prometheus -> alertmanager` |
| sendResolve   | Sends a `[RESOLVED]` message if a heartbeats changes from `nok` to `ok`. Defaults to `true`                   | `true`                                     |
| interval      | Interval in which a heartbeat must receiver a `ping`. Must be a golang duration.                              | `2m`                                       |
| grace         | Grace period after the interval elapsed. Must be a golang duration.                                           | `1m`                                       |
| notifications | List of notifications to send if the heartbeat grace period elapsed. Must match the key of the notifications. | - `int-slack`                              |

### Example

```yaml
hartbeats:
  prometheus-int:
    description: test workflow prometheus -> alertmanager
    sendResolve: true
    interval: 2s
    grace: 1s
    notifications:
      - int-slack
      - gmail
```

## Notification

Each notification can only have one config. If you want to send multiple notifications, you must create multiple notifications entries.

You can use all properties from heartbeats in `title`, `text`, `subject` or `body`. The variables must start with a dot, a capital letter and be surrounded by double curly brackets. Example: `{{ .Status }}`.
The project includes [masterminds/sprig](https://github.com/Masterminds/sprig) for additional template functions, and also introduces two functions, `isTrue` and `isFalse`. These functions are particularly useful for `SendResolve` and `Enabled`, as these fields default to `true` if not explicitly set.

### Resolving variables

Credentials such as passwords or tokens can be provided in one of the following formats:

- **Plain Text**: Simply input the credentials directly in plain text.
- **Environment Variable**: Use the `env:` prefix, followed by the name of the environment variable that stores the credentials.
- **File**: Use the `file:` prefix, followed by the path of the file that contains the credentials. The file should contain only the credentials.

In case the file contains multiple key-value pairs, the specific key for the credentials can be selected by appending `//KEY` to the end of the path. Each key-value pair in the file must follow the `key = value` format. The system will use the value corresponding to the specified `//KEY`.

All Keys marked with `*` will be resolved as described before.

### slack_config

| Key     | Description                                 | Example                                                                 |
| :------ | :------------------------------------------ | ----------------------------------------------------------------------- |
| channel | Slack channel `*`                           | `monitoring`                                                            |
| token   | Slack token. `*`                            | `file:/secrets/slack_token.txt`                                         |
| title   | Notification title. Can be go-template. `*` | `Heartbeat {{ .Name }} {{ upper .Status }}`                             |
| text    | Notification text. Can be go-template. `*`  | `\*Description:\*\n{{ .Description }}.\nLast ping: {{ ago .LastPing }}` |

`*` will be resolved as described in [Resolve Variables](#Resolve_variables).

### mail_config

#### smpt

| Key                | Description           | Example             |
| :----------------- | :-------------------- | ------------------- |
| host               | smtp host `*`         | `smtp.gmail.com`    |
| port               | smtp port `*`         | `587`               |
| from               | from address `*`      | `example@gmail.com` |
| username           | smtp username`*`      | `env:MAIL_USERNAME` |
| password           | smtp password `*`     | `env:MAIL_PASSWORD` |
| startTLS           | start TLS             | `true`              |
| skipInsecureVerify | ignore Certifications | `true`              |

`*` will be resolved as described in [Resolve Variables](#Resolve_variables).

#### email

| Key     | Description                            | Example                                                                 |
| :------ | :------------------------------------- | ----------------------------------------------------------------------- |
| isHTML  | send emails as HTML                    | `true`                                                                  |
| subject | email subject. Can be go-template. `*` | `Heartbeat {{ .Name }} {{ upper .Status }}`                             |
| body    | email subject. Can be go-template. `*` | `\*Description:\*\n{{ .Description }}./nLast ping: {{ ago .LastPing }}` |
| to      | list of receivers. `*`                 | - `example@gmail.com`                                                   |
| cc      | list of cc. `*`                        | - `cc_user@gmail.com`                                                   |
| bcc     | list of bcc. `*`                       | - `bcc_user@gmail.com`                                                  |

`*` will be resolved as described in [Resolve Variables](#Resolve_variables).

#### MS Teams

| Key        | Description                                 | Example                                                                 |
| :--------- | :------------------------------------------ | ----------------------------------------------------------------------- |
| webhookURL | MS Teams webhook URL. `*`                   | `file:/secrets/teams/webhooks//int-msteasm`                             |
| title      | Notification title. Can be go-template. `*` | `Heartbeat {{ .Name }} {{ upper .Status }}`                             |
| text       | Notification text. Can be go-template. `*`  | `\*Description:\*\n{{ .Description }}.\nLast ping: {{ ago .LastPing }}` |

### Examples

```yaml
---
notifications: # keys must be lowercase!
  dev-slack:
    enabled: true
    slack_config:
      channel: int-monitoring
      token: env:SLACK_TOKEN
      title: Heartbeat {{ .Name }} {{ upper .Status }}
      text: |
        *Description:*
        {{ .Description }}.
        Last ping: {{ if eq (ago .LastPing) "0s" }}now{{ else }}{{ ago .LastPing }}{{ end }}
  int-slack:
    enabled: true
    slack_config:
      channel: int-monitoring
      token: env:SLACK_TOKEN
      title: Heartbeat {{ .Name }} {{ upper .Status }}
      text: |
        *Description:*
        {{ .Description }}.
        Last ping: {{ if eq (ago .LastPing) "0s" }}now{{ else }}{{ ago .LastPing }}{{ end }}
  gmail:
    enabled: false
    mail_config:
      smtp:
        host: smtp.gmail.com
        port: 587
        from: env:MAIL_FROM
        username: env:MAIL_USERNAME
        password: env:MAIL_PASSWORD
        startTLS: true
        skipInsecureVerify: true
      email:
        isHTML: true
        subject: Heartbeat {{ .Name }} {{ upper .Status }}
        body: |
          <b>Description:</b><br>
          {{ .Description }}.<br>
          Last ping: {{ .LastPing }}
        to:
          - monitoring@gmail.com
          - env:EMAIL_TO
  int-teams:
    enabled: false
    msteams_config:
      title: Heartbeat {{ .Name }} {{ upper .Status }}
      text: |
        *Description:*
        {{ .Description }}.
        Last ping: {{ if eq (ago .LastPing) "0s" }}now{{ else }}{{ ago .LastPing }}{{ end }}
      webhook_url: file:/secrets/teams/webhooks//int-teams
```

## Deployment

Download the binary and update the example [config.yaml](./deploy/config.yaml) according your needs.
If you prefer to run heartbeats in docker, you find a `docker-compose.yaml` & `config.yaml` [here](./deploy/).
For a kubernetes deployment you find the manifests [here](./deploy/kubernetes).
