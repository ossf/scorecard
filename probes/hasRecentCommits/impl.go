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

//nolint:stylecheck
package hasRecentCommits

import (
	"embed"
	"fmt"
	"time"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/internal/utils/uerror"
)

//go:embed *.yml
var fs embed.FS

const (
	lookBackDays  = 90
	CommitsValue  = "commitsWithinThreshold"
	LookBackValue = "lookBackDays"
)

const Probe = "hasRecentCommits"

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	var findings []finding.Finding

	r := raw.MaintainedResults
	threshold := time.Now().AddDate(0 /*years*/, 0 /*months*/, -1*lookBackDays /*days*/)
	commitsWithinThreshold := 0

	for i := range r.DefaultBranchCommits {
		commit := r.DefaultBranchCommits[i]
		if commit.CommittedDate.After(threshold) {
			commitsWithinThreshold++
		}
	}

	if commitsWithinThreshold > 0 {
		f, err := finding.NewWith(fs, Probe,
			"Found a contribution within the threshold.", nil,
			finding.OutcomePositive)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		f = f.WithValues(map[string]int{
			CommitsValue:  commitsWithinThreshold,
			LookBackValue: lookBackDays,
		})
		findings = append(findings, *f)
	} else {
		f, err := finding.NewWith(fs, Probe,
			"Did not find contribution within the threshold.", nil,
			finding.OutcomeNegative)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	}

	return findings, Probe, nil
}
