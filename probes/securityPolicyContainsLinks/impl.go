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

// nolint:stylecheck
package securityPolicyContainsLinks

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/internal/utils"
)

//go:embed *.yml
var fs embed.FS

var probe = "securityPolicyContainsLinks"

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	var findings []finding.Finding
	policies := raw.SecurityPolicyResults.PolicyFiles
	for i := range policies {
		policy := policies[i]
		emails := utils.CountSecInfo(policy.Information, checker.SecurityPolicyInformationTypeEmail, true)
		urls := utils.CountSecInfo(policy.Information, checker.SecurityPolicyInformationTypeLink, true)

		if (urls + emails) > 0 {
			f, err := finding.NewPositive(fs, probe,
				"URL or email contacts found", policy.File.Location())
			if err != nil {
				return nil, probe, fmt.Errorf("create finding: %w", err)
			}
			findings = append(findings, *f)
		} else {
			f, err := finding.NewNegative(fs, probe,
				"no URL or email contacts found", nil)
			if err != nil {
				return nil, probe, fmt.Errorf("create finding: %w", err)
			}
			findings = append(findings, *f)
		}
	}

	if len(findings) == 0 {
		f, err := finding.NewNegative(fs, probe, "no security file to analyze", nil)
		if err != nil {
			return nil, probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	}
	return findings, probe, nil
}
