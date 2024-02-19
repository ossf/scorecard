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
package runsStatusChecksBeforeMerging

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/internal/utils/uerror"
)

//go:embed *.yml
var fs embed.FS

const (
	Probe         = "runsStatusChecksBeforeMerging"
	BranchNameKey = "branchName"
)

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	r := raw.BranchProtectionResults
	var findings []finding.Finding

	for i := range r.Branches {
		branch := &r.Branches[i]
		switch {
		case len(branch.BranchProtectionRule.CheckRules.Contexts) > 0:
			f, err := finding.NewWith(fs, Probe,
				fmt.Sprintf("status check found to merge onto on branch '%s'", *branch.Name), nil,
				finding.OutcomePositive)
			if err != nil {
				return nil, Probe, fmt.Errorf("create finding: %w", err)
			}
			f = f.WithValue(BranchNameKey, *branch.Name)
			findings = append(findings, *f)
		default:
			f, err := finding.NewWith(fs, Probe,
				fmt.Sprintf("no status checks found to merge onto branch '%s'", *branch.Name), nil,
				finding.OutcomeNegative)
			if err != nil {
				return nil, Probe, fmt.Errorf("create finding: %w", err)
			}
			f = f.WithValue(BranchNameKey, *branch.Name)
			findings = append(findings, *f)
		}
	}
	return findings, Probe, nil
}
