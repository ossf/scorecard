## Caching

Scorecard uses `httpcache` with
<https://docs.github.com/en/rest/overview/resources-in-the-rest-api#conditional-requests>
for caching httpresponse. The default cache is in-memory.

Some details on caching
<https://github.com/ossf/scorecard/issues/80#issuecomment-782723182>

### Blob Cache

Scorecard results can be cached into a blob for increasing throughput for
subsequent runs.

To use blob cache two env variables have to be set `USE_BLOB_CACHE=true` and
`BLOB_URL=gs://scorecards-cache/`.

The code uses <https://github.com/google/go-cloud> for blob caching. It is
compatible with GCS,S3 and Azure blob.

### Disk Cache

Scorecard results can be cached into a disk for increasing throughput for
subsequent runs.

To use disk cache two env variables have to be set `USE_DISK_CACHE=true` and
`DISK_CACHE_PATH=./cache`.

There is no TTL on cache.

The default cache size is 10GB.
