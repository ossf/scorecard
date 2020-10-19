package checks

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"strings"

	"github.com/dlorenc/scorecard/checker"
)

func init() {
	registerCheck("Frozen-Deps", FrozenDeps)
}

var passResult = checker.CheckResult{
	Pass:       true,
	Confidence: 10,
}

func FrozenDeps(c checker.Checker) checker.CheckResult {
	r, _, err := c.Client.Repositories.Get(c.Ctx, c.Owner, c.Repo)
	if err != nil {
		return checker.RetryResult(err)
	}
	url := r.GetArchiveURL()
	url = strings.Replace(url, "{archive_format}", "tarball", 1)
	url = strings.Replace(url, "{/ref}", r.GetDefaultBranch(), 1)

	// Download
	resp, err := c.HttpClient.Get(url)
	if err != nil {
		return checker.RetryResult(err)
	}
	defer resp.Body.Close()

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return checker.RetryResult(err)
	}
	tr := tar.NewReader(gz)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return checker.RetryResult(err)
		}

		// Strip the repo name
		names := strings.SplitN(hdr.Name, "/", 2)
		if len(names) < 2 {
			continue
		}

		name := names[1]

		switch strings.ToLower(name) {
		case "go.mod", "go.sum":
			c.Logf("go modules found: %s", name)
			return passResult
		case "vendor/", "third_party/", "third-party/":
			c.Logf("vendor dir found: %s", name)
			return passResult
		case "package-lock.json":
			c.Logf("nodejs packages found: %s", name)
			return passResult
		case "requirements.txt", "pipfile.lock":
			c.Logf("python requirements found: %s", name)
			return passResult
		case "gemfile.lock":
			c.Logf("ruby gems found: %s", name)
			return passResult
		case "cargo.lock":
			c.Logf("rust crates found: %s", name)
			return passResult
		}
	}
	return checker.CheckResult{
		Pass:       false,
		Confidence: 5,
	}
}
