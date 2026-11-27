## MODIFIED Requirements

### Requirement: MCP server runtime over stdio

The server SHALL be started by `scorecard mcp` and SHALL communicate over
stdio. It SHALL complete the MCP initialization handshake and serve its
registered tools and resources until the client disconnects or the
process receives a termination signal. The server SHALL advertise the
`scorecard` build version.

#### Scenario: Server initializes and advertises capabilities

- **WHEN** an MCP client starts `scorecard mcp`, connects over stdio, and
  sends the initialize request
- **THEN** the server SHALL complete the handshake and advertise its tools
  and resources

#### Scenario: Server reports the scorecard version

- **WHEN** a client initializes the server
- **THEN** the advertised server version SHALL equal the version printed
  by `scorecard version`

#### Scenario: Graceful shutdown

- **WHEN** the client disconnects or the process receives SIGINT or SIGTERM
- **THEN** the server SHALL stop serving and exit without error
