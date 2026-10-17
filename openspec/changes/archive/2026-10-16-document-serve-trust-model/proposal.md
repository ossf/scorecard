# Proposal: Document the `scorecard serve` trust model

## Why

`scorecard serve` has no authentication or rate limiting, listens on all
interfaces by default, and runs every scan with the server's own
credentials. Nothing tells operators this today: the command's help text,
the README, and the `http-server` spec don't mention it. An operator who
exposes the server to untrusted clients lets them spend the server's API
rate limit and read results for any private repository its token can
access. The Scorecard threat model (`docs/security/threat-catalog.yaml`,
OSSF.SCORECARD.ENGINE.THR10) records this threat; operators need the
same warning where they start the server.

## What Changes

- The long help text of `scorecard serve` states the server's trust model:
  no authentication or rate limiting, every scan runs with the server's
  credentials, and it listens on all interfaces by default, so it belongs
  on a trusted network with a token limited to the repositories it should
  scan. Current behavior: the help text says only "Start an HTTP server to
  run scorecard checks on repositories with REST API support."
- No change to request handling, authentication, or defaults.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `openspec/specs/http-server`: adds a requirement that the `serve` help
  text states the trust model.

## Impact

- `cmd/serve.go`: the `Long` help string.
- No API, flag, dependency, or behavior change beyond help output.
