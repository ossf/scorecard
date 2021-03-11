# Scorecard badges

The scorecard provide badges similar to other https://github.com/badges/shields OSS badges for compliance.


## Goals

* Scorecard should provide badge results based on scorecard runs on the (cron)server to ensure compliance and validation. 
* Scorecard should provide a predictable API to fetch the badges. An example could be https://somefqdn/github/ossf/scorecard/badge
* Scorecard results calculation - _TBD_ - The discussion of calculation should be a separate issue.

## Implementaion

- Scorecard badge has a separate HTTP application that generates the badge.
- Scorecard would use the Results from the scheduled cron run to generate the badge. The results of the cron are stored within the GCS bucket as `latest.json`
- The HTTP application would be stateless.


[![](https://mermaid.ink/img/eyJjb2RlIjoic2VxdWVuY2VEaWFncmFtXG5rOHMgUkVBRE1FLm1kLT4-c2NvcmVjYXJkIGJhZGdlIHNlcnZpY2U6IEdldCBzY29yZWNhcmQgYmFkZ2VcbnNjb3JlY2FyZCBiYWRnZSBzZXJ2aWNlLT4-R0NTOiBGZXRjaCBsYXRlc3QuanNvblxuc2NvcmVjYXJkIGJhZGdlIHNlcnZpY2UtPj5zY29yZWNhcmQgYmFkZ2Ugc2VydmljZTogY2FsY3VsYXRlIHNjb3JlY2FyZCBzY29yZVxuc2NvcmVjYXJkIGJhZGdlIHNlcnZpY2UtPj5rOHMgUkVBRE1FLm1kOiByZXR1cm4gdGhlIGJhZGdlLnN2Z1xuIiwibWVybWFpZCI6eyJ0aGVtZSI6ImRlZmF1bHQifSwidXBkYXRlRWRpdG9yIjpmYWxzZX0)](https://mermaid-js.github.io/mermaid-live-editor/#/edit/eyJjb2RlIjoic2VxdWVuY2VEaWFncmFtXG5rOHMgUkVBRE1FLm1kLT4-c2NvcmVjYXJkIGJhZGdlIHNlcnZpY2U6IEdldCBzY29yZWNhcmQgYmFkZ2VcbnNjb3JlY2FyZCBiYWRnZSBzZXJ2aWNlLT4-R0NTOiBGZXRjaCBsYXRlc3QuanNvblxuc2NvcmVjYXJkIGJhZGdlIHNlcnZpY2UtPj5zY29yZWNhcmQgYmFkZ2Ugc2VydmljZTogY2FsY3VsYXRlIHNjb3JlY2FyZCBzY29yZVxuc2NvcmVjYXJkIGJhZGdlIHNlcnZpY2UtPj5rOHMgUkVBRE1FLm1kOiByZXR1cm4gdGhlIGJhZGdlLnN2Z1xuIiwibWVybWFpZCI6eyJ0aGVtZSI6ImRlZmF1bHQifSwidXBkYXRlRWRpdG9yIjpmYWxzZX0)

