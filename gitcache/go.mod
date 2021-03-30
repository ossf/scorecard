module github.com/ossf/scorecard/gitcache

go 1.15

require (
	github.com/go-git/go-git/v5 v5.3.0
	github.com/go-vela/archiver v2.1.0+incompatible // indirect
	github.com/mholt/archiver v3.1.1+incompatible
	github.com/onsi/ginkgo v1.15.2
	github.com/onsi/gomega v1.11.0
	github.com/pierrec/lz4 v2.6.0+incompatible // indirect
	github.com/pkg/errors v0.9.1
	go.uber.org/zap v1.16.0
	gocloud.dev v0.22.0
	golang.org/x/tools v0.1.0 // indirect
)

// The originial repository does not handle symlinks and there is an open PR https://github.com/mholt/archiver/pull/230.
// The replace uses the fork that has the fix for the above issue.
replace github.com/mholt/archiver => github.com/go-vela/archiver v1.1.3-0.20200811184543-d2452770f58c
