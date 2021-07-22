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
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ossf/scorecard/checker"
	sce "github.com/ossf/scorecard/errors"
)

// CheckCIIBestPractices is the registered name for CIIBestPractices.
const CheckCIIBestPractices = "CII-Best-Practices"

//nolint:gochecknoinits
func init() {
	registerCheck(CheckCIIBestPractices, CIIBestPractices)
}

type response struct {
	BadgeLevel string `json:"badge_level"`
}

func CIIBestPractices(c *checker.CheckRequest) checker.CheckResult {
	repoURL := fmt.Sprintf("https://github.com/%s/%s", c.Owner, c.Repo)
	url := fmt.Sprintf("https://bestpractices.coreinfrastructure.org/projects.json?url=%s", repoURL)
	req, err := http.NewRequestWithContext(c.Ctx, "GET", url, nil)
	if err != nil {
		e := sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("http.NewRequestWithContext: %v", err))
		return checker.CreateRuntimeErrorResult(CheckCIIBestPractices, e)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		e := sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("HTTPClient.Do: %v", err))
		return checker.CreateRuntimeErrorResult(CheckCIIBestPractices, e)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		e := sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("ioutil.ReadAll: %v", err))
		return checker.CreateRuntimeErrorResult(CheckCIIBestPractices, e)
	}

	parsedResponse := []response{}
	if err := json.Unmarshal(b, &parsedResponse); err != nil {
		e := sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("json.Unmarshal: %v", err))
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
		switch {
		default:
			e := sce.Create(sce.ErrScorecardInternal, "unsupported badge")
			return checker.CreateRuntimeErrorResult(CheckCIIBestPractices, e)
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
