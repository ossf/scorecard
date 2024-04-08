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
package sastToolRunsOnAllCommits

import (
	"embed"
	"fmt"
	"strconv"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/internal/probes"
	"github.com/ossf/scorecard/v4/probes/internal/utils/uerror"
)

func init() {
	probes.MustRegister(Probe, Run, []probes.CheckName{probes.SAST})
}

//go:embed *.yml
var fs embed.FS

const (
	Probe = "sastToolRunsOnAllCommits"
	// TotalPRsKey is the Values map key which specifies the total number of PRs being evaluated.
	TotalPRsKey = "totalPullRequestsMerged"
	// AnalyzedPRsKey is the Values map key which specifies the number of PRs analyzed by a SAST.
	AnalyzedPRsKey = "totalPullRequestsAnalyzed"
)

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	r := raw.SASTResults

	f, err := finding.NewWith(fs, Probe,
		"", nil,
		finding.OutcomeTrue)
	if err != nil {
		return nil, Probe, fmt.Errorf("create finding: %w", err)
	}

	totalPullRequestsMerged := len(r.Commits)
	totalPullRequestsAnalyzed := 0

	for i := range r.Commits {
		wf := &r.Commits[i]
		if wf.Compliant {
			totalPullRequestsAnalyzed++
		}
	}

	if totalPullRequestsMerged == 0 {
		f = f.WithOutcome(finding.OutcomeNotApplicable)
		f = f.WithMessage("no pull requests merged into dev branch")
		return []finding.Finding{*f}, Probe, nil
	}

	f = f.WithValue(AnalyzedPRsKey, strconv.Itoa(totalPullRequestsAnalyzed))
	f = f.WithValue(TotalPRsKey, strconv.Itoa(totalPullRequestsMerged))

	if totalPullRequestsAnalyzed == totalPullRequestsMerged {
		msg := fmt.Sprintf("all commits (%v) are checked with a SAST tool", totalPullRequestsMerged)
		f = f.WithOutcome(finding.OutcomeTrue).WithMessage(msg)
	} else {
		msg := fmt.Sprintf("%v commits out of %v are checked with a SAST tool",
			totalPullRequestsAnalyzed, totalPullRequestsMerged)
		f = f.WithOutcome(finding.OutcomeFalse).WithMessage(msg)
	}
	return []finding.Finding{*f}, Probe, nil
}
