[![logvac logo](http://nano-assets.gopagoda.io/readme-headers/logvac.png)](http://nanobox.io/open-source#logvac)  
[![Build Status](https://travis-ci.org/nanopack/logvac.svg)](https://travis-ci.org/nanopack/logvac)

# Logvac

Simple, lightweight, api-driven log aggregation service with realtime push capabilities and historical persistence.

## Routes:

| Route | Description | Payload | Output |
| --- | --- | --- | --- |
| **Get** /remove-key | Remove a log read/write key | *'X-AUTH-TOKEN' and 'X-ADMIN-TOKEN' headers  | nil |
| **Get** /add-key | Add a log read/write key | *'X-AUTH-TOKEN' and 'X-ADMIN-TOKEN' headers  | nil |
| **Post** / | Post a log | *'X-AUTH-TOKEN' header and json Log object | success message string |
| **Get** / | List all services | *'X-AUTH-TOKEN' header | json array of Log objects |
Note: * = only if 'auth-address' configured

### Query Parameters:
| Parameter | Description |
| --- | --- |
| **id** | Filter by id |
| **tag** | Filter by tag |
| **type** | Filter by type |
| **start** | Start time (unix epoch(nanoseconds)) at which to view logs (defaults to 0) |
| **limit** | Number of logs to read (defaults to 100) |
| **level** | Severity of logs to view (defaults to 'trace') |
`?id=my-app&tag=apache%5Berror%5D&type=deploy&start=0&limit=5`

## Data types:
### Log:
```json
{
  "id": "my-app",
  "tag": "build-1234",
  "type": "deploy",
  "priority": "4",
  "message": "$ mv nanobox/.htaccess .htaccess\n[✓] SUCCESS"
}
```
| Field | Description |
| --- | --- |
| **time** | Timestamp of log (`time.Now()` on post) |
| **id** | Id or hostname of sender |
| **tag** | Tag for log |
| **type** | Log type (commonly 'app' or 'deploy'. default value configured via `log-type`) |
| **priority** | Severity of log (0(trace)-5(fatal)) |
| **message*** | Log data |
Note: * = required on submit


## Usage

add auth key - attempt
```
$ curl -ik https://localhost:1234/add-key -H 'X-AUTH-TOKEN: user'
HTTP/1.1 401 Unauthorized
```

add auth key - success
```
$ curl -ik https://localhost:1234/add-key -H 'X-AUTH-TOKEN: user' -H 'X-ADMIN-TOKEN: secret'
HTTP/1.1 200 OK
```

publish log - attempt
```
$ curl -ik https://localhost:1234 -d '{"id":"my-app","type":"deploy","message":"$ mv nanobox/.htaccess .htaccess\n[✓] SUCCESS"}'
HTTP/1.1 401 Unauthorized
```

publish log - success
```
$ curl -ik https://localhost:1234 -H 'X-AUTH-TOKEN: user' -d '{"id":"my-app","type":"deploy","message":"$ mv nanobox/.htaccess .htaccess\n[✓] SUCCESS"}'
sucess!
HTTP/1.1 200 OK
```

get deploy logs
```
$ curl -k https://localhost:1234?kind=deploy -H 'X-AUTH-TOKEN: user'
[{"time":"2016-03-07T15:48:57.668893791-07:00","id":"my-app","tag":"","type":"deploy","priority":0,"message":"$ mv nanobox/.htaccess .htaccess\n[✓] SUCCESS"}]
```

get app logs
```
$ curl -k https://localhost:1234 -H 'X-AUTH-TOKEN: user'
[]
```

### Contributing

Contributions to the logvac project are welcome and encouraged. Logvac is a [Nanobox](https://nanobox.io) project and contributions should follow the [Nanobox Contribution Process & Guidelines](https://docs.nanobox.io/contributing/).

### Licence

Mozilla Public License Version 2.0

[![open source](http://nano-assets.gopagoda.io/open-src/nanobox-open-src.png)](http://nanobox.io/open-source)
