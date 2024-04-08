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
package requiresPRsToChangeCode

import (
	"embed"
	"errors"
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
	Probe         = "requiresPRsToChangeCode"
	BranchNameKey = "branchName"
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

		nilMsg := fmt.Sprintf("could not determine whether branch '%s' requires PRs to change code", *branch.Name)
		trueMsg := fmt.Sprintf("PRs are required in order to make changes on branch '%s'", *branch.Name)
		falseMsg := fmt.Sprintf("PRs are not required to make changes on branch '%s'; ", *branch.Name) +
			"or we don't have data to detect it." +
			"If you think it might be the latter, make sure to run Scorecard with a PAT or use Repo " +
			"Rules (that are always public) instead of Branch Protection settings"

		p := branch.BranchProtectionRule.PullRequestRule.Required

		f, err := finding.NewWith(fs, Probe, "", nil, finding.OutcomeNotAvailable)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}

		switch {
		case p == nil:
			f = f.WithMessage(nilMsg).WithOutcome(finding.OutcomeNotAvailable)
		case *p:
			f = f.WithMessage(trueMsg).WithOutcome(finding.OutcomeTrue)
		case !*p:
			f = f.WithMessage(falseMsg).WithOutcome(finding.OutcomeFalse)
		default:
			return nil, Probe, fmt.Errorf("create finding: %w", errWrongValue)
		}
		f = f.WithValue(BranchNameKey, *branch.Name)
		findings = append(findings, *f)
	}
	return findings, Probe, nil
}
