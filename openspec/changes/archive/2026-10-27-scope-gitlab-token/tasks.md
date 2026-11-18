# Tasks: Send the GitLab token only to its own instance

## 1. Token host

- [x] 1.1 Add one helper that returns `GITLAB_AUTH_TOKEN` only for the
  token host (`GL_HOST`, else `gitlab.com`)
- [x] 1.2 Use it in the REST client (`CreateGitlabClient`)
- [x] 1.3 Use it in the GraphQL client; send no `Authorization` header
  without a token
- [x] 1.4 Use it for archive downloads; send no `PRIVATE-TOKEN` header
  without a token

## 2. URI parsing

- [x] 2.1 Reject non-`https` URIs in `Repo.IsValid` before the instance
  check

## 3. Tests

- [x] 3.1 One test per scenario in the delta spec, and update the redirect
  test to set the project host

## 4. Docs

- [x] 4.1 Update the self-hosted GitLab section of `README.md`
