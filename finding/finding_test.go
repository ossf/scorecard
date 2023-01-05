// Copyright 2021 OpenSSF Scorecard Authors
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

package finding

import (
	"embed"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v4/rule"
)

func errCmp(e1, e2 error) bool {
	return errors.Is(e1, e2) || errors.Is(e2, e1)
}

//go:embed testdata/*
var testfs embed.FS

func Test_FindingNew(t *testing.T) {
	snippet := "some code snippet"
	patch := "some patch values"
	sline := uint(10)
	eline := uint(46)
	t.Parallel()
	tests := []struct {
		name    string
		id      string
		err     error
		env     map[string]string
		finding *Finding
	}{
		{
			name: "risk high",
			id:   "testdata/risk-high",
			finding: &Finding{
				RuleName: "testdata/risk-high",
				Outcome:  OutcomeNegative,
				Risk:     rule.RiskHigh,
				Remediation: &rule.Remediation{
					Text:     "step1\nstep2 https://www.google.com/something",
					Markdown: "step1\nstep2 [google.com](https://www.google.com/something)",
					Effort:   rule.RemediationEffortLow,
				},
			},
		},
		{
			name: "effort low",
			id:   "testdata/effort-low",
			finding: &Finding{
				RuleName: "testdata/effort-low",
				Outcome:  OutcomeNegative,
				Risk:     rule.RiskHigh,
				Remediation: &rule.Remediation{
					Text:     "step1\nstep2 https://www.google.com/something",
					Markdown: "step1\nstep2 [google.com](https://www.google.com/something)",
					Effort:   rule.RemediationEffortLow,
				},
			},
		},
		{
			name: "effort high",
			id:   "testdata/effort-high",
			finding: &Finding{
				RuleName: "testdata/effort-high",
				Outcome:  OutcomeNegative,
				Risk:     rule.RiskHigh,
				Remediation: &rule.Remediation{
					Text:     "step1\nstep2 https://www.google.com/something",
					Markdown: "step1\nstep2 [google.com](https://www.google.com/something)",
					Effort:   rule.RemediationEffortHigh,
				},
			},
		},
		{
			name: "env variables",
			id:   "testdata/env-variables",
			env:  map[string]string{"branch": "master", "repo": "ossf/scorecard"},
			finding: &Finding{
				RuleName: "testdata/env-variables",
				Outcome:  OutcomeNegative,
				Risk:     rule.RiskHigh,
				Remediation: &rule.Remediation{
					Text:     "step1\nstep2 google.com/ossf/scorecard@master",
					Markdown: "step1\nstep2 [google.com/ossf/scorecard@master](google.com/ossf/scorecard@master)",
					Effort:   rule.RemediationEffortLow,
				},
			},
		},
		{
			name: "patch",
			id:   "testdata/env-variables",
			env:  map[string]string{"branch": "master", "repo": "ossf/scorecard"},
			finding: &Finding{
				RuleName: "testdata/env-variables",
				Outcome:  OutcomeNegative,
				Risk:     rule.RiskHigh,
				Remediation: &rule.Remediation{
					Text:     "step1\nstep2 google.com/ossf/scorecard@master",
					Markdown: "step1\nstep2 [google.com/ossf/scorecard@master](google.com/ossf/scorecard@master)",
					Effort:   rule.RemediationEffortLow,
					Patch:    &patch,
				},
			},
		},
		{
			name: "location",
			id:   "testdata/env-variables",
			env:  map[string]string{"branch": "master", "repo": "ossf/scorecard"},
			finding: &Finding{
				RuleName: "testdata/env-variables",
				Outcome:  OutcomeNegative,
				Risk:     rule.RiskHigh,
				Remediation: &rule.Remediation{
					Text:     "step1\nstep2 google.com/ossf/scorecard@master",
					Markdown: "step1\nstep2 [google.com/ossf/scorecard@master](google.com/ossf/scorecard@master)",
					Effort:   rule.RemediationEffortLow,
				},
				Location: &Location{
					Type:      FileTypeSource,
					Value:     "path/to/file.txt",
					LineStart: &sline,
					LineEnd:   &eline,
					Snippet:   &snippet,
				},
			},
		},
		{
			name: "text",
			id:   "testdata/env-variables",
			env:  map[string]string{"branch": "master", "repo": "ossf/scorecard"},
			finding: &Finding{
				RuleName: "testdata/env-variables",
				Outcome:  OutcomeNegative,
				Risk:     rule.RiskHigh,
				Remediation: &rule.Remediation{
					Text:     "step1\nstep2 google.com/ossf/scorecard@master",
					Markdown: "step1\nstep2 [google.com/ossf/scorecard@master](google.com/ossf/scorecard@master)",
					Effort:   rule.RemediationEffortLow,
				},
				Message: "some text",
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, err := FindingNew(testfs, tt.id)
			if err != nil || tt.err != nil {
				if !errCmp(err, tt.err) {
					t.Fatalf("unexpected error: %v", cmp.Diff(err, tt.err, cmpopts.EquateErrors()))
				}
				return
			}
			r = r.WithMessage(tt.finding.Message)

			if len(tt.env) > 1 {
				r = r.WithRemediationMetadata(tt.env)
			}

			if tt.finding.Remediation != nil && tt.finding.Remediation.Patch != nil {
				r = r.WithPatch(*tt.finding.Remediation.Patch)
			}

			if tt.finding.Location != nil {
				r = r.WithLocation(*tt.finding.Location)
			}

			if diff := cmp.Diff(*tt.finding, *r); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
