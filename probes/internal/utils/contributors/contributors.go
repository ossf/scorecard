// Copyright 2023 OpenSSF Scorecard Authors
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

package contributors

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
)

const (
	minContributionsPerUser    = 5
)

func Run(raw *checker.RawResults, fs embed.FS, probeID string) ([]finding.Finding, string, error) {
	var findings []finding.Finding

	users := raw.ContributorsResults.Users

	if len(users) == 0 {
		return findings, probeID, nil
	}

	entities := make(map[string]bool)

	for _, user := range users {
		if user.NumContributions < minContributionsPerUser {
			continue
		}

		for _, org := range user.Organizations {
			if _, ok := entities[org.Login]; !ok {
				entities[org.Login] = true
			}
		}

		for _, comp := range user.Companies {
			if _, ok := entities[comp]; !ok {
				entities[comp] = true
			}
		}
	}

	// Convert entities map to findings slice
	for e := range entities {
		f, err := finding.NewWith(fs, probeID,
			fmt.Sprintf("%s contributor org/company found", e), nil,
			finding.OutcomePositive)
		if err != nil {
			return nil, probeID, fmt.Errorf("create finding: %w", err)
		}
		
		findings = append(findings, *f)
	}

	return findings, probeID, nil
}
