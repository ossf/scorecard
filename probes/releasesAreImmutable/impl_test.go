// Copyright 2025 OpenSSF Scorecard Authors
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

package releasesAreImmutable

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/internal/utils/test"
)

func Test_Run(t *testing.T) {
	t.Parallel()
	//nolint:govet
	tests := []struct {
		name     string
		raw      *checker.RawResults
		outcomes []finding.Outcome
		err      error
	}{
		{
			name: "No releases returns NotApplicable.",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Releases: []clients.Release{},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeNotApplicable,
			},
		},
		{
			name: "One immutable release.",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Releases: []clients.Release{
						{
							TagName:     "v1.0",
							IsImmutable: true,
							Assets: []clients.ReleaseAsset{
								{Name: "binary.tar.gz"},
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue,
			},
		},
		{
			name: "One non-immutable release.",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Releases: []clients.Release{
						{
							TagName:     "v1.0",
							IsImmutable: false,
							Assets: []clients.ReleaseAsset{
								{Name: "binary.tar.gz"},
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeFalse,
			},
		},
		{
			name: "Mix of immutable and non-immutable releases.",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Releases: []clients.Release{
						{
							TagName:     "v1.0",
							IsImmutable: true,
							Assets: []clients.ReleaseAsset{
								{Name: "binary.tar.gz"},
							},
						},
						{
							TagName:     "v2.0",
							IsImmutable: false,
							Assets: []clients.ReleaseAsset{
								{Name: "binary.tar.gz"},
							},
						},
						{
							TagName:     "v3.0",
							IsImmutable: true,
							Assets: []clients.ReleaseAsset{
								{Name: "binary.tar.gz"},
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue,
				finding.OutcomeFalse,
				finding.OutcomeTrue,
			},
		},
		{
			name: "Only checks last 5 releases.",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Releases: []clients.Release{
						immutableRelease("v1.0"),
						immutableRelease("v2.0"),
						immutableRelease("v3.0"),
						immutableRelease("v4.0"),
						immutableRelease("v5.0"),
						immutableRelease("v6.0"),
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue,
				finding.OutcomeTrue,
				finding.OutcomeTrue,
				finding.OutcomeTrue,
				finding.OutcomeTrue,
			},
		},
		{
			name: "Release with no assets is skipped.",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Releases: []clients.Release{
						immutableRelease("v1.0"),
						{TagName: "v2.0", IsImmutable: true, Assets: []clients.ReleaseAsset{}},
						immutableRelease("v3.0"),
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue,
				finding.OutcomeTrue,
			},
		},
	}
	for _, tt := range tests {
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
			test.AssertOutcomes(t, findings, tt.outcomes)
		})
	}
}

func immutableRelease(version string) clients.Release {
	return clients.Release{
		TagName:     version,
		IsImmutable: true,
		Assets: []clients.ReleaseAsset{
			{Name: "binary.tar.gz"},
		},
	}
}
