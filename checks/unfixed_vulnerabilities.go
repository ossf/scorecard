package checks

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/go-github/v32/github"

	"github.com/ossf/scorecard/checker"
)

const (
	// UnfixedVulnerabilities is the registered name for the OSV check.
	UnfixedVulnerabilities = "Unfixed-Vulnerabilities"
	osvQueryEndpoint       = "https://api.osv.dev/v1/query"
)

// ErrNoCommits is the error for when there are no commits found.
var ErrNoCommits = errors.New("no commits found")

type osvQuery struct {
	Commit string `json:"commit"`
}

type osvResponse struct {
	Vulns []struct {
		ID string `json:"id"`
	} `json:"vulns"`
}

//nolint:gochecknoinits
func init() {
	registerCheck(UnfixedVulnerabilities, HasUnfixedVulnerabilities)
}

func HasUnfixedVulnerabilities(c *checker.CheckRequest) checker.CheckResult {
	commits, _, err := c.Client.Repositories.ListCommits(c.Ctx, c.Owner, c.Repo, &github.CommitsListOptions{
		ListOptions: github.ListOptions{
			PerPage: 1,
		},
	})
	if err != nil {
		return checker.MakeRetryResult(UnfixedVulnerabilities, err)
	}

	if len(commits) != 1 || commits[0].SHA == nil {
		return checker.MakeInconclusiveResult(UnfixedVulnerabilities, ErrNoCommits)
	}

	query, err := json.Marshal(&osvQuery{
		Commit: *commits[0].SHA,
	})
	if err != nil {
		panic("!! failed to marshal OSV query.")
	}

	req, err := http.NewRequestWithContext(c.Ctx, http.MethodPost, osvQueryEndpoint, bytes.NewReader(query))
	if err != nil {
		return checker.MakeRetryResult(UnfixedVulnerabilities, err)
	}

	// Use our own http client as the one from CheckResult adds GitHub tokens to the headers.
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return checker.MakeRetryResult(UnfixedVulnerabilities, err)
	}
	defer resp.Body.Close()

	var osvResp osvResponse
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&osvResp); err != nil {
		return checker.MakeRetryResult(UnfixedVulnerabilities, err)
	}

	if len(osvResp.Vulns) > 0 {
		fmt.Printf("%v\n", osvResp)
		for _, vuln := range osvResp.Vulns {
			c.Logf("HEAD is vulnerable to %s", vuln.ID)
		}
		return checker.MakeFailResult(UnfixedVulnerabilities, nil)
	}

	return checker.MakePassResult(UnfixedVulnerabilities)
}
