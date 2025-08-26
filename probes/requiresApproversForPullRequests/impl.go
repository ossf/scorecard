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
package requiresApproversForPullRequests

import (
	"embed"
	"errors"
	"fmt"
	"strconv"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/checknames"
	"github.com/ossf/scorecard/v5/internal/probes"
	"github.com/ossf/scorecard/v5/probes/internal/utils/uerror"
)

func init() {
	probes.MustRegister(Probe, Run, []checknames.CheckName{checknames.BranchProtection})
}

//go:embed *.yml
var fs embed.FS

const (
	Probe                = "requiresApproversForPullRequests"
	BranchNameKey        = "branchName"
	RequiredReviewersKey = "numberOfRequiredReviewers"
)

var errWrongValue = errors.New("wrong value, should not happen")

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

		nilMsg := fmt.Sprintf("could not determine whether branch '%s' has required approving review count", *branch.Name)
		falseMsg := fmt.Sprintf("branch '%s' does not require approvers", *branch.Name)

		p := branch.ProtectionRule.PullRequestRule.RequiredApprovingReviewCount

		f, err := finding.NewWith(fs, Probe, "", nil, finding.OutcomeNotAvailable)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		f = f.WithValue(BranchNameKey, *branch.Name)
		switch {
		case p == nil:
			f = f.WithMessage(nilMsg).WithOutcome(finding.OutcomeNotAvailable)
		case *p > 0:
			msg := fmt.Sprintf("required approving review count is %d on branch '%s'", *p, *branch.Name)
			f = f.WithMessage(msg).WithOutcome(finding.OutcomeTrue)
			f = f.WithValue(RequiredReviewersKey, strconv.Itoa(int(*p)))
		case *p == 0:
			f = f.WithMessage(falseMsg).WithOutcome(finding.OutcomeFalse)
			f = f.WithValue(RequiredReviewersKey, strconv.Itoa(int(*p)))
		default:
			return nil, Probe, fmt.Errorf("create finding: %w", errWrongValue)
		}
		findings = append(findings, *f)
	}
	return findings, Probe, nil
}
