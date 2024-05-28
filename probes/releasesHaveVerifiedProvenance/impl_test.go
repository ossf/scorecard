// Copyright 2024 OpenSSF Scorecard Authors
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
package releasesHaveVerifiedProvenance

import (
	"testing"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
)

func Test_Run(t *testing.T) {
	tests := []struct {
		desc     string
		pkgs     []checker.ProjectPackage
		outcomes []finding.Outcome
		err      error
	}{
		{
			desc:     "no packages found",
			outcomes: []finding.Outcome{finding.OutcomeNotAvailable},
		},
		{
			desc: "some releases with verified provenance",
			pkgs: []checker.ProjectPackage{
				{
					Name:       "a",
					Version:    "1.0.0",
					Provenance: checker.PackageProvenance{IsVerified: true},
				},
				{
					Name:    "a",
					Version: "1.0.1",
				},
			},
			outcomes: []finding.Outcome{finding.OutcomeTrue, finding.OutcomeFalse},
		},
		{
			desc: "all releases with verified provenance",
			pkgs: []checker.ProjectPackage{
				{
					Name:       "a",
					Version:    "1.0.0",
					Provenance: checker.PackageProvenance{IsVerified: true},
				},
				{
					Name:       "a",
					Version:    "1.0.1",
					Provenance: checker.PackageProvenance{IsVerified: true},
				},
			},
			outcomes: []finding.Outcome{finding.OutcomeTrue, finding.OutcomeTrue},
		},
		{
			desc: "no verified provenance",
			pkgs: []checker.ProjectPackage{
				{
					Name:    "a",
					Version: "1.0.0",
				},
				{
					Name:    "a",
					Version: "1.0.1",
				},
			},
			outcomes: []finding.Outcome{finding.OutcomeFalse, finding.OutcomeFalse},
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.desc, func(t *testing.T) {
			raw := checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Packages: tt.pkgs,
				},
			}

			outcomes, _, err := Run(&raw)

			if tt.err != err {
				t.Errorf("expected %+v got %+v", tt.err, err)
			}

			if !cmpOutcomes(tt.outcomes, outcomes) {
				t.Errorf("expected %+v got %+v", tt.outcomes, outcomes)
			}
		})
	}
}

func cmpOutcomes(ex []finding.Outcome, act []finding.Finding) bool {
	if len(ex) != len(act) {
		return false
	}

	for i := range ex {
		if act[i].Outcome != ex[i] {
			return false
		}
	}

	return true
}
