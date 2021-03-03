# gitcache - Reducing the GitHub API usage for scaling scorecard


To scale 1000's repositories, the codebase must be aware of the GitHub API's usage.

One of the strategies to avoid using the API directly is cloning the repository.

### In the initial run
1. Clone the repository anonymously (not using GitHub API token).
2. Tarball and compress it.
3. Store the compressed file into a blob store GCS.
4. Store the last commit date within the blob.
5. Also compress the folder without `.git` for the consumers.

### On the subsequent runs
1. pull gzip from GCS
1. unzip git repo
1. git pull origin
1. update metadata (last sync, etc.)
1. gzip, reupload to GCS





## How is the above going to scale scorecard scans?

Scorecard checks for these don't need GitHub API. It requires a Git API.
1. Active - Checks for the last commit in 90 days. This can be fetched from the above blob storage.
1. Frozen-Deps - Checks for a file with in the tarball.
1. CodeQLInCheckDefinitions - Checks for a file content within the tarball.
1. Security-Policy - Checks for a file name within the tarball.
1. Packaging - Checks for a file content within the tarball.

The number of checks within scorecard https://github.com/ossf/scorecard#checks is `14` out of which `12` use GitHub API's. The two that do not use the GitHub APIs are `Fuzzing` and `CII`.

With the above strategy, we have reduced around **41%** of the GitHub API calls.


https://github.com/ossf/scorecard/issues/202


## Is this an optional feature? Can the scorecard run without this option?

This is a separate package that has to run to populate the blob storage. The scorecard will have a flag to either load from the `gitcache`, which is the blob store, or use the standard API


## Can this be used for Non-GitHub repositories like Gitlab?

Yes, this can be used for GitLab or other Non-GitHub repositories.

## Can gitcache be scaled?

Yes, gitcache can be scaled because it does not hold any state or need to be throttled by an external API. gitcache would be exposed as an HTTP API that can be deployed in `k8s` with autoscaling capabilities.


## How can scorecard determine when the last sync was?

gitcache stores the last sync for each repository within its folder.

![](https://i.imgur.com/dWszY76.png)


## Can gitcache be extended to expose other properties of the Git Repo?

Yes, it can be extended for the other properties of the git repo, because it stored as key/value pairs.


## Has this been tested with real data?

Yes, this has been tested with real data. It is syncing `2000` git repositories from _projects.txt_ which is within the `cron` folder.

![](https://i.imgur.com/xLwy7jx.png)

The bottom of the picture shows `588` folders.
