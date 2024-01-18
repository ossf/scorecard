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
package releasesHaveProvenance

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/internal/utils/test"
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
			name: "Has one release with provenance.",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Releases: []clients.Release{
						{
							TagName: "v1.0",
							Assets: []clients.ReleaseAsset{
								{Name: "binary.tar.gz"},
								{Name: "binary.tar.gz.sig"},
								{Name: "binary.tar.gz.intoto.jsonl"},
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomePositive,
			},
		},
		{
			name: "Has two releases with provenance.",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Releases: []clients.Release{
						{
							TagName: "v1.0",
							Assets: []clients.ReleaseAsset{
								{Name: "binary.tar.gz"},
								{Name: "binary.tar.gz.sig"},
								{Name: "binary.tar.gz.intoto.jsonl"},
							},
						},
						{
							TagName: "v2.0",
							Assets: []clients.ReleaseAsset{
								{Name: "binary.tar.gz"},
								{Name: "binary.tar.gz.sig"},
								{Name: "binary.tar.gz.intoto.jsonl"},
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomePositive,
				finding.OutcomePositive,
			},
		},
		{
			name: "Has two releases without provenance.",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Releases: []clients.Release{
						{
							TagName: "v1.0",
							Assets: []clients.ReleaseAsset{
								{Name: "binary.tar.gz"},
								{Name: "binary.tar.gz.sig"},
								{Name: "binary.tar.gz.intoto.notJsonl"},
							},
						},
						{
							TagName: "v2.0",
							Assets: []clients.ReleaseAsset{
								{Name: "binary.tar.gz"},
								{Name: "binary.tar.gz.sig"},
								{Name: "binary.tar.gz.intoto.notJsonl"},
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeNegative,
				finding.OutcomeNegative,
			},
		},
		{
			name: "Has two releases without provenance and one with.",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Releases: []clients.Release{
						{
							TagName: "v1.0",
							Assets: []clients.ReleaseAsset{
								{Name: "binary.tar.gz"},
								{Name: "binary.tar.gz.notSig"},
								{Name: "binary.tar.gz.intoto.notJsonl"},
							},
						},
						{
							TagName: "v2.0",
							Assets: []clients.ReleaseAsset{
								{Name: "binary.tar.gz"},
								{Name: "binary.tar.gz.sig"},
								{Name: "binary.tar.gz.intoto.jsonl"},
							},
						},

						{
							TagName: "v3.0",
							Assets: []clients.ReleaseAsset{
								{Name: "binary.tar.gz"},
								{Name: "binary.tar.gz.sig"},
								{Name: "binary.tar.gz.intoto.notJsonl"},
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeNegative,
				finding.OutcomePositive,
				finding.OutcomeNegative,
			},
		},
		{
			name: "enforce lookback limit of 5 releases",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Releases: []clients.Release{
						{
							TagName: "v6.0",
							Assets: []clients.ReleaseAsset{
								{Name: "binary.tar.gz"},
								{Name: "binary.tar.gz.sig"},
								{Name: "binary.tar.gz.intoto.jsonl"},
							},
						},
						{
							TagName: "v5.0",
							Assets: []clients.ReleaseAsset{
								{Name: "binary.tar.gz"},
								{Name: "binary.tar.gz.sig"},
								{Name: "binary.tar.gz.intoto.jsonl"},
							},
						},
						{
							TagName: "v4.0",
							Assets: []clients.ReleaseAsset{
								{Name: "binary.tar.gz"},
								{Name: "binary.tar.gz.sig"},
								{Name: "binary.tar.gz.intoto.jsonl"},
							},
						},
						{
							TagName: "v3.0",
							Assets: []clients.ReleaseAsset{
								{Name: "binary.tar.gz"},
								{Name: "binary.tar.gz.sig"},
								{Name: "binary.tar.gz.intoto.jsonl"},
							},
						},
						{
							TagName: "v2.0",
							Assets: []clients.ReleaseAsset{
								{Name: "binary.tar.gz"},
								{Name: "binary.tar.gz.sig"},
								{Name: "binary.tar.gz.intoto.jsonl"},
							},
						},
						{
							TagName: "v1.0",
							Assets: []clients.ReleaseAsset{
								{Name: "binary.tar.gz"},
								{Name: "binary.tar.gz.sig"},
								{Name: "binary.tar.gz.intoto.jsonl"},
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomePositive,
				finding.OutcomePositive,
				finding.OutcomePositive,
				finding.OutcomePositive,
				finding.OutcomePositive,
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
			test.AssertOutcomes(t, findings, tt.outcomes)
		})
	}
}
