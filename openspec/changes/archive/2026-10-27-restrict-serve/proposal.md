# Proposal: Restrict `scorecard serve` to the local machine by default

## Why

Current behavior: `scorecard serve` has no authentication, runs every scan
with the operator's credentials, listens on all interfaces, and accepts
requests from any origin. The help text and the threat model
(`docs/security/threat-catalog.yaml`, OSSF.SCORECARD.ENGINE.THR10) ask
operators to keep it on a trusted network. A server like this should be
local by default, and exposing it should be an explicit choice.

## What Changes

- **BREAKING** Target state: the server listens on `127.0.0.1` unless
  `--listen` names another host, or a host and port. `PORT` still sets the
  port. Deployments that reach the server from other hosts, including
  containers with a published port, need `--listen`, for example
  `scorecard serve --listen 0.0.0.0`.
- Target state: when the host is not a loopback address, the server logs a
  startup message stating that it has no authentication and is reachable
  from other hosts.
- **BREAKING** Target state: the server sends no CORS headers and rejects
  requests a browser marks as cross-site, on every method. `OPTIONS` gets
  404 like other unsupported methods. Current behavior allows any origin.
- Target state: on a loopback address, the server rejects requests whose
  `Host` is not `localhost` or a loopback IP address.
- Target state: `POST` bodies over 1 MiB get 413.
- Target state: if the server can't listen on its address, the command
  exits with a non-zero status. Current behavior logs the error and waits
  for a signal.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `openspec/specs/http-server`: `Listening Address`, `Scan Endpoint
  Routing`, `Cross-Origin Requests`, and `Trust Model in Help Text`
  change; `Non-Loopback Startup Message`, `Loopback Host Check`, and
  `Request Body Size Limit` are added.
- `openspec/specs/cli`: `Command Tree` adds the `serve` flag `--listen`.

## Impact

- `cmd/serve.go`: listener, middleware, request decoding, help text.
- `docs/security/threat-catalog.yaml` (THR10) and
  `docs/security/capability-catalog.yaml`: the default listening address.
- No change to scan results, checks, probes, or output formats.
