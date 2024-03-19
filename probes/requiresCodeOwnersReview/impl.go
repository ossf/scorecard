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
package requiresCodeOwnersReview

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/internal/probes"
	"github.com/ossf/scorecard/v4/probes/internal/utils/uerror"
)

func init() {
	probes.MustRegister(Probe, Run, []probes.CheckName{probes.BranchProtection})
}

//go:embed *.yml
var fs embed.FS

const (
	Probe         = "requiresCodeOwnersReview"
	BranchNameKey = "branchName"
)

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	r := raw.BranchProtectionResults
	var findings []finding.Finding

	if len(r.Branches) == 0 {
		f, err := finding.NewWith(fs, Probe, "no branches found", nil, finding.OutcomeNotApplicable)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
		return findings, Probe, nil
	}

	for i := range r.Branches {
		branch := &r.Branches[i]

		reqOwnerReviews := branch.BranchProtectionRule.RequiredPullRequestReviews.RequireCodeOwnerReviews
		var text string
		var outcome finding.Outcome

		switch {
		case reqOwnerReviews == nil:
			text = "could not determine whether codeowners review is allowed"
			outcome = finding.OutcomeNotAvailable
		case !*reqOwnerReviews:
			text = fmt.Sprintf("codeowners review is not required on branch '%s'", *branch.Name)
			outcome = finding.OutcomeNegative
		case len(r.CodeownersFiles) == 0:
			text = "codeowners review is required - but no codeowners file found in repo"
			outcome = finding.OutcomeNegative
		default:
			text = fmt.Sprintf("codeowner review is required on branch '%s'", *branch.Name)
			outcome = finding.OutcomePositive
		}
		f, err := finding.NewWith(fs, Probe, text, nil, outcome)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		f = f.WithValue(BranchNameKey, *branch.Name)
		findings = append(findings, *f)
	}
	return findings, Probe, nil
}
