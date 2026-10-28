## MODIFIED Requirements

### Requirement: Listening Address

The system SHALL start an HTTP server on `127.0.0.1` at the port given by
the `PORT` environment variable, defaulting to `8080`, when the user runs
`scorecard serve`. A `--listen` value with a port SHALL set the host and
port; a value without one SHALL set only the host. The server SHALL log
`Server listening on <host>:<port>`, SHALL time out reading request
headers after 10 seconds, and SHALL log at `info` level. If the server
cannot listen, the command MUST exit with a non-zero status.

#### Scenario: Default port

- **GIVEN** `PORT` is not set
- **WHEN** a user runs `scorecard serve`
- **THEN** the server listens on `127.0.0.1:8080`

#### Scenario: Custom port

- **GIVEN** `PORT=9000`
- **WHEN** a user runs `scorecard serve`
- **THEN** the server listens on `127.0.0.1:9000`

#### Scenario: Listen host

- **GIVEN** `PORT=9000`
- **WHEN** a user runs `scorecard serve --listen 0.0.0.0`
- **THEN** the server listens on `0.0.0.0:9000`

#### Scenario: Listen host and port

- **GIVEN** `PORT=9000`
- **WHEN** a user runs `scorecard serve --listen 0.0.0.0:9001`
- **THEN** the server listens on `0.0.0.0:9001`

#### Scenario: Port unavailable

- **GIVEN** the port is already in use
- **WHEN** a user runs `scorecard serve`
- **THEN** the command exits with a non-zero status and an error naming
  the address
- **AND** no requests are served

### Requirement: Scan Endpoint Routing

The system SHALL run a scan for `GET` and `POST` requests to any path other
than `/health`. Requests with any other method SHALL receive status 404.

#### Scenario: GET on root

- **WHEN** a client sends
  `GET /?repo=github.com/ossf/scorecard`
- **THEN** the server scans `github.com/ossf/scorecard`

#### Scenario: Unsupported method

- **WHEN** a client sends `PUT /`
- **THEN** the server responds with status 404

#### Scenario: OPTIONS request

- **WHEN** a client sends `OPTIONS /`
- **THEN** the server responds with status 404 and no scan is performed

### Requirement: Cross-Origin Requests

The system MUST NOT add `Access-Control-Allow-*` headers to any response.
A request is cross-site when its `Sec-Fetch-Site` header is present and is
neither `same-origin` nor `none`. Without that header, a request is
cross-site when its `Origin` host and port differ from the `Host` header.
A cross-site request on any path and method MUST receive status 403 and
`cross-origin request rejected`, and no scan runs.

#### Scenario: Preflight request

- **WHEN** a browser on another site sends `OPTIONS /`
- **THEN** the response has no `Access-Control-Allow-Origin` header

#### Scenario: Cross-site GET

- **WHEN** a client sends `GET /?repo=owner/repo` with
  `Sec-Fetch-Site: cross-site`
- **THEN** the server responds with status 403 and no scan is performed

#### Scenario: Mismatched Origin without Sec-Fetch-Site

- **WHEN** a client sends `POST /` with `Host: 127.0.0.1:8080` and
  `Origin: http://example.com`
- **THEN** the server responds with status 403 and no scan is performed

#### Scenario: Address typed into the browser

- **WHEN** a client sends `GET /?repo=owner/repo` with
  `Sec-Fetch-Site: none`
- **THEN** the server runs the scan

#### Scenario: Non-browser client

- **WHEN** a client sends `GET /?repo=owner/repo` with neither
  `Sec-Fetch-Site` nor `Origin`
- **THEN** the server runs the scan

### Requirement: Trust Model in Help Text

The `serve` command's long help text SHALL state that the server has no
authentication or rate limiting and runs every scan with the server's own
credentials. It SHALL state that the server listens on `127.0.0.1` unless
`--listen` names another address. It SHALL recommend using `--listen`
only on a trusted network, with a token limited to the repositories the
server is meant to scan.

#### Scenario: Help text states the trust model

- **WHEN** a user runs `scorecard serve --help`
- **THEN** the output states that the server has no authentication or rate
  limiting
- **AND** the output states that scans run with the server's credentials
- **AND** the output states that the server listens on `127.0.0.1` by
  default
- **AND** the output recommends a trusted network and a limited token

## ADDED Requirements

### Requirement: Non-Loopback Startup Message

A host is loopback when it is `localhost` or a loopback IP address; an
empty host is not. When the listening host is not loopback, the system
SHALL log at startup that the server has no authentication and is
reachable from other hosts.

#### Scenario: All interfaces

- **WHEN** a user runs `scorecard serve --listen 0.0.0.0`
- **THEN** the server logs that it has no authentication and is reachable
  from other hosts

#### Scenario: Loopback

- **WHEN** a user runs `scorecard serve`
- **THEN** the server does not log that message

### Requirement: Loopback Host Check

When the listening host is loopback, a request whose `Host` header names
a host that is not loopback MUST receive status 403 and `host not
allowed`, and no scan runs. The port in `Host` SHALL be ignored, and names
SHALL be compared without regard to case.

#### Scenario: Other host name

- **GIVEN** the server listens on `127.0.0.1:8080`
- **WHEN** a client sends `GET /?repo=owner/repo` with
  `Host: other.example:8080`
- **THEN** the server responds with status 403 and no scan is performed

#### Scenario: Loopback names

- **GIVEN** the server listens on `127.0.0.1:8080`
- **WHEN** a client sends `GET /health` with `Host: localhost:8080`,
  `Host: 127.0.0.1:8080`, or `Host: [::1]:8080`
- **THEN** the server responds with status 200

#### Scenario: Non-loopback listener

- **GIVEN** the server listens on `0.0.0.0:8080`
- **WHEN** a client sends `GET /health` with `Host: scorecard.internal`
- **THEN** the server responds with status 200

### Requirement: Request Body Size Limit

The system MUST respond with status 413 and `request body too large`,
without running a scan, to a `POST` request whose body exceeds 1 MiB.

#### Scenario: Oversized body

- **WHEN** a client posts a 2 MiB body
- **THEN** the server responds with status 413 and no scan is performed
