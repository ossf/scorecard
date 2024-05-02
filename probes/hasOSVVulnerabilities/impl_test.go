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
package hasOSVVulnerabilities

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/finding"
)

func Test_Run(t *testing.T) {
	t.Parallel()
	//nolint:govet
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
				finding.OutcomeTrue,
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
				finding.OutcomeFalse,
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

func TestRun_remediation(t *testing.T) {
	t.Parallel()
	//nolint:govet
	tests := []struct {
		name              string
		raw               *checker.RawResults
		wantRemediation   bool
		remediationSubstr string
		err               error
	}{
		{
			name: "no remediation needed",
			raw: &checker.RawResults{
				VulnerabilitiesResults: checker.VulnerabilitiesData{
					Vulnerabilities: []clients.Vulnerability{},
				},
			},
			wantRemediation: false,
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
			wantRemediation:   true,
			remediationSubstr: "Fix the foo",
		},
		{
			name: "only one vuln ID appears in the findings remediation text link.",
			raw: &checker.RawResults{
				VulnerabilitiesResults: checker.VulnerabilitiesData{
					Vulnerabilities: []clients.Vulnerability{
						{ID: "foo"},
						{
							ID:      "bar",
							Aliases: []string{"foo"},
						},
					},
				},
			},
			wantRemediation:   true,
			remediationSubstr: "https://osv.dev/bar .",
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			findings, _, err := Run(tt.raw)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for i := range findings {
				gotRemediation := findings[i].Remediation != nil
				if tt.wantRemediation != gotRemediation {
					t.Errorf("wanted remediation?: %t, had remediation?: %t", tt.wantRemediation, gotRemediation)
				}
				if tt.wantRemediation && !strings.Contains(findings[i].Remediation.Text, tt.remediationSubstr) {
					t.Errorf("wanted to find substr %q in remediation text, but didn't", tt.remediationSubstr)
				}
			}
		})
	}
}
