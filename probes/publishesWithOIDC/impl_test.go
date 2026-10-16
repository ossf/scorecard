// Copyright 2026 OpenSSF Scorecard Authors
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

package publishesWithOIDC

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/internal/utils/test"
	"github.com/ossf/scorecard/v5/probes/internal/utils/uerror"
)

func Test_Run(t *testing.T) {
	t.Parallel()
	msg := "debug"
	//nolint:govet
	tests := []struct {
		name     string
		raw      *checker.RawResults
		outcomes []finding.Outcome
		wantErr  error
	}{
		{
			name:    "nil raw",
			wantErr: uerror.ErrNil,
		},
		{
			name: "oidc",
			raw: &checker.RawResults{
				PackagingResults: checker.PackagingData{
					Packages: []checker.Package{{
						PublishAuth:      checker.PublishAuthOIDC,
						PublishEcosystem: "pypi",
					}},
				},
			},
			outcomes: []finding.Outcome{finding.OutcomeTrue},
		},
		{
			name: "credential",
			raw: &checker.RawResults{
				PackagingResults: checker.PackagingData{
					Packages: []checker.Package{{
						PublishAuth:      checker.PublishAuthCredential,
						PublishEcosystem: "npm",
					}},
				},
			},
			outcomes: []finding.Outcome{finding.OutcomeFalse},
		},
		{
			name: "debug package is ignored",
			raw: &checker.RawResults{
				PackagingResults: checker.PackagingData{
					Packages: []checker.Package{{Msg: &msg}},
				},
			},
			outcomes: []finding.Outcome{finding.OutcomeNotApplicable},
		},
		{
			name: "unclassified packaging workflow",
			raw: &checker.RawResults{
				PackagingResults: checker.PackagingData{
					Packages: []checker.Package{{}},
				},
			},
			outcomes: []finding.Outcome{finding.OutcomeNotApplicable},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			findings, s, err := Run(tt.raw)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Run() error = %v, want %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(Probe, s); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
			test.AssertOutcomes(t, findings, tt.outcomes)
		})
	}
}
