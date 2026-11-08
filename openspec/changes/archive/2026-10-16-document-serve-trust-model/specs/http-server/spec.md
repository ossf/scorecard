## ADDED Requirements

### Requirement: Trust Model in Help Text

The `serve` command's long help text SHALL state that the server has no
authentication or rate limiting, runs every scan with the server's own
credentials, and listens on all interfaces by default. It SHALL recommend
running the server only on a trusted network, with a token limited to the
repositories it is meant to scan.

#### Scenario: Help text states the trust model

- **WHEN** a user runs `scorecard serve --help`
- **THEN** the output states that the server has no authentication or rate
  limiting
- **AND** the output states that scans run with the server's credentials
- **AND** the output recommends a trusted network and a limited token
