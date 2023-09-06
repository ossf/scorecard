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
package securityPolicyContainsText

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/internal/utils/secpolicy"
	"github.com/ossf/scorecard/v4/probes/internal/utils/uerror"
)

//go:embed *.yml
var fs embed.FS

const Probe = "securityPolicyContainsText"

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}
	var findings []finding.Finding
	policies := raw.SecurityPolicyResults.PolicyFiles
	for i := range policies {
		policy := &policies[i]
		linkedContentLen := 0
		emails := secpolicy.CountSecInfo(policy.Information, checker.SecurityPolicyInformationTypeEmail, true)
		urls := secpolicy.CountSecInfo(policy.Information, checker.SecurityPolicyInformationTypeLink, true)
		for _, i := range secpolicy.FindSecInfo(policy.Information, checker.SecurityPolicyInformationTypeEmail, true) {
			linkedContentLen += len(i.InformationValue.Match)
		}
		for _, i := range secpolicy.FindSecInfo(policy.Information, checker.SecurityPolicyInformationTypeLink, true) {
			linkedContentLen += len(i.InformationValue.Match)
		}

		if policy.File.FileSize > 1 && (policy.File.FileSize > uint(linkedContentLen+((urls+emails)*2))) {
			f, err := finding.NewPositive(fs, Probe,
				"Found text in security policy", policy.File.Location())
			if err != nil {
				return nil, Probe, fmt.Errorf("create finding: %w", err)
			}
			findings = append(findings, *f)
		} else {
			f, err := finding.NewNegative(fs, Probe,
				"No text (besides links / emails) found in security policy", nil)
			if err != nil {
				return nil, Probe, fmt.Errorf("create finding: %w", err)
			}
			findings = append(findings, *f)
		}
	}

	if len(findings) == 0 {
		f, err := finding.NewNegative(fs, Probe, "no security file to analyze", nil)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	}
	return findings, Probe, nil
}
