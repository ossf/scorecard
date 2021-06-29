// Copyright 2021 Security Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package checks

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/go-github/v32/github"

	"github.com/ossf/scorecard/checker"
)

const (
	// CheckVulnerabilities is the registered name for the OSV check.
	CheckVulnerabilities = "Vulnerabilities"
	osvQueryEndpoint     = "https://api.osv.dev/v1/query"
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
	registerCheck(CheckVulnerabilities, HasUnfixedVulnerabilities)
}

func HasUnfixedVulnerabilities(c *checker.CheckRequest) checker.CheckResult {
	commits, _, err := c.Client.Repositories.ListCommits(c.Ctx, c.Owner, c.Repo, &github.CommitsListOptions{
		ListOptions: github.ListOptions{
			PerPage: 1,
		},
	})
	if err != nil {
		return checker.MakeRetryResult(CheckVulnerabilities, err)
	}

	if len(commits) != 1 || commits[0].SHA == nil {
		return checker.MakeInconclusiveResult(CheckVulnerabilities, ErrNoCommits)
	}

	query, err := json.Marshal(&osvQuery{
		Commit: *commits[0].SHA,
	})
	if err != nil {
		panic("!! failed to marshal OSV query.")
	}

	req, err := http.NewRequestWithContext(c.Ctx, http.MethodPost, osvQueryEndpoint, bytes.NewReader(query))
	if err != nil {
		return checker.MakeRetryResult(CheckVulnerabilities, err)
	}

	// Use our own http client as the one from CheckResult adds GitHub tokens to the headers.
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return checker.MakeRetryResult(CheckVulnerabilities, err)
	}
	defer resp.Body.Close()

	var osvResp osvResponse
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&osvResp); err != nil {
		return checker.MakeRetryResult(CheckVulnerabilities, err)
	}

	if len(osvResp.Vulns) > 0 {
		for _, vuln := range osvResp.Vulns {
			c.Logf("HEAD is vulnerable to %s", vuln.ID)
		}
		return checker.MakeFailResult(CheckVulnerabilities, nil)
	}

	return checker.MakePassResult(CheckVulnerabilities)
}
