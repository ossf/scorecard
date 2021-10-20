// Copyright 2020 Security Scorecard Authors
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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/ossf/scorecard/v3/checker"
	sce "github.com/ossf/scorecard/v3/errors"
)

// CheckCIIBestPractices is the registered name for CIIBestPractices.
const CheckCIIBestPractices = "CII-Best-Practices"

var errTooManyRequests = errors.New("failed after exponential backoff")

//nolint:gochecknoinits
func init() {
	registerCheck(CheckCIIBestPractices, CIIBestPractices)
}

type expBackoffTransport struct {
	numRetries uint8
}

func (transport *expBackoffTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for i := 0; i < int(transport.numRetries); i++ {
		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode != http.StatusTooManyRequests {
			// nolint: wrapcheck
			return resp, err
		}
		time.Sleep(time.Duration(math.Pow(2, float64(i))) * time.Second)
	}
	return nil, errTooManyRequests
}

type response struct {
	BadgeLevel string `json:"badge_level"`
}

// CIIBestPractices runs CII-Best-Practices check.
func CIIBestPractices(c *checker.CheckRequest) checker.CheckResult {
	if c.RepoClient.IsLocal() {
		e := sce.WithMessage(sce.ErrScorecardInternal, "not supported for local repos")
		return checker.CreateRuntimeErrorResult(CheckFuzzing, e)
	}

	repoURI := fmt.Sprintf("https://%s", c.Repo.URI())
	url := fmt.Sprintf("https://bestpractices.coreinfrastructure.org/projects.json?url=%s", repoURI)
	req, err := http.NewRequestWithContext(c.Ctx, "GET", url, nil)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("http.NewRequestWithContext: %v", err))
		return checker.CreateRuntimeErrorResult(CheckCIIBestPractices, e)
	}

	httpClient := http.Client{
		Transport: &expBackoffTransport{
			numRetries: 3,
		},
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("http.NewRequestWithContext: %v", err))
		return checker.CreateRuntimeErrorResult(CheckCIIBestPractices, e)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("ioutil.ReadAll: %v", err))
		return checker.CreateRuntimeErrorResult(CheckCIIBestPractices, e)
	}

	parsedResponse := []response{}
	if err := json.Unmarshal(b, &parsedResponse); err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("json.Unmarshal on %s - %s: %v", resp.Status, parsedResponse, err))
		return checker.CreateRuntimeErrorResult(CheckCIIBestPractices, e)
	}

	if len(parsedResponse) < 1 {
		return checker.CreateMinScoreResult(CheckCIIBestPractices, "no badge found")
	}

	result := parsedResponse[0]

	if result.BadgeLevel != "" {
		// Three levels: passing, silver and gold,
		// https://bestpractices.coreinfrastructure.org/en/criteria.
		const silverScore = 7
		const passingScore = 5
		const inProgressScore = 2
		switch {
		default:
			e := sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("unsupported badge: %v", result.BadgeLevel))
			return checker.CreateRuntimeErrorResult(CheckCIIBestPractices, e)
		case strings.Contains(result.BadgeLevel, "in_progress"):
			return checker.CreateResultWithScore(CheckCIIBestPractices, "badge detected: in_progress", inProgressScore)
		case strings.Contains(result.BadgeLevel, "silver"):
			return checker.CreateResultWithScore(CheckCIIBestPractices, "badge detected: silver", silverScore)
		case strings.Contains(result.BadgeLevel, "gold"):
			return checker.CreateMaxScoreResult(CheckCIIBestPractices, "badge detected: gold")
		case strings.Contains(result.BadgeLevel, "passing"):
			return checker.CreateResultWithScore(CheckCIIBestPractices, "badge detected: passing", passingScore)
		}
	}

	return checker.CreateMinScoreResult(CheckCIIBestPractices, "no badge detected")
}
