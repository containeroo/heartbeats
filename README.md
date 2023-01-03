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
| `/`                      | `GET`         | Home Page                                |
| `/config`                | `GET`         | Shows current configuration              |
| `/ping/{HEARTBEAT}`      | `GET`, `POST` | Resets timer at configured interval      |
| `/ping/{HEARTBEAT}/fail` | `GET`, `POST` | Mark heartbeat as failed                 |
| `/status`                | `GET`         | Returns current status of all Heartbeats |
| `/status/{HEARTBEAT}`    | `GET`         | Returns current status of Heartbeat      |
| `/history`               | `GET`         | Show history of all Heartbeats           |
| `/history/{HEARTBEAT}`   | `GET`         | Show history of Heartbeat                |
| `/metrics`               | `GET`         | Entrypoint for prometheus metrics        |
| `/healthz`               | `GET`         | Show if Heartbeats is healthy            |
| `/version`               | `GET`         | Show version of Heartbeats server        |

Add the query `output=json|yaml|text` to receive the response in the corresponding format.

## Configuration

Heartbeats and notifications must be configured in a file.
The config file can be `yaml`, `json` or `toml`. The config file should be loaded automatically if changed. Please check the log output to control if the automatic config reload works in your environment.
If `interval` and `grace` where changed, they will be reset to the corresponding __new value__!

## Deployment

Download the binary and update the example [config.yaml](./deploy/config.yaml) according your needs.
If you prefer to run heartbeats in docker, you find a `docker-compose.yaml` & `config.yaml` [here](./deploy/).
For a kubernetes deployment you find the manifests [here](./deploy/kubernetes).

## Notification

Heartbeat uses the library [https://github.com/containrrr/shoutrrr/](https://github.com/containrrr/shoutrrr/) for notifications.
See their documentation for how to generate the necessary url.

## Documentation

Start heartbeats and go to the url [http://localhost:8090/docs](http://localhost:8090/docs).
