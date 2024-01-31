//  Copyright 2023 OpenSSF Scorecard Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package evaluation

import (
	"fmt"
	"testing"

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/releasesAreSigned"
	"github.com/ossf/scorecard/v4/probes/releasesHaveProvenance"
	scut "github.com/ossf/scorecard/v4/utests"
)

const (
	release0 = 0
	release1 = 1
	release2 = 2
	release3 = 3
	release4 = 4
	release5 = 5
)

const (
	asset0 = 0
	asset1 = 1
	asset2 = 2
	asset3 = 3
)

func signedProbe(release, asset int, outcome finding.Outcome) finding.Finding {
	return finding.Finding{
		Probe:   releasesAreSigned.Probe,
		Outcome: outcome,
		Values: map[string]string{
			releasesAreSigned.ReleaseNameKey: fmt.Sprintf("v%d", release),
			releasesAreSigned.AssetNameKey:   fmt.Sprintf("artifact-%d", asset),
		},
	}
}

func provenanceProbe(release, asset int, outcome finding.Outcome) finding.Finding {
	return finding.Finding{
		Probe:   releasesHaveProvenance.Probe,
		Outcome: outcome,
		Values: map[string]string{
			releasesHaveProvenance.ReleaseNameKey: fmt.Sprintf("v%d", release),
			releasesHaveProvenance.AssetNameKey:   fmt.Sprintf("artifact-%d", asset),
		},
	}
}

func TestSignedReleases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		findings []finding.Finding
		result   scut.TestReturn
	}{
		{
			name: "Has one release that is signed but no provenance",
			findings: []finding.Finding{
				signedProbe(0, 0, finding.OutcomePositive),
				provenanceProbe(0, 0, finding.OutcomeNegative),
			},
			result: scut.TestReturn{
				Score:         8,
				NumberOfInfo:  1,
				NumberOfWarn:  1,
				NumberOfDebug: 1,
			},
		},
		{
			name: "Has one release that is signed and has provenance",
			findings: []finding.Finding{
				signedProbe(0, 0, finding.OutcomePositive),
				provenanceProbe(0, 0, finding.OutcomePositive),
			},
			result: scut.TestReturn{
				Score:         10,
				NumberOfInfo:  2,
				NumberOfDebug: 1,
			},
		},
		{
			name: "Has one release that is not signed but has provenance",
			findings: []finding.Finding{
				signedProbe(0, 0, finding.OutcomeNegative),
				provenanceProbe(0, 0, finding.OutcomePositive),
			},
			result: scut.TestReturn{
				Score:         checker.MaxResultScore,
				NumberOfInfo:  1,
				NumberOfWarn:  1,
				NumberOfDebug: 1,
			},
		},

		{
			name: "3 releases. One release has one signed, and one release has two provenance.",
			findings: []finding.Finding{
				// Release 1:
				//     Asset 1:
				signedProbe(release0, asset0, finding.OutcomeNegative),
				provenanceProbe(release0, asset0, finding.OutcomeNegative),
				//     Asset 2:
				signedProbe(release0, asset1, finding.OutcomePositive),
				provenanceProbe(release0, asset1, finding.OutcomeNegative),
				// Release 2
				//     Asset 1:
				signedProbe(release1, asset0, finding.OutcomeNegative),
				provenanceProbe(release1, asset0, finding.OutcomeNegative),
				// Release 2
				//     Asset 2:
				signedProbe(release1, asset1, finding.OutcomeNegative),
				provenanceProbe(release1, asset1, finding.OutcomeNegative),
				// Release 2
				//     Asset 3:
				signedProbe(release1, asset2, finding.OutcomeNegative),
				provenanceProbe(release1, asset2, finding.OutcomeNegative),
				// Release 3
				//     Asset 1:
				signedProbe(release2, asset0, finding.OutcomeNegative),
				provenanceProbe(release2, asset0, finding.OutcomePositive),
				//     Asset 2:
				signedProbe(release2, asset1, finding.OutcomeNegative),
				provenanceProbe(release2, asset1, finding.OutcomePositive),
				//     Asset 3:
				signedProbe(release2, asset2, finding.OutcomeNegative),
				provenanceProbe(release2, asset2, finding.OutcomeNegative),
			},
			result: scut.TestReturn{
				Score:         6,
				NumberOfInfo:  3,
				NumberOfWarn:  13,
				NumberOfDebug: 3,
			},
		},
		{
			name: "5 releases. Two releases have one signed each, and two releases have one provenance each.",
			findings: []finding.Finding{
				// Release 1:
				// Release 1, Asset 1:
				signedProbe(release0, asset0, finding.OutcomeNegative),
				provenanceProbe(release0, asset0, finding.OutcomeNegative),
				signedProbe(release0, asset1, finding.OutcomePositive),
				provenanceProbe(release0, asset1, finding.OutcomeNegative),
				// Release 2:
				// Release 2, Asset 1:
				signedProbe(release1, asset1, finding.OutcomePositive),
				provenanceProbe(release1, asset0, finding.OutcomeNegative),
				// Release 2, Asset 2:
				signedProbe(release1, asset1, finding.OutcomeNegative),
				provenanceProbe(release1, asset1, finding.OutcomeNegative),
				// Release 2, Asset 3:
				signedProbe(release1, asset2, finding.OutcomeNegative),
				provenanceProbe(release1, asset2, finding.OutcomeNegative),
				// Release 3, Asset 1:
				signedProbe(release2, asset0, finding.OutcomeNegative),
				provenanceProbe(release2, asset0, finding.OutcomePositive),
				// Release 3, Asset 2:
				signedProbe(release2, asset1, finding.OutcomeNegative),
				provenanceProbe(release2, asset1, finding.OutcomeNegative),
				// Release 3, Asset 3:
				signedProbe(release2, asset2, finding.OutcomeNegative),
				provenanceProbe(release2, asset2, finding.OutcomeNegative),
				// Release 4, Asset 1:
				signedProbe(release3, asset0, finding.OutcomeNegative),
				provenanceProbe(release3, asset0, finding.OutcomePositive),
				// Release 4, Asset 2:
				signedProbe(release3, asset1, finding.OutcomeNegative),
				provenanceProbe(release3, asset1, finding.OutcomeNegative),
				// Release 4, Asset 3:
				signedProbe(release3, asset2, finding.OutcomeNegative),
				provenanceProbe(release3, asset2, finding.OutcomeNegative),
				// Release 5, Asset 1:
				signedProbe(release4, asset0, finding.OutcomeNegative),
				provenanceProbe(release4, asset0, finding.OutcomeNegative),
				// Release 5, Asset 2:
				signedProbe(release4, asset1, finding.OutcomeNegative),
				provenanceProbe(release4, asset1, finding.OutcomeNegative),
				// Release 5, Asset 3:
				signedProbe(release4, asset2, finding.OutcomeNegative),
				provenanceProbe(release4, asset2, finding.OutcomeNegative),
				// Release 5, Asset 4:
				signedProbe(release4, asset3, finding.OutcomeNegative),
				provenanceProbe(release4, asset3, finding.OutcomeNegative),
			},
			result: scut.TestReturn{
				Score:         7,
				NumberOfInfo:  4,
				NumberOfWarn:  26,
				NumberOfDebug: 5,
			},
		},
		{
			name: "5 releases. All have one signed artifact.",
			findings: []finding.Finding{
				// Release 1:
				// Release 1, Asset 1:
				signedProbe(release0, asset0, finding.OutcomeNegative),
				provenanceProbe(release0, asset0, finding.OutcomeNegative),
				signedProbe(release0, asset1, finding.OutcomePositive),
				provenanceProbe(release0, asset1, finding.OutcomeNegative),
				// Release 2:
				// Release 2, Asset 1:
				signedProbe(release1, asset0, finding.OutcomePositive),
				provenanceProbe(release1, asset0, finding.OutcomeNegative),
				// Release 2, Asset 2:
				signedProbe(release1, asset1, finding.OutcomeNegative),
				provenanceProbe(release1, asset1, finding.OutcomeNegative),
				// Release 2, Asset 3:
				signedProbe(release1, asset2, finding.OutcomeNegative),
				provenanceProbe(release1, asset2, finding.OutcomeNegative),
				// Release 3, Asset 1:
				signedProbe(release2, asset0, finding.OutcomePositive),
				provenanceProbe(release2, asset0, finding.OutcomePositive),
				// Release 3, Asset 2:
				signedProbe(release2, asset1, finding.OutcomeNegative),
				provenanceProbe(release2, asset1, finding.OutcomeNegative),
				// Release 3, Asset 3:
				signedProbe(release2, asset2, finding.OutcomeNegative),
				provenanceProbe(release2, asset2, finding.OutcomeNegative),
				// Release 4, Asset 1:
				signedProbe(release3, asset0, finding.OutcomePositive),
				provenanceProbe(release3, asset0, finding.OutcomePositive),
				// Release 4, Asset 2:
				signedProbe(release3, asset1, finding.OutcomeNegative),
				provenanceProbe(release3, asset1, finding.OutcomeNegative),
				// Release 4, Asset 3:
				signedProbe(release3, asset2, finding.OutcomeNegative),
				provenanceProbe(release3, asset2, finding.OutcomeNegative),
				// Release 5, Asset 1:
				signedProbe(release4, asset0, finding.OutcomePositive),
				provenanceProbe(release4, asset0, finding.OutcomeNegative),
				// Release 5, Asset 2:
				signedProbe(release4, asset1, finding.OutcomeNegative),
				provenanceProbe(release4, asset1, finding.OutcomeNegative),
				// Release 5, Asset 3:
				signedProbe(release4, asset2, finding.OutcomeNegative),
				provenanceProbe(release4, asset2, finding.OutcomeNegative),
				// Release 5, Asset 4:
				signedProbe(release4, asset3, finding.OutcomeNegative),
				provenanceProbe(release4, asset3, finding.OutcomeNegative),
			},
			result: scut.TestReturn{
				Score:         8,
				NumberOfInfo:  7,
				NumberOfWarn:  23,
				NumberOfDebug: 5,
			},
		},
		{
			name: "too many releases (6 when lookback is 5)",
			findings: []finding.Finding{
				// Release 1:
				// Release 1, Asset 1:
				signedProbe(release0, asset0, finding.OutcomePositive),
				provenanceProbe(release0, asset0, finding.OutcomePositive),
				// Release 2:
				// Release 2, Asset 1:
				signedProbe(release1, asset0, finding.OutcomePositive),
				provenanceProbe(release1, asset0, finding.OutcomePositive),
				// Release 3, Asset 1:
				signedProbe(release2, asset0, finding.OutcomePositive),
				provenanceProbe(release2, asset0, finding.OutcomePositive),
				// Release 4, Asset 1:
				signedProbe(release3, asset0, finding.OutcomePositive),
				provenanceProbe(release3, asset0, finding.OutcomePositive),
				// Release 5, Asset 1:
				signedProbe(release4, asset0, finding.OutcomePositive),
				provenanceProbe(release4, asset0, finding.OutcomePositive),
				// Release 6, Asset 1:
				signedProbe(release5, asset0, finding.OutcomePositive),
				provenanceProbe(release5, asset0, finding.OutcomePositive),
			},
			result: scut.TestReturn{
				Score:         checker.InconclusiveResultScore,
				Error:         sce.ErrScorecardInternal,
				NumberOfInfo:  12, // 2 (signed + provenance) for each release
				NumberOfDebug: 6,  // 1 for each release
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			got := SignedReleases(tt.name, tt.findings, &dl)
			scut.ValidateTestReturn(t, tt.name, &tt.result, &got, &dl)
		})
	}
}

func Test_getReleaseName(t *testing.T) {
	t.Parallel()
	type args struct {
		f *finding.Finding
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no release",
			args: args{
				f: &finding.Finding{
					Values: map[string]string{},
				},
			},
			want: "",
		},
		{
			name: "release",
			args: args{
				f: &finding.Finding{
					Values: map[string]string{
						releasesAreSigned.ReleaseNameKey: "v1",
					},
					Probe: releasesAreSigned.Probe,
				},
			},
			want: "v1",
		},
		{
			name: "release and asset",
			args: args{
				f: &finding.Finding{
					Values: map[string]string{
						releasesAreSigned.ReleaseNameKey: "v1",
						releasesAreSigned.AssetNameKey:   "artifact-1",
					},
					Probe: releasesAreSigned.Probe,
				},
			},
			want: "v1",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := getReleaseName(tt.args.f); got != tt.want {
				t.Errorf("getReleaseName() = %v, want %v", got, tt.want)
			}
		})
	}
}
