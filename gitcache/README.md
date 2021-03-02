# gitcache - Reducing the GitHub API usage for scaling scorecard


To scale 1000's of repositories it is important that the codebase is cognizant on the usage of the GitHub API's.

One of the stratergies to avoid using the API directly cloning the repository.

1. Clone the repository anonymously (not using GitHub API using token).
2. Tarball the default branch and compress it.
3. Store the compressed file into a blob store GCS.
4. Store the last commit date with in the blob.



## How is the above going to scale scorecard scans?


Scorecard checks for these don't need GitHub API, it requires a Git API.
1. Active - Checks for the last commit in 90 days. This is can be fetched from the above blob storage.
1. Frozen-Deps - Checks for a file with in the tarball.
1. CodeQLInCheckDefinitions - Checks for a file content within the tarball.
1. Security-Policy - Checks for a file name within the tarball.
1. Packaging - Checks for a file content within the tarball.

The number of checks within scorecard https://github.com/ossf/scorecard#checks is `14` out of which `12` use GitHub API's. The two that does not use the GitHub API's are `Fuzzing` and `CII`.

With the above stratergy we have reduced around **41%** of the GitHub API calls.


https://github.com/ossf/scorecard/issues/202


## Is this an optional feature? Can the scorecard run without this option?

This is a separate package that has to ran to populate the blob storage. The scorecard will have a flag to either load from the `gitcache` which is the blob store or use the standard API

## What does the gitcache do? 
![image](https://user-images.githubusercontent.com/172697/109535822-fcc86880-7a8a-11eb-86c0-314c3cdcecfb.png)
- It fetches the git repository only with the **lastcommit**
- Extracts the `lastcommit` time.
- Removes the `.git` folder.
- `tar` and `gzip` the folder.
- Updates the `blob` with 
    - `lastcommit` - time of the last commit.
    - `lastsynctime` - The time that this git repository was synced with the blob cache.
    - `tar` - The tarball of the git repository.

## Can this be used for Non-GitHub repositories like Gitlab?

Yes, this can be used for GitLab or other Non-GitHub repositories.

## Can gitcache be scaled?

Yes, gitcache can be scaled because it does not hold any state or need to be throttled by an exteranl API. gitcache would be exposed as an HTTP API which can be deployed with in `k8s` with autoscaling capabilities.


## How can scorecard determine when was the last sync?

gitcache stores last sync for each repository within its folder.

![](https://i.imgur.com/dWszY76.png)


## Can gitcache be extended to expose other properties of the Git Repo?

Yes, it can be extended for the other properties of the git repo, because it stored as key/value pairs.


## Has this been tested with real data?

Yes, this has been tested with real data. It is syncing `2000` git repositories from _projects.txt_ which is within the `cron` folder.

![](https://i.imgur.com/xLwy7jx.png)

The bottom of the picture shows `588` folders.

## Can this be used for non-scorecard repoistories?

Yes, this data can be used by other projects like https://github.com/ossf/criticality_score
