# http-server Specification

## Purpose

The `scorecard serve` subcommand exposes Scorecard scans over HTTP. A client
sends a repository or package name with optional scan settings and receives
the scan result as JSON. This spec covers the listening address, endpoints,
request parameters, response format, error responses, middleware behavior,
and shutdown.

## Requirements

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

### Requirement: Health Endpoint

The system SHALL respond to requests for `/health` with status 200 and an
empty body without running a scan.

#### Scenario: Health check

- **WHEN** a client sends `GET /health`
- **THEN** the server responds with status 200 and no scan is performed

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

### Requirement: Query Parameters for GET

For `GET` requests the system SHALL read the query parameters `repo`,
`npm`, `pypi`, `rubygems`, `nuget`, `commit`, `file_mode`, `checks`
(comma-separated), `probes` (comma-separated), `show_details`, and
`show_annotations`. `show_details` and `show_annotations` SHALL be enabled
only by the exact value `true`. `GET` requests SHALL NOT accept a commit
depth.

#### Scenario: Select checks by query

- **WHEN** a client sends
  `GET /?repo=owner/repo&checks=License,SAST&show_details=true`
- **THEN** the server reads `License` and `SAST` as the requested checks and
  enables details

#### Scenario: Non-true boolean

- **WHEN** a client sends `show_details=1`
- **THEN** details are not included

### Requirement: JSON Body for POST

For `POST` requests the system SHALL decode a JSON body with the optional
fields `repo`, `npm`, `pypi`, `rubygems`, `nuget`, `commit`, `file_mode`,
`checks` (array), `probes` (array), `commit_depth` (integer),
`show_details` (boolean), and `show_annotations` (boolean). A body that is
not valid JSON MUST receive status 400 with the text `invalid request body`.

#### Scenario: POST scan request

- **WHEN** a client posts
  `{"repo": "owner/repo", "checks": ["License"], "commit_depth": 10}`
- **THEN** the server runs the License check with a commit depth of 10

#### Scenario: Malformed body

- **WHEN** a client posts `{not json`
- **THEN** the server responds with status 400 and `invalid request body`

### Requirement: Request Defaults

The system SHALL default an empty `commit` to `HEAD` and an empty
`file_mode` to `archive`. An empty or absent check list SHALL select every
check that applies to the commit and repository type. When the commit is
not `HEAD`, the default set SHALL contain only checks supporting
commit-based analysis, and an explicitly requested check without
commit-based support SHALL fail the request with status 500 and a body
beginning `GetEnabled:` that contains `does not support commit-based data`. Checks
whose documentation excludes the target repository's host type SHALL be
skipped. When the internal GitHub integration check override is active,
its check list SHALL replace the `checks` given in the request; the
override is defined in `cli`.

#### Scenario: No checks requested

- **WHEN** a client posts `{"repo": "owner/repo"}`
- **THEN** the server runs every check supported for a GitHub repository at
  `HEAD`

#### Scenario: Check not supported on the host

- **GIVEN** a requested check does not support GitLab repositories
- **WHEN** a client requests that check for a GitLab repository
- **THEN** the check is omitted from the result

#### Scenario: Default set at a non-HEAD commit

- **WHEN** a client posts `{"repo": "owner/repo", "commit": "<sha>"}`
- **THEN** the server runs only the checks that support commit-based
  analysis

#### Scenario: Explicit check without commit support at a non-HEAD commit

- **GIVEN** a check that does not support commit-based analysis
- **WHEN** a client requests that check with a `commit` other than `HEAD`
- **THEN** the server responds with status 500 and a body beginning
  `GetEnabled:` that contains `does not support commit-based data`

#### Scenario: GitHub integration override replaces requested checks

- **GIVEN** the internal GitHub integration check override is active
- **WHEN** a client posts `{"repo": "owner/repo", "checks": ["License"]}`
- **THEN** the server selects checks from the override list instead of
  `License`

### Requirement: Request Validation

The system MUST require exactly one of `repo`, `npm`, `pypi`, `rubygems`, or
`nuget` and MUST reject an unsupported `file_mode`. A request failing
validation MUST receive status 400 with a body beginning
`invalid options:` followed by the validation errors.

#### Scenario: No repository

- **WHEN** a client sends `GET /` with no input parameter
- **THEN** the server responds with status 400 and a body containing
  "exactly one of"

#### Scenario: Repository and package together

- **WHEN** a client posts `{"repo": "owner/repo", "npm": "left-pad"}`
- **THEN** the server responds with status 400

#### Scenario: Bad file mode

- **WHEN** a client sends `file_mode=svn`
- **THEN** the server responds with status 400 and "unsupported file mode"

### Requirement: Probe Selection Ignored

The system SHALL accept a `probes` value in `GET` and `POST` requests but
MUST NOT use it: every scan runs checks, never individual probes.

#### Scenario: Probes by POST

- **WHEN** a client posts `{"repo": "owner/repo", "probes": ["hasLicenseFile"]}`
- **THEN** the server runs checks as if `probes` were absent
- **AND** the response contains check results, not probe findings

#### Scenario: GET without probes

- **WHEN** a client sends `GET /?repo=owner/repo`
- **THEN** the server runs checks and responds with status 200

### Requirement: Successful Response

On success the system SHALL respond with status 200, the header
`Content-Type: application/json`, and the scan result in the same JSON
structure as the CLI `--format=json` output, with checks sorted by name.
Details and annotations SHALL be included only when requested.

#### Scenario: Successful scan

- **WHEN** a valid scan completes
- **THEN** the response is status 200 with `Content-Type: application/json`
- **AND** the body lists checks in alphabetical order

### Requirement: Server-Side Error Responses

The system SHALL respond with status 500 and a plain-text message when a
request passes validation but cannot be served. The message SHALL begin
with a prefix naming the failed step:

- `fetchGitRepositoryFromPackageManagers:` when package resolution fails
- `making remote repo:` when the repository cannot be parsed
- `cannot read yaml file:` when the check documentation cannot be read
- `GetEnabled:` when check selection fails
- `scorecard.Run:` when the scan fails
- `failed to format results:` when the result cannot be serialized

#### Scenario: Unknown check

- **WHEN** a client posts `{"repo": "owner/repo", "checks": ["Nope"]}`
- **THEN** the server responds with status 500 and a body beginning
  `GetEnabled:` that contains `"Nope" is not a known check`

#### Scenario: Unparseable repository

- **WHEN** a client posts `{"repo": "::bad::"}`
- **THEN** the server responds with status 500 and a body beginning
  `making remote repo:`

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

### Requirement: Panic Recovery and Request Logging

The system SHALL recover from a panic while handling a request, log it, and
respond with status 500 and `internal server error`. The system SHALL log
each completed request with its method, path, and duration.

#### Scenario: Handler panic

- **WHEN** handling a request panics
- **THEN** the server responds with status 500 and `internal server error`
- **AND** continues serving later requests

#### Scenario: Request log line

- **WHEN** a request completes
- **THEN** the server logs
  `request completed method=<method> path=<path> duration=<duration>`

### Requirement: Graceful Shutdown

The system SHALL begin shutting down when it receives `SIGINT` or `SIGTERM`,
log `Shutting down server...`, and allow in-flight requests up to 5 seconds
to finish. If shutdown does not complete in time, the command MUST exit
with a non-zero status and a `server shutdown:` error.

#### Scenario: Clean shutdown

- **GIVEN** the server is idle
- **WHEN** it receives `SIGTERM`
- **THEN** it logs `Shutting down server...` and exits with status 0

#### Scenario: Shutdown timeout

- **GIVEN** a scan is still running
- **WHEN** the server receives `SIGINT` and the scan does not finish within
  5 seconds
- **THEN** the command exits with a non-zero status and a
  `server shutdown:` error

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
