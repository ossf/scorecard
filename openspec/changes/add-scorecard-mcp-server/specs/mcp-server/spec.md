## ADDED Requirements

### Requirement: MCP server runtime over stdio
The server SHALL run as an MCP server communicating over stdio, completing the MCP
initialization handshake and serving its registered tools and resources until the
client disconnects or the process receives a termination signal.

#### Scenario: Server initializes and advertises capabilities
- **WHEN** an MCP client connects over stdio and sends the initialize request
- **THEN** the server SHALL complete the handshake and advertise its tools and resources

#### Scenario: Graceful shutdown
- **WHEN** the client disconnects or the process receives SIGINT or SIGTERM
- **THEN** the server SHALL stop serving and exit without error

### Requirement: Transport-agnostic server construction
The server SHALL be constructed by a single factory that registers all tools,
resources, and instructions independently of the transport, so that an alternative
transport (such as Streamable HTTP) can be added later without modifying any tool or
resource handler.

#### Scenario: Server core served over stdio
- **WHEN** the server is started with the stdio transport
- **THEN** all factory-registered tools and resources SHALL be available to the client

#### Scenario: Adding a transport requires no handler changes
- **WHEN** a new transport binding is introduced
- **THEN** it SHALL reuse the same factory-registered tools and resources without changing their handlers

### Requirement: Repository reference parsing
The server SHALL accept repository references of the form `platform/owner/repo`,
defaulting the platform to `github.com` when omitted, accepting `gitlab.com`,
tolerating a trailing slash, and rejecting malformed references with a clear error.

#### Scenario: Owner and repo default to GitHub
- **WHEN** a tool receives `owner/repo` with no platform segment
- **THEN** the server SHALL interpret the platform as `github.com`

#### Scenario: Explicit GitLab platform
- **WHEN** a tool receives `gitlab.com/owner/repo`
- **THEN** the server SHALL target the GitLab platform

#### Scenario: Malformed reference is rejected
- **WHEN** a tool receives a reference that is not of the form `platform/owner/repo`
- **THEN** the server SHALL return a validation error describing the expected format

### Requirement: Structured tool output
Every tool SHALL return structured output using a declared output schema and
structured content, and SHALL also include an equivalent JSON text representation for
hosts that do not consume structured content.

#### Scenario: Tool returns structured and text content
- **WHEN** a tool completes successfully
- **THEN** the response SHALL include structured content matching the tool's output schema
- **AND** the response SHALL include an equivalent JSON text representation

### Requirement: Consistent error handling
The server SHALL report recoverable failures as MCP tool errors that include an
actionable hint, and SHALL NOT terminate the transport on such failures.

#### Scenario: Recoverable error includes a hint and keeps serving
- **WHEN** a tool cannot complete (for example, an unknown repository or an upstream error)
- **THEN** the server SHALL return an MCP tool error whose message explains the cause and suggests a next step
- **AND** the server SHALL remain running

### Requirement: Responsible-AI framing
The server SHALL provide framing via the MCP `instructions` field stating that
Scorecard results are heuristic signals rather than a verdict, and SHALL NOT assert
that a repository "is secure" or "is insecure." Tool descriptions SHALL NOT contain
instructions directing the model's behavior.

#### Scenario: Framing present in server instructions
- **WHEN** a client initializes the server
- **THEN** the advertised instructions SHALL describe results as heuristic signals and note cached-data caveats

#### Scenario: Output states no verdict
- **WHEN** any tool returns results
- **THEN** the output SHALL present signals and SHALL NOT state that the repository is secure or insecure

### Requirement: Tool annotations and read-only surface
Every tool SHALL declare the annotations required for Anthropic MCP Directory review —
`readOnlyHint: true`, `destructiveHint: false`, and a human-readable `title` — and
SHALL additionally declare `idempotentHint: true`. Tools that access the network SHALL
also declare `openWorldHint: true`. Every tool name SHALL be at most 64 characters, and
the server SHALL expose only read operations, satisfying the read/write separation
required by Directory review.

#### Scenario: Tools carry the required annotations
- **WHEN** the server lists its tools
- **THEN** each tool SHALL carry `readOnlyHint: true`, `destructiveHint: false`, `idempotentHint: true`, and a `title`
- **AND** network-accessing tools SHALL additionally carry `openWorldHint: true`

#### Scenario: No write operations are exposed
- **WHEN** the server lists its tools
- **THEN** every tool SHALL be a read operation, with no tool performing a create, update, or delete

### Requirement: Constrained and documented tool inputs
Tool input parameters SHALL be validated with constrained schemas, and each parameter
SHALL carry a description. In particular, the platform SHALL be constrained to the
supported set (`github.com`, `gitlab.com`); a supplied commit SHA SHALL be constrained
to a 40-character hexadecimal pattern; and the repository count accepted by the
comparison tool SHALL be bounded by an explicit maximum.

#### Scenario: Invalid commit format is rejected before any provider call
- **WHEN** a client supplies a commit argument that is not 40 hexadecimal characters
- **THEN** the server SHALL reject the call with a validation error before contacting any provider

#### Scenario: Unsupported platform is rejected
- **WHEN** a client supplies a platform outside the supported set
- **THEN** the server SHALL reject the call with a validation error naming the supported platforms

### Requirement: Self-describing tool descriptions
Each tool description SHALL state what the tool does, what it returns, and what it does
not do; SHALL disambiguate sibling tools where their behavior overlaps; and SHALL
reference the OpenSSF Scorecard documentation or REST API as the source of its data.

#### Scenario: Description acts as a contract
- **WHEN** the server lists a tool
- **THEN** its description SHALL convey the tool's purpose, its return shape, and its limitations, and SHALL reference the Scorecard documentation or API
