[![logvac logo](http://nano-assets.gopagoda.io/readme-headers/logvac.png)](http://nanobox.io/open-source#logvac)  
[![Build Status](https://travis-ci.org/nanopack/logvac.svg)](https://travis-ci.org/nanopack/logvac)

# Logvac

Simple, lightweight, api-driven log aggregation service with realtime push capabilities and historical persistence.

## Usage (including http collector)

add auth key - attempt
```
$ curl -ik https://localhost:1234/add-key -H 'X-LOGVAC-KEY: user'
HTTP/1.1 401 Unauthorized
```

add auth key - success
```
$ curl -ik https://localhost:1234/add-key -H 'X-LOGVAC-KEY: user' -H 'X-NANOBOX-TOKEN: secret'
HTTP/1.1 200 OK
```

publish log - attempt
```
$ curl -ik https://localhost:1234/ -d 'some log from my system'
HTTP/1.1 401 Unauthorized
```

publish log - success
```
$ curl -ik https://localhost:1234/ -d 'some log from my system' -H 'X-LOGVAC-KEY: user'
HTTP/1.1 200 OK
```

get app logs
```
$ curl -k https://localhost:1234
[]
```

get deploy logs
```
$ curl -k https://localhost:1234?kind=deploy
[]
```

### Contributing

Contributions to the logvac project are welcome and encouraged. Logvac is a [Nanobox](https://nanobox.io) project and contributions should follow the [Nanobox Contribution Process & Guidelines](https://docs.nanobox.io/contributing/).

### Licence

Mozilla Public License Version 2.0

[![open source](http://nano-assets.gopagoda.io/open-src/nanobox-open-src.png)](http://nanobox.io/open-source)

