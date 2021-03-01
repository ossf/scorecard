module github.com/ossf/scorecard/gitcache

go 1.15

require (
	github.com/go-git/go-git/v5 v5.2.0
	github.com/go-vela/archiver v2.1.0+incompatible // indirect
	github.com/mholt/archiver v3.1.1+incompatible
	github.com/pierrec/lz4 v2.6.0+incompatible // indirect
	github.com/pkg/errors v0.9.1
	gocloud.dev v0.22.0
)

// The originial repoisotry does not handle symlinks and there is an open PR https://github.com/mholt/archiver/pull/230.
// The replace uses the fork that has the fix for the above issue.
replace github.com/mholt/archiver => github.com/go-vela/archiver v1.1.3-0.20200811184543-d2452770f58c
