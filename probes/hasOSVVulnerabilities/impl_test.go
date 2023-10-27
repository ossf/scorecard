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
package hasOSVVulnerabilities

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/finding/probe"
)

func Test_Run(t *testing.T) {
	t.Parallel()
	// nolint:govet
	tests := []struct {
		name            string
		raw             *checker.RawResults
		outcomes        []finding.Outcome
		expectedFinding *finding.Finding
		err             error
	}{
		{
			name: "vulnerabilities present",
			raw: &checker.RawResults{
				VulnerabilitiesResults: checker.VulnerabilitiesData{
					Vulnerabilities: []clients.Vulnerability{
						{ID: "foo"},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeNegative,
			},
		},
		{
			name: "vulnerabilities not present",
			raw: &checker.RawResults{
				VulnerabilitiesResults: checker.VulnerabilitiesData{
					Vulnerabilities: []clients.Vulnerability{},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomePositive,
			},
		},
		{
			name: "vulnerabilities not present",
			raw: &checker.RawResults{
				VulnerabilitiesResults: checker.VulnerabilitiesData{
					Vulnerabilities: []clients.Vulnerability{},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomePositive,
			},
		},
		{
			name: "vulnerabilities and metadata present. 'foo' must appear in the findings remediation text.",
			raw: &checker.RawResults{
				VulnerabilitiesResults: checker.VulnerabilitiesData{
					Vulnerabilities: []clients.Vulnerability{
						{ID: "foo"},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeNegative,
			},
			expectedFinding: &finding.Finding{
				Probe:   "hasOSVVulnerabilities",
				Message: "Project is vulnerable to: foo",
				Remediation: &probe.Remediation{
					//nolint
					Text: `Fix the foo by following information from https://osv.dev/foo.
If the vulnerability is in a dependency, update the dependency to a non-vulnerable version. If no update is available, consider whether to remove the dependency.
If you believe the vulnerability does not affect your project, the vulnerability can be ignored. To ignore, create an osv-scanner.toml file next to the dependency manifest (e.g. package-lock.json) and specify the ID to ignore and reason. Details on the structure of osv-scanner.toml can be found on OSV-Scanner repository.`,
					//nolint
					Markdown: `Fix the foo by following information from [OSV](https://osv.dev/foo).
If the vulnerability is in a dependency, update the dependency to a non-vulnerable version. If no update is available, consider whether to remove the dependency.
If you believe the vulnerability does not affect your project, the vulnerability can be ignored. To ignore, create an osv-scanner.toml ([example](https://github.com/google/osv.dev/blob/eb99b02ec8895fe5b87d1e76675ddad79a15f817/vulnfeeds/osv-scanner.toml)) file next to the dependency manifest (e.g. package-lock.json) and specify the ID to ignore and reason. Details on the structure of osv-scanner.toml can be found on [OSV-Scanner repository](https://github.com/google/osv-scanner#ignore-vulnerabilities-by-id).`,
					Effort: 3,
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			findings, s, err := Run(tt.raw)
			if !cmp.Equal(tt.err, err, cmpopts.EquateErrors()) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(tt.err, err, cmpopts.EquateErrors()))
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(Probe, s); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(len(tt.outcomes), len(findings)); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
			for i := range tt.outcomes {
				outcome := &tt.outcomes[i]
				f := &findings[i]
				if diff := cmp.Diff(*outcome, f.Outcome); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
				if tt.expectedFinding != nil {
					f := &findings[i]
					if diff := cmp.Diff(tt.expectedFinding, f); diff != "" {
						t.Errorf("mismatch (-want +got):\n%s", diff)
					}
				}
			}
		})
	}
}
