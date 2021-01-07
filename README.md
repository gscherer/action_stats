# Action Stats Library

This library includes functions and methods for creating/modifying and reading "action stats". 

## Interface

### `AddAction(json string) error`

Public method used to add action and time for the action. 
The request is a JSON string in the following format:
```json
{ "action": "someaction", "time": 10 }
```
The "time" field must be an integer > 0.
This method may be called safely concurrently.

### `GetStats() string`

Public method used to return the entire state of the `actionMap` as
a JSON string. It will return an array of object where each object
holds the action name and current average time for that action.
This method may be called safely concurrently.

Example response:

```json
[{"action":"jump","avg":85},{"action":"walk","avg":150}]
```

### `NewActionMap() *actionMap`

Get a new instance of an `actionMap`. It won't hold any values until a request to `AddAction` is made.

### `StartServer(portnumber string)`

Public function used to start the HTTP server. This server hosts two endpoints:
  - GET /action-stats
  - POST /action-stats
When started, it creates a new `actionMap`, which is modified by successful requests to `POST /action-stats`.

## Running Tests

Tests can be run by running `go test -v` in this repo's directory. Output should look similar to this:

```sh
$ go test -v
=== RUN   TestConcurrentActions
--- PASS: TestConcurrentActions (0.00s)
PASS
ok  	github.com/gscherer/action_stats	0.221s
```
