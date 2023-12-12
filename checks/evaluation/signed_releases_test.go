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
)

const (
	asset0 = 0
	asset1 = 1
	asset2 = 2
	asset3 = 3
	asset4 = 4
	asset5 = 5
	asset6 = 6
	asset7 = 7
	asset8 = 8
	asset9 = 9
)

func negativeSignedProbe(totalReleases, release, asset int) finding.Finding {
	return finding.Finding{
		Probe:   releasesAreSigned.Probe,
		Outcome: finding.OutcomeNegative,
		Values: map[string]int{
			fmt.Sprintf("v%d", release):       int(releasesAreSigned.ValueTypeRelease),
			fmt.Sprintf("artifact-%d", asset): int(releasesAreSigned.ValueTypeReleaseAsset),
		},
	}
}

func positiveSignedProbe(totalReleases, release, asset int) finding.Finding {
	return finding.Finding{
		Probe:   releasesAreSigned.Probe,
		Outcome: finding.OutcomePositive,
		Values: map[string]int{
			fmt.Sprintf("v%d", release):       int(releasesAreSigned.ValueTypeRelease),
			fmt.Sprintf("artifact-%d", asset): int(releasesAreSigned.ValueTypeReleaseAsset),
		},
	}
}

func negativeProvenanceProbe(totalReleases, release, asset int) finding.Finding {
	return finding.Finding{
		Probe:   releasesHaveProvenance.Probe,
		Outcome: finding.OutcomeNegative,
		Values: map[string]int{
			fmt.Sprintf("v%d", release):       int(releasesHaveProvenance.ValueTypeRelease),
			fmt.Sprintf("artifact-%d", asset): int(releasesHaveProvenance.ValueTypeReleaseAsset),
		},
	}
}

func positiveProvenanceProbe(totalReleases, release, asset int) finding.Finding {
	return finding.Finding{
		Probe:   releasesHaveProvenance.Probe,
		Outcome: finding.OutcomePositive,
		Values: map[string]int{
			fmt.Sprintf("v%d", release):       int(releasesHaveProvenance.ValueTypeRelease),
			fmt.Sprintf("artifact-%d", asset): int(releasesHaveProvenance.ValueTypeReleaseAsset),
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
				positiveSignedProbe(1, 0, 0),
				negativeProvenanceProbe(1, 0, 0),
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
				positiveSignedProbe(1, 0, 0),
				positiveProvenanceProbe(1, 0, 0),
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
				negativeSignedProbe(1, 0, 0),
				positiveProvenanceProbe(1, 0, 0),
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
				negativeSignedProbe(3, release0, asset0),
				negativeProvenanceProbe(3, release0, asset0),
				//     Asset 2:
				positiveSignedProbe(3, release0, asset1),
				negativeProvenanceProbe(3, release0, asset1),
				// Release 2
				//     Asset 1:
				negativeSignedProbe(3, release1, asset0),
				negativeProvenanceProbe(3, release1, asset0),
				// Release 2
				//     Asset 2:
				negativeSignedProbe(3, release1, asset1),
				negativeProvenanceProbe(3, release1, asset1),
				// Release 2
				//     Asset 3:
				negativeSignedProbe(3, release1, asset2),
				negativeProvenanceProbe(3, release1, asset2),
				// Release 3
				//     Asset 1:
				negativeSignedProbe(3, release2, asset0),
				positiveProvenanceProbe(3, release2, asset0),
				//     Asset 2:
				negativeSignedProbe(3, release2, asset1),
				positiveProvenanceProbe(3, release2, asset1),
				//     Asset 3:
				negativeSignedProbe(3, release2, asset2),
				negativeProvenanceProbe(3, release2, asset2),
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
				negativeSignedProbe(5, release0, asset0),
				negativeProvenanceProbe(5, release0, asset0),
				positiveSignedProbe(5, release0, asset1),
				negativeProvenanceProbe(5, release0, asset1),
				// Release 2:
				// Release 2, Asset 1:
				positiveSignedProbe(5, release1, asset1),
				negativeProvenanceProbe(5, release1, asset0),
				// Release 2, Asset 2:
				negativeSignedProbe(5, release1, asset1),
				negativeProvenanceProbe(5, release1, asset1),
				// Release 2, Asset 3:
				negativeSignedProbe(5, release1, asset2),
				negativeProvenanceProbe(5, release1, asset2),
				// Release 3, Asset 1:
				negativeSignedProbe(5, release2, asset0),
				positiveProvenanceProbe(5, release2, asset0),
				// Release 3, Asset 2:
				negativeSignedProbe(5, release2, asset1),
				negativeProvenanceProbe(5, release2, asset1),
				// Release 3, Asset 3:
				negativeSignedProbe(5, release2, asset2),
				negativeProvenanceProbe(5, release2, asset2),
				// Release 4, Asset 1:
				negativeSignedProbe(5, release3, asset0),
				positiveProvenanceProbe(5, release3, asset0),
				// Release 4, Asset 2:
				negativeSignedProbe(5, release3, asset1),
				negativeProvenanceProbe(5, release3, asset1),
				// Release 4, Asset 3:
				negativeSignedProbe(5, release3, asset2),
				negativeProvenanceProbe(5, release3, asset2),
				// Release 5, Asset 1:
				negativeSignedProbe(5, release4, asset0),
				negativeProvenanceProbe(5, release4, asset0),
				// Release 5, Asset 2:
				negativeSignedProbe(5, release4, asset1),
				negativeProvenanceProbe(5, release4, asset1),
				// Release 5, Asset 3:
				negativeSignedProbe(5, release4, asset2),
				negativeProvenanceProbe(5, release4, asset2),
				// Release 5, Asset 4:
				negativeSignedProbe(5, release4, asset3),
				negativeProvenanceProbe(5, release4, asset3),
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
				negativeSignedProbe(5, release0, asset0),
				negativeProvenanceProbe(5, release0, asset0),
				positiveSignedProbe(5, release0, asset1),
				negativeProvenanceProbe(5, release0, asset1),
				// Release 2:
				// Release 2, Asset 1:
				positiveSignedProbe(5, release1, asset0),
				negativeProvenanceProbe(5, release1, asset0),
				// Release 2, Asset 2:
				negativeSignedProbe(5, release1, asset1),
				negativeProvenanceProbe(5, release1, asset1),
				// Release 2, Asset 3:
				negativeSignedProbe(5, release1, asset2),
				negativeProvenanceProbe(5, release1, asset2),
				// Release 3, Asset 1:
				positiveSignedProbe(5, release2, asset0),
				positiveProvenanceProbe(5, release2, asset0),
				// Release 3, Asset 2:
				negativeSignedProbe(5, release2, asset1),
				negativeProvenanceProbe(5, release2, asset1),
				// Release 3, Asset 3:
				negativeSignedProbe(5, release2, asset2),
				negativeProvenanceProbe(5, release2, asset2),
				// Release 4, Asset 1:
				positiveSignedProbe(5, release3, asset0),
				positiveProvenanceProbe(5, release3, asset0),
				// Release 4, Asset 2:
				negativeSignedProbe(5, release3, asset1),
				negativeProvenanceProbe(5, release3, asset1),
				// Release 4, Asset 3:
				negativeSignedProbe(5, release3, asset2),
				negativeProvenanceProbe(5, release3, asset2),
				// Release 5, Asset 1:
				positiveSignedProbe(5, release4, asset0),
				negativeProvenanceProbe(5, release4, asset0),
				// Release 5, Asset 2:
				negativeSignedProbe(5, release4, asset1),
				negativeProvenanceProbe(5, release4, asset1),
				// Release 5, Asset 3:
				negativeSignedProbe(5, release4, asset2),
				negativeProvenanceProbe(5, release4, asset2),
				// Release 5, Asset 4:
				negativeSignedProbe(5, release4, asset3),
				negativeProvenanceProbe(5, release4, asset3),
			},
			result: scut.TestReturn{
				Score:         8,
				NumberOfInfo:  7,
				NumberOfWarn:  23,
				NumberOfDebug: 5,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			got := SignedReleases(tt.name, tt.findings, &dl)
			if !scut.ValidateTestReturn(t, tt.name, &tt.result, &got, &dl) {
				t.Errorf("got %v, expected %v", got, tt.result)
			}
		})
	}
}
