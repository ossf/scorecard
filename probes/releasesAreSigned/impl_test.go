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
package releasesAreSigned

import (
	"fmt"
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
			name: "Has one signed release.",
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
				finding.OutcomeTrue,
			},
		},
		{
			name: "Has two signed releases.",
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
				finding.OutcomeTrue,
				finding.OutcomeTrue,
			},
		},
		{
			name: "Has two unsigned releases.",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Releases: []clients.Release{
						{
							TagName: "v1.0",
							Assets: []clients.ReleaseAsset{
								{Name: "binary.tar.gz"},
								{Name: "binary.tar.gz.notSig"},
								{Name: "binary.tar.gz.intoto.jsonl"},
							},
						},
						{
							TagName: "v2.0",
							Assets: []clients.ReleaseAsset{
								{Name: "binary.tar.gz"},
								{Name: "binary.tar.gz.notSig"},
								{Name: "binary.tar.gz.intoto.jsonl"},
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
			name: "Has two unsigned releases and one signed release.",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Releases: []clients.Release{
						{
							TagName: "v1.0",
							Assets: []clients.ReleaseAsset{
								{Name: "binary.tar.gz"},
								{Name: "binary.tar.gz.notSig"},
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
							TagName: "v3.0",
							Assets: []clients.ReleaseAsset{
								{Name: "binary.tar.gz"},
								{Name: "binary.tar.gz.notSig"},
								{Name: "binary.tar.gz.intoto.jsonl"},
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeNegative,
				finding.OutcomeTrue,
				finding.OutcomeNegative,
			},
		},
		{
			name: "Many releases.",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Releases: []clients.Release{
						release("v0.8.5"),
						release("v0.8.4"),
						release("v0.8.3"),
						release("v0.8.2"),
						release("v0.8.1"),
						release("v0.8.0"),
						release("v0.7.0"),
						release("v0.6.0"),
						release("v0.5.0"),
						release("v0.4.0"),
						release("v0.3.0"),
						release("v0.2.0"),
						release("v0.1.0"),
						release("v0.0.6"),
						release("v0.0.5"),
						release("v0.0.4"),
						release("v0.0.3"),
						release("v0.0.2"),
						release("v0.0.1"),
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

func release(version string) clients.Release {
	return clients.Release{
		TagName:         version,
		URL:             fmt.Sprintf("https://github.com/test/test_artifact/releases/tag/%s", version),
		TargetCommitish: "master",
		Assets: []clients.ReleaseAsset{
			{
				Name: fmt.Sprintf("%s_checksums.txt", version),
				URL:  fmt.Sprintf("https://github.com/test/repo/releases/%s/%s_checksums.txt", version, version),
			},
			{
				Name: fmt.Sprintf("%s_checksums.txt.sig", version),
				URL:  fmt.Sprintf("https://github.com/test/repo/releases/%s/%s_checksums.txt.sig", version, version),
			},
			{
				Name: fmt.Sprintf("%s_darwin_x86_64.tar.gz", version),
				URL:  fmt.Sprintf("https://github.com/test/repo/releases/%s/%s_darwin_x86_64.tar.gz", version, version),
			},
			{
				Name: fmt.Sprintf("%s_Linux_arm64.tar.gz", version),
				URL:  fmt.Sprintf("https://github.com/test/repo/releases/%s/%s_Linux_arm64.tar.gz", version, version),
			},
			{
				Name: fmt.Sprintf("%s_Linux_i386.tar.gz", version),
				URL:  fmt.Sprintf("https://github.com/test/repo/releases/%s/%s_Linux_i386.tar.gz", version, version),
			},
			{
				Name: fmt.Sprintf("%s_Linux_x86_64.tar.gz", version),
				URL:  fmt.Sprintf("https://github.com/test/repo/releases/%s/%s_Linux_x86_64.tar.gz", version, version),
			},
			{
				Name: fmt.Sprintf("%s_windows_i386.tar.gz", version),
				URL:  fmt.Sprintf("https://github.com/test/repo/releases/%s/%s_windows_i386.tar.gz", version, version),
			},
			{
				Name: fmt.Sprintf("%s_windows_x86_64.tar.gz", version),
				URL:  fmt.Sprintf("https://github.com/test/repo/releases/%s/%s_windows_x86_64.tar.gz", version, version),
			},
		},
	}
}
