# Enable OSSF Scorecard Action at an Organization Level

This tool will add the [OpenSSF's Scorecard workflow](https://github.com/ossf/scorecard-action) to all accessible repositories under a given organization. A PR will be created so that owners can decide whether or not they want to include the workflow.

## Setup

Running this tool requires three parameters, which are defined at the top of `org-workflow-add.go`:
1. ORG_NAME - the name of the organization for which the workflow should be enabled.
2. PAT - a Personal Access Token with the following scopes:
    - `repo > public_repo`
    - `admin:org > read:org`
3. REPO_LIST (OPTIONAL) - repository names under the organization that the workflow should be added to. If not provided, every repository will be updated.

Another PAT should also be defined as an organization secret for `scorecards-analysis.yml` using steps listed in [scorecard-action](https://github.com/ossf/scorecard-action#pat-token-creation).

## Execution

Execute this process by running `go run org-workflow-add.go` in the command line. Output will be produced for each successfully updated repository.
