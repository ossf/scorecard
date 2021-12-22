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
	"fmt"
	"net/http"
	"strings"

	"github.com/ossf/scorecard/v3/checker"
	sce "github.com/ossf/scorecard/v3/errors"
)

const (
	// CheckVulnerabilities is the registered name for the OSV check.
	CheckVulnerabilities = "Vulnerabilities"
	osvQueryEndpoint     = "https://api.osv.dev/v1/query"
)

type osvQuery struct {
	Commit string `json:"commit"`
}

type osvResponse struct {
	Vulns []struct {
		ID string `json:"id"`
	} `json:"vulns"`
}

// Vulnerabilities cheks for vulnerabilities in api.osv.dev.
type Vulnerabilities interface {
	HasUnfixedVulnerabilities(c *checker.CheckRequest) checker.CheckResult
}
type vulns struct{}

//nolint:gochecknoinits
func init() {
	v := &vulns{}
	registerCheck(CheckVulnerabilities, v.HasUnfixedVulnerabilities)
}

func (resp *osvResponse) getVulnerabilities() []string {
	ids := make([]string, 0, len(resp.Vulns))
	for _, vuln := range resp.Vulns {
		ids = append(ids, vuln.ID)
	}
	return ids
}

// NewVulnerabilities creates a new Vulnerabilities check.
func NewVulnerabilities() Vulnerabilities {
	return &vulns{}
}

// HasUnfixedVulnerabilities runs Vulnerabilities check.
func (v *vulns) HasUnfixedVulnerabilities(c *checker.CheckRequest) checker.CheckResult {
	commits, err := c.RepoClient.ListCommits()
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "Client.Repositories.ListCommits")
		return checker.CreateRuntimeErrorResult(CheckVulnerabilities, e)
	}

	if len(commits) < 1 || commits[0].SHA == "" {
		return checker.CreateInconclusiveResult(CheckVulnerabilities, "no commits found")
	}

	query, err := json.Marshal(&osvQuery{
		Commit: commits[0].SHA,
	})
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "json.Marshal")
		return checker.CreateRuntimeErrorResult(CheckVulnerabilities, e)
	}

	req, err := http.NewRequestWithContext(c.Ctx, http.MethodPost, osvQueryEndpoint, bytes.NewReader(query))
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("http.NewRequestWithContext: %v", err))
		return checker.CreateRuntimeErrorResult(CheckVulnerabilities, e)
	}

	// Use our own http client as the one from CheckRequest adds GitHub tokens to the headers.
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("httpClient.Do: %v", err))
		return checker.CreateRuntimeErrorResult(CheckVulnerabilities, e)
	}
	defer resp.Body.Close()

	var osvResp osvResponse
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&osvResp); err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("decoder.Decode: %v", err))
		return checker.CreateRuntimeErrorResult(CheckVulnerabilities, e)
	}

	// TODO: take severity into account.
	vulnIDs := osvResp.getVulnerabilities()
	if len(vulnIDs) > 0 {
		c.Dlogger.Warn3(&checker.LogMessage{
			Text: fmt.Sprintf("HEAD is vulnerable to %s", strings.Join(vulnIDs, ", ")),
		})
		return checker.CreateMinScoreResult(CheckVulnerabilities, "existing vulnerabilities detected")
	}

	return checker.CreateMaxScoreResult(CheckVulnerabilities, "no vulnerabilities detected")
}
