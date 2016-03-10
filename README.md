[![logvac logo](http://nano-assets.gopagoda.io/readme-headers/logvac.png)](http://nanobox.io/open-source#logvac)  
[![Build Status](https://travis-ci.org/nanopack/logvac.svg)](https://travis-ci.org/nanopack/logvac)

# Logvac

Simple, lightweight, api-driven log aggregation service with realtime push capabilities and historical persistence.

## Status

Incomplete/Experimental

## Memory Usage

Currently uses around 600k of memory while idling.

## Usage

```json
{
  "listen-http": "127.0.0.1:1234",
  "listen-udp": "127.0.0.1:1234",
  "listen-tcp": "127.0.0.1:1235",
  "pub-address": "",
  "db-address": "boltdb:///tmp/logvac.bolt",
  "auth-address": "",
  "log-keep": "{\"app\":\"2w\"}",
  "log-type": "app",
  "log-level": "info",
  "token": "secret",
  "server": false
}
```

```
logvac -s --pub-address="127.0.0.1:1445" --db-address="/tmp/logvac.boltdb" --token="secret" --auth-address="boltdb:///tmp/auth.bolt"
```

## Todo

- Documentation
- Tests

### Contributing

Contributions to the logvac project are welcome and encouraged. Logvac is a [Nanobox](https://nanobox.io) project and contributions should follow the [Nanobox Contribution Process & Guidelines](https://docs.nanobox.io/contributing/).

### Licence

Mozilla Public License Version 2.0

[![open source](http://nano-assets.gopagoda.io/open-src/nanobox-open-src.png)](http://nanobox.io/open-source)
