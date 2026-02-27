# Platform Clients

## Purpose

Platform clients provide an abstraction layer over repository hosting platforms (GitHub, GitLab, Azure DevOps, local directories), allowing checks and probes to operate platform-agnostically through the `clients.RepoClient` interface.

## Requirements

### Requirement: RepoClient Interface
All platform clients SHALL implement the `clients.RepoClient` interface.

### Requirement: Platform Agnosticism
Checks and probes SHALL NOT contain platform-specific logic; all platform differences SHALL be handled within client implementations.

### Requirement: Authentication
Clients SHALL support authentication via environment variables (`GITHUB_AUTH_TOKEN`, `GITLAB_AUTH_TOKEN`, `AZURE_DEVOPS_AUTH_TOKEN`).

### Requirement: Rate Limiting
The GitHub client SHALL support round-robin token rotation for rate limit management.

## Supported Platforms

- **GitHub** (`clients/githubrepo/`) - Stable. REST and GraphQL APIs.
- **GitLab** (`clients/gitlabrepo/`) - Stable.
- **Azure DevOps** (`clients/azuredevopsrepo/`) - Experimental.
- **Local Directory** (`clients/localdir/`) - Stable. File-system-only checks.
