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
package requiresUpToDateBranches

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/internal/utils/uerror"
)

//go:embed *.yml
var fs embed.FS

const Probe = "requiresUpToDateBranches"

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	r := raw.BranchProtectionResults
	var findings []finding.Finding

	for i := range r.Branches {
		branch := &r.Branches[i]
		//nolint:nestif
		if branch.BranchProtectionRule.CheckRules.UpToDateBeforeMerge == nil {
			f, err := finding.NewWith(fs, Probe,
				fmt.Sprintf("unable to retrieve whether up-to-date branches are needed to merge on branch '%s'", *branch.Name),
				nil, finding.OutcomeNotAvailable)
			if err != nil {
				return nil, Probe, fmt.Errorf("create finding: %w", err)
			}
			f = f.WithValues(map[string]int{
				*branch.Name: 1,
			})
			findings = append(findings, *f)
		} else {
			// Note: `This setting will not take effect unless at least one status check is enabled`.
			if *branch.BranchProtectionRule.CheckRules.UpToDateBeforeMerge {
				f, err := finding.NewWith(fs, Probe,
					fmt.Sprintf("status checks require up-to-date branches for '%s'", *branch.Name),
					nil, finding.OutcomePositive)
				if err != nil {
					return nil, Probe, fmt.Errorf("create finding: %w", err)
				}
				f = f.WithValues(map[string]int{
					*branch.Name: 1,
				})
				findings = append(findings, *f)
			} else {
				f, err := finding.NewWith(fs, Probe,
					fmt.Sprintf("status checks do not require up-to-date branches for '%s'", *branch.Name),
					nil, finding.OutcomeNegative)
				if err != nil {
					return nil, Probe, fmt.Errorf("create finding: %w", err)
				}
				f = f.WithValues(map[string]int{
					*branch.Name: 1,
				})
				findings = append(findings, *f)
			}
		}
	}
	return findings, Probe, nil
}
