// Copyright 2022 OpenSSF Scorecard Authors
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

package evaluation

import (
	"testing"

	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/hasOSVVulnerabilities"
	"github.com/ossf/scorecard/v5/probes/releasesDirectDepsAreVulnFree"
	scut "github.com/ossf/scorecard/v5/utests"
)

// TestVulnerabilities tests the vulnerabilities checker.
func TestVulnerabilities(t *testing.T) {
	t.Parallel()
	//nolint:govet
	tests := []struct {
		name     string
		findings []finding.Finding
		result   scut.TestReturn
		expected []struct {
			lineNumber uint
		}
	}{
		{
			name: "no vulnerabilities - current and releases clean",
			findings: []finding.Finding{
				{
					Probe:   "hasOSVVulnerabilities",
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   "releasesDirectDepsAreVulnFree",
					Outcome: finding.OutcomeTrue,
					Message: "release v1.0.0 has no known vulnerabilities",
				},
			},
			result: scut.TestReturn{
				Score: 10,
			},
		},
		{
			name:     "three current vulnerabilities, all releases clean",
			findings: append(vulnFindings(t, 3), cleanReleaseFindings(t, 5)...),
			result: scut.TestReturn{
				Score:        7, // 6 - 3*0.6 + 4 = 4.2 + 4 = 8.2 rounds to 8, but 3 vulns = 3 so 6-3=3, 3+4=7
				NumberOfWarn: 3,
			},
		},
		{
			name:     "twelve current vulnerabilities, all releases clean",
			findings: append(vulnFindings(t, 12), cleanReleaseFindings(t, 5)...),
			result: scut.TestReturn{
				Score:        4, // max penalty is 6 points, so 0 + 4 = 4
				NumberOfWarn: 12,
			},
		},
		{
			name: "no current vulnerabilities, but releases have issues",
			findings: append(
				[]finding.Finding{
					{
						Probe:   "hasOSVVulnerabilities",
						Outcome: finding.OutcomeFalse,
					},
				},
				vulnReleaseFindings(t, 2, 5)...,
			),
			result: scut.TestReturn{
				Score:        8, // 6 + (4 * 2/5) = 6 + 1.6 = 7.6 rounds to 8
				NumberOfWarn: 3, // 3 vulnerable releases
			},
		},
		{
			name: "both current and release vulnerabilities",
			findings: append(
				vulnFindings(t, 2),
				vulnReleaseFindings(t, 2, 4)...,
			),
			result: scut.TestReturn{
				Score:        6, // (6 - 2) + (4 * 2/4) = 4 + 2 = 6
				NumberOfWarn: 4,
			},
		},
		{
			name:     "invalid findings - missing current vuln probe",
			findings: cleanReleaseFindings(t, 3),
			result: scut.TestReturn{
				Score: -1,
				Error: sce.ErrScorecardInternal,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			got := Vulnerabilities(tt.name, tt.findings, &dl)
			scut.ValidateTestReturn(t, tt.name, &tt.result, &got, &dl)
		})
	}
}

// helper to generate repeated vuln findings.
func vulnFindings(t *testing.T, n int) []finding.Finding {
	t.Helper()
	findings := make([]finding.Finding, n)
	for i := range findings {
		findings[i] = finding.Finding{
			Probe:   hasOSVVulnerabilities.Probe,
			Outcome: finding.OutcomeTrue,
		}
	}
	return findings
}

// helper to generate clean release findings.
func cleanReleaseFindings(t *testing.T, n int) []finding.Finding {
	t.Helper()
	findings := make([]finding.Finding, n)
	for i := range findings {
		findings[i] = finding.Finding{
			Probe:   releasesDirectDepsAreVulnFree.Probe,
			Outcome: finding.OutcomeTrue,
			Message: "release is clean",
		}
	}
	return findings
}

// helper to generate vulnerable release findings.
// cleanCount = number of clean releases, totalCount = total releases.
func vulnReleaseFindings(t *testing.T, cleanCount, totalCount int) []finding.Finding {
	t.Helper()
	findings := make([]finding.Finding, totalCount)
	for i := 0; i < cleanCount; i++ {
		findings[i] = finding.Finding{
			Probe:   releasesDirectDepsAreVulnFree.Probe,
			Outcome: finding.OutcomeTrue,
			Message: "release is clean",
		}
	}
	for i := cleanCount; i < totalCount; i++ {
		findings[i] = finding.Finding{
			Probe:   releasesDirectDepsAreVulnFree.Probe,
			Outcome: finding.OutcomeFalse,
			Message: "release has vulnerabilities",
		}
	}
	return findings
}
