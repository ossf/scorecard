# Scorecard CDN

-   Status: Complete

## Background

Serving the weekly scan results at `api.scorecard.dev` represents a sizable
chunk of our monthly cloud bill, both in terms of number of requests (30 million
per week) and bandwidth egress (680 GiB). These results are largely static,
updating once per week for most repositories, representing a great opportunity
to use a CDN. The OpenSSF Scorecard project successfully applied to Fastly’s
Fast Forward program, so we have CDN credits available for use.

The Scorecard API has three main endpoints:
- `GET /projects/{platform}/{org}/{repo}`: Get a repository's Scorecard
- `POST /projects/{platform}/{org}/{repo}`: Publish results from Scorecard Action
- `GET /projects/{platform}/{org}/{repo}/badge`: Get a repository's Scorecard badge

This document focuses on the two GET endpoints, as we don’t want to prevent
repositories from uploading their latest scores with Scorecard Action.

## Request Traffic

In the last week, the Scorecard API received 30.1M requests (50 QPS), the
traffic can be classified simply by its HTTP response status code:

| Response Code | Description | Requests | Update Frequency |
| --- | --- | --- | --- |
| 200 | Result request for a project in our corpus | 17.7M (59%) | 7 days if published by weekly scan, or variable if by Scorecard Action. |
| 404 | Result request for a project not in our corpus | 11.6M (38%) | Never, until a project is added to the weekly scans or installs Scorecard Action |
| 302 | Badge request | 0.83M (3%) | Never, assuming we don’t change providers |

## Caching Strategy

### DNS Entries

We modified our DNS to connect our app to Fastly. We did this in two phases:
  1. Added a new CNAME DNS entry (`cdn.scorecard.dev`), so we could test the CDN
without affecting our normal traffic 
```cdn.scorecard.dev. 3600 IN CNAME
dualstack.<letter>.sni.global.fastly.net.
```

  2. Changed the `api.scorecard.dev`
CNAME entry away from the Cloud Run entry, towards Fastly instead.
```
api.scorecard.dev. 3600 IN CNAME dualstack.<letter>.sni.global.fastly.net.
```

### Time-to-Live (TTL)

Our caching strategy is to cache all responses for the maximum TTL supported by
Fastly (1 year). We rely on cache purging to invalidate results when new results
are available. This provides slightly more flexibility if results ever update at
a different timeframe (either due to issues with our weekly scan, or if we
intentionally reduce frequency).

We implemented caching with the `Surrogate-Control` header to support split
policies, namely so we can set longer cache times for Fastly while maintaining
shorter cache times for browsers.

> **Note**: While uncommon, approximately 80,000 requests per week (0.3%) set
> the commit query parameter to get results from another commit. Requests
> containing the commit query parameter should be treated as unique cache keys,
> allowing them to be cached separately from the commit-less requests.

### Purging

Fastly supports purging individual URLs, as well as groups of URLs, and we make
use of both in our strategy. We purge our result endpoint URLs individually when
new results become available. Our weekly scans process approximately 10,000
repos per hour, which is below Fastly’s purge API limit of 100,000/hr. The badge
endpoint would only need to be changed if we make infrastructure changes, so
they can use a surrogate key to enable purging as a group if manually needed by
maintainers.

One important thing to note is authentication. URL purges are unauthenticated by
default in Fastly, so anyone could send a PURGE request with `curl -X PURGE
<URL>` so we have enabled authenticated purges.

### Shielding

Fastly operates multiple points of presence (POP) across the globe, which act
independently, including when contacting the origin server. We have designated
their Chicago POP as an Origin Shield, which will route all origin requests
through a single POP, reducing origin load further.
