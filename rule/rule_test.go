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

package rule

import (
	"embed"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gopkg.in/yaml.v3"
)

func errCmp(e1, e2 error) bool {
	return errors.Is(e1, e2) || errors.Is(e2, e1)
}

//go:embed testdata/*
var testfs embed.FS

func Test_New(t *testing.T) {
	t.Parallel()
	// nolint: govet
	tests := []struct {
		name string
		id   string
		err  error
		rule *Rule
	}{
		{
			name: "all fields set",
			id:   "testdata/all-fields",
			rule: &Rule{
				Name:       "testdata/all-fields",
				Short:      "short description",
				Desc:       "description",
				Motivation: "line1 line2\n",
				Risk:       RiskHigh,
				Remediation: &Remediation{
					Text:     "step1\nstep2 https://www.google.com/something",
					Markdown: "step1\nstep2 [google.com](https://www.google.com/something)",
					Effort:   RemediationEffortLow,
				},
			},
		},
		{
			name: "invalid risk",
			id:   "testdata/invalid-risk",
			err:  errInvalid,
		},
		{
			name: "invalid effort",
			id:   "testdata/invalid-effort",
			err:  errInvalid,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, err := New(testfs, tt.id)
			if err != nil || tt.err != nil {
				if !errCmp(err, tt.err) {
					t.Fatalf("unexpected error: %v", cmp.Diff(err, tt.err, cmpopts.EquateErrors()))
				}
				return
			}

			if diff := cmp.Diff(*tt.rule, *r); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRisk_GreaterThan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		r    Risk
		rr   Risk
		want bool
	}{
		{
			name: "greater than",
			r:    RiskHigh,
			rr:   RiskLow,
			want: true,
		},
		{
			name: "less than",
			r:    RiskLow,
			rr:   RiskHigh,
			want: false,
		},
		{
			name: "equal",
			r:    RiskMedium,
			rr:   RiskMedium,
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.GreaterThan(tt.rr); got != tt.want {
				t.Errorf("Risk.GreaterThan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRisk_String(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet
		name string
		r    Risk
		want string
	}{
		{
			name: "RiskNone",
			r:    RiskNone,
			want: "None",
		},
		{
			name: "RiskLow",
			r:    RiskLow,
			want: "Low",
		},
		{
			name: "RiskMedium",
			r:    RiskMedium,
			want: "Medium",
		},
		{
			name: "RiskHigh",
			r:    RiskHigh,
			want: "High",
		},
		{
			name: "RiskCritical",
			r:    RiskCritical,
			want: "Critical",
		},
		{
			name: "invalid",
			r:    Risk(100),
			want: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.String(); got != tt.want {
				t.Errorf("Risk.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemediationEffort_String(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet
		name   string
		effort RemediationEffort
		want   string
	}{
		{
			name:   "RemediationEffortNone",
			effort: RemediationEffortNone,
			want:   "",
		},
		{
			name:   "RemediationEffortLow",
			effort: RemediationEffortLow,
			want:   "Low",
		},
		{
			name:   "RemediationEffortMedium",
			effort: RemediationEffortMedium,
			want:   "Medium",
		},
		{
			name:   "RemediationEffortHigh",
			effort: RemediationEffortHigh,
			want:   "High",
		},
		{
			name:   "invalid",
			effort: RemediationEffort(100),
			want:   "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.effort.String(); got != tt.want {
				t.Errorf("RemediationEffort.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRisk_UnmarshalYAML(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet
		name    string
		input   string
		wantErr error
		want    Risk
	}{
		{
			name:  "RiskNone",
			input: "None",
			want:  RiskNone,
		},
		{
			name:  "RiskLow",
			input: "Low",
			want:  RiskLow,
		},
		{
			name:  "RiskMedium",
			input: "Medium",
			want:  RiskMedium,
		},
		{
			name:  "RiskHigh",
			input: "High",
			want:  RiskHigh,
		},
		{
			name:  "RiskCritical",
			input: "Critical",
			want:  RiskCritical,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var r Risk
			err := yaml.Unmarshal([]byte(tt.input), &r)
			if err != nil {
				if tt.wantErr == nil || !errors.Is(err, tt.wantErr) {
					t.Errorf("Risk.UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if r != tt.want {
				t.Errorf("Risk.UnmarshalYAML() got = %v, want %v", r, tt.want)
			}
		})
	}
}

func TestRemediationEffort_UnmarshalYAML(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet
		name    string
		input   string
		wantErr error
		want    RemediationEffort
	}{
		{
			name:  "RemediationEffortLow",
			input: "Low",
			want:  RemediationEffortLow,
		},
		{
			name:  "RemediationEffortMedium",
			input: "Medium",
			want:  RemediationEffortMedium,
		},
		{
			name:  "RemediationEffortHigh",
			input: "High",
			want:  RemediationEffortHigh,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var r RemediationEffort
			err := yaml.Unmarshal([]byte(tt.input), &r)
			if err != nil {
				if tt.wantErr == nil || !errors.Is(err, tt.wantErr) {
					t.Errorf("RemediationEffort.UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if r != tt.want {
				t.Errorf("RemediationEffort.UnmarshalYAML() got = %v, want %v", r, tt.want)
			}
		})
	}
}

func Test_validate(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet
		name    string
		rule    *jsonRule
		wantErr error
	}{
		{
			name: "valid",
			rule: &jsonRule{
				Risk: RiskLow,
				Remediation: jsonRemediation{
					Effort: RemediationEffortHigh,
				},
			},
			wantErr: nil,
		},
		{
			name: "invalid risk",
			rule: &jsonRule{
				Risk: Risk(100),
				Remediation: jsonRemediation{
					Effort: RemediationEffortHigh,
				},
			},
			wantErr: fmt.Errorf("%w: invalid: risk '100'", errInvalid),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.rule)
			if err != nil {
				if tt.wantErr == nil || !cmp.Equal(tt.wantErr.Error(), err.Error()) {
					t.Logf("got: %s", err.Error())
					t.Errorf("validate() error = %v, wantErr %v", err, cmp.Diff(tt.wantErr.Error(), err.Error()))
				}
				return
			}
			if tt.wantErr != nil {
				t.Errorf("validate() error = %v, wantErr %v", err, cmp.Diff(tt.wantErr, err))
			}
		})
	}
}
