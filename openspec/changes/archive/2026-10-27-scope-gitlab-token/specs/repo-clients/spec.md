## MODIFIED Requirements

### Requirement: GitLab Repository URI Parsing

The GitLab client SHALL require a host and MUST reject `owner/repo` shorthand;
the first path segment SHALL be the owner and the rest the project. A
`GL_HOST` path prefix SHALL be treated as part of a matching host. A URI
without a scheme SHALL be treated as `https`, and any other scheme MUST be
rejected before a request is sent. A host containing `gitlab.` SHALL be
accepted, `github.com` MUST be rejected, and any other host SHALL be accepted
only if its project list API answers an unauthenticated request with HTTP 200.

#### Scenario: Shorthand rejected

- **WHEN** the input is `ossf/scorecard`
- **THEN** GitLab parsing fails stating a gitlab repo must specify host

#### Scenario: GL_HOST with a path prefix

- **GIVEN** `GL_HOST=https://foo.com/gitlab/`
- **WHEN** the input is `foo.com/gitlab/ssdlc/scorecard-scanner`
- **THEN** the host is `foo.com/gitlab`, owner `ssdlc`, project
  `scorecard-scanner`

#### Scenario: Self-managed instance probe

- **WHEN** the input host does not contain `gitlab.` and is not `github.com`
- **AND** the instance's project list API does not return HTTP 200
- **THEN** parsing fails with a repo unreachable error

#### Scenario: GitHub project named like GitLab

- **WHEN** the input is `https://github.com/foo/gitlab.test`
- **THEN** GitLab parsing fails because the host is not a GitLab host

#### Scenario: Plain http rejected

- **WHEN** the input is `http://gitlab.example.com/owner/project`
- **THEN** GitLab parsing fails with an invalid URL error
- **AND** no request is sent to the host

### Requirement: GitLab Authentication and Endpoints

The GitLab client SHALL use `https://<host>` for REST,
`https://<host>/api/graphql` for GraphQL, and the `PRIVATE-TOKEN` header for
archive downloads. `GITLAB_AUTH_TOKEN` SHALL be sent only to the token host:
the `GL_HOST` host with any path prefix, or `gitlab.com` when `GL_HOST` is
unset, matched exactly and ignoring case. All other requests SHALL be
unauthenticated. An archive redirect to another host or scheme MUST drop the
`PRIVATE-TOKEN` header.

#### Scenario: Token set

- **GIVEN** `GITLAB_AUTH_TOKEN` is set and `GL_HOST` is unset
- **WHEN** the GitLab client downloads the archive of a `gitlab.com` project
- **THEN** the request includes the token in the `PRIVATE-TOKEN` header

#### Scenario: Token withheld from another instance

- **GIVEN** `GITLAB_AUTH_TOKEN` is set and `GL_HOST` is unset
- **WHEN** the GitLab client sends REST, GraphQL, or archive requests for a
  project on `gitlab.example.com`
- **THEN** no request includes the token

#### Scenario: Token withheld from a look-alike host

- **GIVEN** `GITLAB_AUTH_TOKEN` is set and `GL_HOST` is unset
- **WHEN** the project host is `gitlab.com.example.net`
- **THEN** no request includes the token

#### Scenario: Token sent to GL_HOST

- **GIVEN** `GITLAB_AUTH_TOKEN` is set and `GL_HOST=gitlab.example.com`
- **WHEN** the GitLab client downloads the archive of a project on
  `gitlab.example.com`
- **THEN** the request includes the token in the `PRIVATE-TOKEN` header

#### Scenario: Token withheld from gitlab.com when GL_HOST is set

- **GIVEN** `GITLAB_AUTH_TOKEN` is set and `GL_HOST=gitlab.example.com`
- **WHEN** the GitLab client sends requests for a `gitlab.com` project
- **THEN** no request includes the token

#### Scenario: Redirect to another host

- **GIVEN** the token is sent with an archive request
- **WHEN** the response redirects to a different host or scheme
- **THEN** the redirected request does not include the `PRIVATE-TOKEN`
  header
