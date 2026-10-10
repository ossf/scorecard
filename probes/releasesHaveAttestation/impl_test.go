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

package releasesHaveAttestation

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
			name: "no releases",
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
			name: "release with no assets",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Releases: []clients.Release{
						{
							TagName: "v1.0",
							Assets:  []clients.ReleaseAsset{},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeNotApplicable,
			},
		},
		{
			name: "one release, all assets attested",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Releases: []clients.Release{
						{
							TagName: "v1.0",
							Assets: []clients.ReleaseAsset{
								{
									Name:           "binary.tar.gz",
									Digest:         "sha256:abc123",
									HasAttestation: true,
								},
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
			name: "one release, asset missing digest",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Releases: []clients.Release{
						{
							TagName: "v1.0",
							Assets: []clients.ReleaseAsset{
								{
									Name:           "binary.tar.gz",
									Digest:         "",
									HasAttestation: false,
								},
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
			name: "one release, asset has digest but no attestation",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Releases: []clients.Release{
						{
							TagName: "v1.0",
							Assets: []clients.ReleaseAsset{
								{
									Name:           "binary.tar.gz",
									Digest:         "sha256:abc123",
									HasAttestation: false,
								},
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
			name: "two releases, both fully attested",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Releases: []clients.Release{
						{
							TagName: "v2.0",
							Assets: []clients.ReleaseAsset{
								{Name: "binary.tar.gz", Digest: "sha256:def456", HasAttestation: true},
							},
						},
						{
							TagName: "v1.0",
							Assets: []clients.ReleaseAsset{
								{Name: "binary.tar.gz", Digest: "sha256:abc123", HasAttestation: true},
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue,
				finding.OutcomeTrue,
			},
		},
		{
			name: "two releases, one attested and one not",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Releases: []clients.Release{
						{
							TagName: "v2.0",
							Assets: []clients.ReleaseAsset{
								{Name: "binary.tar.gz", Digest: "sha256:def456", HasAttestation: true},
							},
						},
						{
							TagName: "v1.0",
							Assets: []clients.ReleaseAsset{
								{Name: "binary.tar.gz", Digest: "sha256:abc123", HasAttestation: false},
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue,
				finding.OutcomeFalse,
			},
		},
		{
			name: "release with multiple assets, one missing attestation",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Releases: []clients.Release{
						{
							TagName: "v1.0",
							Assets: []clients.ReleaseAsset{
								{Name: "binary.tar.gz", Digest: "sha256:abc123", HasAttestation: true},
								{Name: "binary.exe", Digest: "sha256:def456", HasAttestation: false},
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
			name: "enforce lookback limit of 5 releases",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Releases: []clients.Release{
						{TagName: "v6.0", Assets: []clients.ReleaseAsset{{Name: "f", Digest: "sha256:1", HasAttestation: true}}},
						{TagName: "v5.0", Assets: []clients.ReleaseAsset{{Name: "f", Digest: "sha256:2", HasAttestation: true}}},
						{TagName: "v4.0", Assets: []clients.ReleaseAsset{{Name: "f", Digest: "sha256:3", HasAttestation: true}}},
						{TagName: "v3.0", Assets: []clients.ReleaseAsset{{Name: "f", Digest: "sha256:4", HasAttestation: true}}},
						{TagName: "v2.0", Assets: []clients.ReleaseAsset{{Name: "f", Digest: "sha256:5", HasAttestation: true}}},
						// v1.0 should not be checked
						{TagName: "v1.0", Assets: []clients.ReleaseAsset{{Name: "f", Digest: "sha256:6", HasAttestation: false}}},
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
