# Logvac
[![GoDoc](https://godoc.org/github.com/nanopack/logvac?status.svg)](https://godoc.org/github.com/nanopack/logvac)

Simple, lightweight, api-driven log aggregation service with realtime push capabilities and historical persistence.

## Quickstart

```sh
# start server (may require commented flags)
logvac -s # -A /tmp/auth.db -d /tmp/logvac.db -u 127.0.0.1:6361

# add auth token (using default 'auth-address')
logvac add-token -t TOKEN

# add a log via http
curl -k https://127.0.0.1:6360/logs -H "X-USER-TOKEN: TOKEN" \
     -d '{"id":"log-test", "type":"log", "message":"my first log"}'

# view log via http
curl -k "https://127.0.0.1:6360/logs?type=log&auth=TOKEN"

# Congratulations logmaster!
```

#### Gotchas
- If you're seeing any of the following errors, run logvac with admin or sudo privileges (or adjust your configuration):
  - `Authenticator failed to initialize - open /var/db/log-auth.bolt: permission denied`
  - `Collector failed to initialize - listen udp 127.0.0.1:514: bind: permission denied`

- If logvac doesn't seem to be doing anything (adding/fecthing logs), there is a chance you've started the server with authentication (the default) but have forgotten to add a token:
`logvac add-token -t TOKEN`

- If your logs aren't showing up where you think they should, try checking the 'app' type and see if they are there. By default logvac will log to `type=app` (unless changed via config options). If you have a malformed entry (even with a type specified) it will end up there:
`curl -k "https://127.0.0.1:6360?type=app&auth=TOKEN"`

## Usage
```
  logvac [flags]
  logvac [command]
```

Available Commands:
```
  add-token   Add http publish/subscribe authentication token
  export      Export http publish/subscribe authentication tokens
  import      Import http publish/subscribe authentication tokens
```

Flags:
```
  -A, --auth-address string   Address or file location of authentication db. ('boltdb:///var/db/logvac.bolt' or 'postgresql://127.0.0.1') (default "boltdb:///var/db/log-auth.bolt")
  -c, --config-file string    config file location for server
  -C, --cors-allow string     Sets the 'Access-Control-Allow-Origin' header (default "*")
  -d, --db-address string     Log storage address (default "boltdb:///var/db/logvac.bolt")
  -i, --insecure              Don't use TLS (used for testing)
  -a, --listen-http string    API listen address (same endpoint for http log collection) (default "127.0.0.1:6360")
  -t, --listen-tcp string     TCP log collection endpoint (default "127.0.0.1:6361")
  -u, --listen-udp string     UDP log collection endpoint (default "127.0.0.1:514")
  -k, --log-keep string       Age or number of logs to keep per type '{"app":"2w", "deploy": 10}'' (int or X(m)in, (h)our,  (d)ay, (w)eek, (y)ear) (default "{\"app\":\"2w\"}")
  -l, --log-level string      Level at which to log (default "info")
  -L, --log-type string       Default type to apply to incoming logs (commonly used: app|deploy) (default "app")
  -p, --pub-address string    Log publisher (mist) address ("mist://127.0.0.1:1445")
  -P, --pub-auth string       Log publisher (mist) auth token
  -s, --server                Run as server
  -T, --token string          Administrative token to add/remove 'X-USER-TOKEN's used to pub/sub via http (default "secret")
  -v, --version               Print version info and exit
```

Config File: (takes precedence over cli flags)
```json
// logvac.json
{
  "listen-http": "127.0.0.1:6360",
  "listen-udp": "127.0.0.1:514",
  "listen-tcp": "127.0.0.1:6361",
  "pub-address": "",
  "pub-auth": "",
  "db-address": "boltdb:///var/db/logvac.bolt",
  "auth-address": "boltdb:///var/db/log-auth.bolt",
  "cors-allow": "*",
  "log-keep": "{\"app\":\"2w\"}",
  "log-type": "app",
  "log-level": "info",
  "token": "secret",
  "insecure": false,
  "server": true
}
```

#### As a Server
```
logvac -c logvac.json
## OR (uses defaults seen in config file)
logvac -s
```

#### Cli uses
export|import
```sh
# logvac export dumps the authenticator's database for importing to another authenticator database
logvac export | logvac import -A '/tmp/copy-log-auth.bolt'
## OR
# works with files too
logvac export -f log-auth.dump
```
add-token
```sh
# unless the end user sets auth-address to "", an auth-token will need to be added in order to publish/fetch logs via http
logvac add-token -t "user1-token"
## if you specified a different auth-address for your server, specify it here as such:
logvac add-token -t "user1-token" -A "boltdb:///tmp/log-auth.bolt"
```

#### Adding|Viewing Logs
See syslog examples [here](./collector/README.md)  
See http examples [here](./api/README.md)  
**Important Note:** javascript clients may see up-to a ~100 nanosecond variance when specifying 'start=xxx' as a query parameter due to javascript's lack of precision for the 'number' datatype  

## Todo

- Negate tag (msg.Tag != tag) in drain.Slice if tag[0] == "!"
- Cleanup postgres authenticator db
- Allow multiple filters (id=web.dashboard&id=data.redis)

## Contributing

Contributions to the logvac project are welcome and encouraged. Logvac is a [Nanobox](https://nanobox.io) project and contributions should follow the [Nanobox Contribution Process & Guidelines](https://docs.nanobox.io/contributing/).

## Licence

Mozilla Public License Version 2.0

[![open source](http://nano-assets.gopagoda.io/open-src/nanobox-open-src.png)](http://nanobox.io/open-source)
