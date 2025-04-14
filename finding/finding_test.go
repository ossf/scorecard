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

package finding

import (
	"errors"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gopkg.in/yaml.v3"
)

func errCmp(e1, e2 error) bool {
	return errors.Is(e1, e2) || errors.Is(e2, e1)
}

func Test_FromBytes(t *testing.T) {
	snippet := "some code snippet"
	patch := "some patch values"
	sline := uint(10)
	eline := uint(46)
	trueOutcome := OutcomeTrue
	falseOutcome := OutcomeFalse
	t.Parallel()
	tests := []struct {
		err      error
		outcome  *Outcome
		metadata map[string]string
		finding  *Finding
		name     string
		id       string
		path     string
	}{
		{
			name:    "effort low",
			id:      "effort-low",
			path:    "testdata/effort-low.yml",
			outcome: &falseOutcome,
			finding: &Finding{
				Probe:   "effort-low",
				Outcome: OutcomeFalse,
				Remediation: &Remediation{
					Text:     "step1\nstep2 https://www.google.com/something",
					Markdown: "step1\nstep2 [google.com](https://www.google.com/something)",
					Effort:   RemediationEffortLow,
				},
			},
		},
		{
			name:    "effort high",
			id:      "effort-high",
			path:    "testdata/effort-high.yml",
			outcome: &falseOutcome,
			finding: &Finding{
				Probe:   "effort-high",
				Outcome: OutcomeFalse,
				Remediation: &Remediation{
					Text:     "step1\nstep2 https://www.google.com/something",
					Markdown: "step1\nstep2 [google.com](https://www.google.com/something)",
					Effort:   RemediationEffortHigh,
				},
			},
		},
		{
			name:     "env variables",
			id:       "metadata-variables",
			path:     "testdata/metadata-variables.yml",
			outcome:  &falseOutcome,
			metadata: map[string]string{"branch": "master", "repo": "ossf/scorecard"},
			finding: &Finding{
				Probe:   "metadata-variables",
				Outcome: OutcomeFalse,
				Remediation: &Remediation{
					Text:     "step1\nstep2 google.com/ossf/scorecard@master",
					Markdown: "step1\nstep2 [google.com/ossf/scorecard@master](google.com/ossf/scorecard@master)",
					Effort:   RemediationEffortLow,
				},
			},
		},
		{
			name:     "patch",
			id:       "metadata-variables",
			path:     "testdata/metadata-variables.yml",
			outcome:  &falseOutcome,
			metadata: map[string]string{"branch": "master", "repo": "ossf/scorecard"},
			finding: &Finding{
				Probe:   "metadata-variables",
				Outcome: OutcomeFalse,
				Remediation: &Remediation{
					Text:     "step1\nstep2 google.com/ossf/scorecard@master",
					Markdown: "step1\nstep2 [google.com/ossf/scorecard@master](google.com/ossf/scorecard@master)",
					Effort:   RemediationEffortLow,
					Patch:    &patch,
				},
			},
		},
		{
			name:     "location",
			id:       "metadata-variables",
			path:     "testdata/metadata-variables.yml",
			outcome:  &falseOutcome,
			metadata: map[string]string{"branch": "master", "repo": "ossf/scorecard"},
			finding: &Finding{
				Probe:   "metadata-variables",
				Outcome: OutcomeFalse,
				Remediation: &Remediation{
					Text:     "step1\nstep2 google.com/ossf/scorecard@master",
					Markdown: "step1\nstep2 [google.com/ossf/scorecard@master](google.com/ossf/scorecard@master)",
					Effort:   RemediationEffortLow,
				},
				Location: &Location{
					Type:      FileTypeSource,
					Path:      "path/to/file.txt",
					LineStart: &sline,
					LineEnd:   &eline,
					Snippet:   &snippet,
				},
			},
		},
		{
			name:     "text",
			id:       "metadata-variables",
			path:     "testdata/metadata-variables.yml",
			outcome:  &falseOutcome,
			metadata: map[string]string{"branch": "master", "repo": "ossf/scorecard"},
			finding: &Finding{
				Probe:   "metadata-variables",
				Outcome: OutcomeFalse,
				Remediation: &Remediation{
					Text:     "step1\nstep2 google.com/ossf/scorecard@master",
					Markdown: "step1\nstep2 [google.com/ossf/scorecard@master](google.com/ossf/scorecard@master)",
					Effort:   RemediationEffortLow,
				},
				Message: "some text",
			},
		},
		{
			name:    "true outcome",
			id:      "metadata-variables",
			path:    "testdata/metadata-variables.yml",
			outcome: &trueOutcome,
			finding: &Finding{
				Probe:   "metadata-variables",
				Outcome: OutcomeTrue,
				Message: "some text",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			content, err := os.ReadFile(tt.path)
			if err != nil {
				t.Fatal(err.Error())
			}

			r, err := FromBytes(content, tt.id)
			if err != nil || tt.err != nil {
				if !errCmp(err, tt.err) {
					t.Fatalf("unexpected error: %v", cmp.Diff(err, tt.err, cmpopts.EquateErrors()))
				}
				return
			}
			r = r.WithMessage(tt.finding.Message).WithLocation(tt.finding.Location)

			if len(tt.metadata) > 1 {
				r = r.WithRemediationMetadata(tt.metadata)
			}

			if tt.finding.Remediation != nil {
				r = r.WithPatch(tt.finding.Remediation.Patch)
			}

			if tt.outcome != nil {
				r = r.WithOutcome(*tt.outcome)
			}
			if diff := cmp.Diff(*tt.finding, *r, cmpopts.IgnoreUnexported(Finding{})); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestOutcome_UnmarshalYAML(t *testing.T) {
	t.Parallel()
	type args struct {
		n *yaml.Node
	}
	tests := []struct {
		args        args
		name        string
		wantOutcome Outcome
		wantErr     bool
	}{
		{
			name:        "true outcome",
			wantOutcome: OutcomeTrue,
			args: args{
				n: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "True",
				},
			},
			wantErr: false,
		},
		{
			name:        "false outcome",
			wantOutcome: OutcomeFalse,
			args: args{
				n: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "False",
				},
			},
			wantErr: false,
		},
		{
			name:        "NotAvailable outcome",
			wantOutcome: OutcomeNotAvailable,
			args: args{
				n: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "NotAvailable",
				},
			},
			wantErr: false,
		},
		{
			name:        "NotSupported outcome",
			wantOutcome: OutcomeNotSupported,
			args: args{
				n: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "NotSupported",
				},
			},
			wantErr: false,
		},
		{
			name:        "Unknown error",
			wantOutcome: OutcomeError,
			args: args{
				n: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "Error",
				},
			},
			wantErr: false,
		},
		{
			name: "Unknown outcome",
			args: args{
				n: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "Unknown",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var v Outcome
			if err := v.UnmarshalYAML(tt.args.n); (err != nil) != tt.wantErr {
				t.Errorf("Outcome.UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.wantOutcome, v); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
