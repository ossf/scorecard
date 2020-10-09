package checks

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"strings"

	"github.com/dlorenc/scorecard/checker"
)

func init() {
	AllChecks = append(AllChecks, NamedCheck{
		Name: "Frozen-Deps",
		Fn:   FrozenDeps,
	})
}

func FrozenDeps(c *checker.Checker) CheckResult {
	r, _, err := c.Client.Repositories.Get(c.Ctx, c.Owner, c.Repo)
	if err != nil {
		return RetryResult(err)
	}
	url := r.GetArchiveURL()
	url = strings.Replace(url, "{archive_format}", "tarball", 1)
	url = strings.Replace(url, "{/ref}", r.GetDefaultBranch(), 1)

	// Download
	resp, err := c.HttpClient.Get(url)
	if err != nil {
		return RetryResult(err)
	}
	defer resp.Body.Close()

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return RetryResult(err)
	}
	tr := tar.NewReader(gz)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return RetryResult(err)
		}

		// Strip the repo name
		names := strings.SplitN(hdr.Name, "/", 2)
		if len(names) < 2 {
			continue
		}

		name := names[1]

		switch strings.ToLower(name) {
		case "go.mod", "go.sum":
			return CheckResult{
				Pass:       true,
				Confidence: 10,
			}
		case "vendor/", "third_party/":
			return CheckResult{
				Pass:       true,
				Confidence: 10,
			}
		case "package-lock.json":
			return CheckResult{
				Pass:       true,
				Confidence: 10,
			}
		case "requirements.txt":
			return CheckResult{
				Pass:       true,
				Confidence: 10,
			}
		case "gemfile.lock":
			return CheckResult{
				Pass:       true,
				Confidence: 10,
			}
		}
	}
	return CheckResult{
		Pass:       false,
		Confidence: 5,
	}
}
