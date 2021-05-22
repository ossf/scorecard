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

	"github.com/ossf/scorecard/checker"
)

const ciiBestPracticesStr = "CII-Best-Practices"

//nolint:gochecknoinits
func init() {
	registerCheck(ciiBestPracticesStr, CIIBestPractices)
}

type response struct {
	BadgeLevel string `json:"badge_level"`
}

func CIIBestPractices(c *checker.CheckRequest) checker.CheckResult {
	repoURL := fmt.Sprintf("https://github.com/%s/%s", c.Owner, c.Repo)
	url := fmt.Sprintf("https://bestpractices.coreinfrastructure.org/projects.json?url=%s", repoURL)
	req, err := http.NewRequestWithContext(c.Ctx, "GET", url, nil)
	if err != nil {
		return checker.MakeRetryResult(ciiBestPracticesStr, err)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return checker.MakeRetryResult(ciiBestPracticesStr, err)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return checker.MakeRetryResult(ciiBestPracticesStr, err)
	}

	parsedResponse := []response{}
	if err := json.Unmarshal(b, &parsedResponse); err != nil {
		return checker.MakeRetryResult(ciiBestPracticesStr, err)
	}

	if len(parsedResponse) < 1 {
		c.Logf("no badge found")
		return checker.CheckResult{
			Name:       ciiBestPracticesStr,
			Pass:       false,
			Confidence: checker.MaxResultConfidence,
		}
	}

	result := parsedResponse[0]
	c.Logf("badge level: %s", result.BadgeLevel)

	if result.BadgeLevel != "" {
		return checker.CheckResult{
			Name:       ciiBestPracticesStr,
			Pass:       true,
			Confidence: checker.MaxResultConfidence,
		}
	}

	return checker.CheckResult{
		Name:       ciiBestPracticesStr,
		Pass:       false,
		Confidence: checker.MaxResultConfidence,
	}
}
