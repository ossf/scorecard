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

package pkg

import (
	"encoding/json"
	"fmt"
	"io"

	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
)

// JSONScorecardProbeResult exports results as JSON for flat findings without checks.
//
//nolint:govet
type JSONScorecardProbeResult struct {
	Date      string            `json:"date"`
	Repo      jsonRepoV2        `json:"repo"`
	Scorecard jsonScorecardV2   `json:"scorecard"`
	Findings  []finding.Finding `json:"findings"`
	Metadata  map[string]string `json:"metadata"`
}

// TODO: finsinds should enventually be part of the scorecard structure.
func (r *ScorecardResult) AsPJSON(writer io.Writer,
) error {
	encoder := json.NewEncoder(writer)
	out := JSONScorecardProbeResult{
		Repo: jsonRepoV2{
			Name:   r.Repo.Name,
			Commit: r.Repo.CommitSHA,
		},
		Scorecard: jsonScorecardV2{
			Version: r.Scorecard.Version,
			Commit:  r.Scorecard.CommitSHA,
		},
		Date:     r.Date.Format("2006-01-02"),
		Findings: r.ProbeResults,
	}

	if err := encoder.Encode(out); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("encoder.Encode: %v", err))
	}

	return nil
}
