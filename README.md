# Nanoagent-Logtap

Nanoagent-Mist is a small log storage and publishing application.

## Memory Usage

Currently uses around 600k of memory while idling.

## Routes

| Route | Description | Payload | Output |
| --- | --- | --- | --- |
| GET /subscribe/websocket | establishes a websocket connection with Mist | nil | established websocket |
| GET /ping | simple ping pong route | nil | `pong` |

### Notes