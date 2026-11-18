# Proposal: Send the GitLab token only to its own instance

## Why

Current behavior: `GITLAB_AUTH_TOKEN` is sent to every host Scorecard
treats as GitLab. That includes any host whose name contains `gitlab.` and
any host that answers the project list API, so a repository URL, or a
package's registry metadata, can direct the token to a host its owner
controls. A GitLab token belongs to one instance. GitLab's CLI sends an
environment token only to the host named by `GITLAB_HOST` (alias
`GL_HOST`).

## What Changes

- Target state: the token is sent only to the token host, which is
  `GL_HOST` when set and `gitlab.com` otherwise. Every other GitLab host is
  still scanned, without the token, which GitLab serves as public data.
- **BREAKING** for private projects on a self-managed instance that relied
  on the token without setting `GL_HOST`: they now need `GL_HOST`.
- Target state: GitLab URIs must use `https`. Current behavior accepts
  `http` for the instance check, while every later request already uses
  `https`.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `openspec/specs/repo-clients`: `GitLab Repository URI Parsing` requires
  `https`; `GitLab Authentication and Endpoints` limits the token to the
  token host and records the existing redirect rule.

## Impact

- `clients/gitlabrepo`: REST client, GraphQL client, archive download, and
  `Repo.IsValid`.
- `README.md`: the self-hosted GitLab section.
- No change to checks, probes, scores, or output formats.
