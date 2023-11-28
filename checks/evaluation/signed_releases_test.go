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
	"testing"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	scut "github.com/ossf/scorecard/v4/utests"
)

const (
	releaseIndex0 = 0
	releaseIndex1 = 1
	releaseIndex2 = 2
	releaseIndex3 = 3
	releaseIndex4 = 4
)

const (
	assetIndex0 = 0
	assetIndex1 = 1
	assetIndex2 = 2
	assetIndex3 = 3
	assetIndex4 = 4
	assetIndex5 = 5
	assetIndex6 = 6
	assetIndex7 = 7
	assetIndex8 = 8
	assetIndex9 = 9
)

func negativeSignedProbe(totalReleases, releaseindex, assetIndex int) finding.Finding {
	return finding.Finding{
		Probe:   "releasesAreSigned",
		Outcome: finding.OutcomeNegative,
		Values: map[string]int{
			"totalReleases": totalReleases,
			"releaseIndex":  releaseindex,
			"assetIndex":    assetIndex,
		},
	}
}

func negativeProvenanceProbe(totalReleases, releaseindex, assetIndex int) finding.Finding {
	return finding.Finding{
		Probe:   "releasesHaveProvenance",
		Outcome: finding.OutcomeNegative,
		Values: map[string]int{
			"totalReleases": totalReleases,
			"releaseIndex":  releaseindex,
			"assetIndex":    assetIndex,
		},
	}
}

func positiveProvenanceProbe(totalReleases, releaseindex, assetIndex int) finding.Finding {
	return finding.Finding{
		Probe:   "releasesHaveProvenance",
		Outcome: finding.OutcomePositive,
		Values: map[string]int{
			"totalReleases": totalReleases,
			"releaseIndex":  releaseindex,
			"assetIndex":    assetIndex,
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
				{
					Probe:   "releasesAreSigned",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"totalReleases": 1,
						"releaseIndex":  0,
						"assetIndex":    0,
					},
				},
				{
					Probe:   "releasesHaveProvenance",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"totalReleases": 1,
						"releaseIndex":  0,
						"assetIndex":    0,
					},
				},
			},
			result: scut.TestReturn{
				Score:        8,
				NumberOfInfo: 1,
				NumberOfWarn: 1,
			},
		},
		{
			name: "Has one release that is signed and has provenance",
			findings: []finding.Finding{
				{
					Probe:   "releasesAreSigned",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"totalReleases": 1,
						"releaseIndex":  0,
						"assetIndex":    0,
					},
				},
				{
					Probe:   "releasesHaveProvenance",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"totalReleases": 1,
						"releaseIndex":  0,
						"assetIndex":    0,
					},
				},
			},
			result: scut.TestReturn{
				Score:        10,
				NumberOfInfo: 2,
			},
		},
		{
			name: "Has one release that is not signed but has provenance",
			findings: []finding.Finding{
				{
					Probe:   "releasesAreSigned",
					Outcome: finding.OutcomeNegative,
					Values: map[string]int{
						"totalReleases": 1,
						"releaseIndex":  0,
						"assetIndex":    0,
					},
				},
				{
					Probe:   "releasesHaveProvenance",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"totalReleases": 1,
						"releaseIndex":  0,
						"assetIndex":    0,
					},
				},
			},
			result: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 1,
				NumberOfWarn: 1,
			},
		},

		{
			name: "3 releases. One release has one signed, and one release has two provenance.",
			findings: []finding.Finding{
				// Release 1:
				// Release 1, Asset 1:
				negativeSignedProbe(3, releaseIndex0, assetIndex0),
				negativeProvenanceProbe(3, releaseIndex0, assetIndex0),
				{
					Probe:   "releasesAreSigned",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"totalReleases": 3,
						"releaseIndex":  0,
						"assetIndex":    1,
					},
				},
				negativeProvenanceProbe(3, releaseIndex0, assetIndex1),
				// Release 2:
				// Release 2, Asset 1:
				negativeSignedProbe(3, releaseIndex1, assetIndex0),
				negativeProvenanceProbe(3, releaseIndex1, assetIndex0),
				// Release 2, Asset 2:
				negativeSignedProbe(3, releaseIndex1, assetIndex1),
				negativeProvenanceProbe(3, releaseIndex1, assetIndex1),
				// Release 2, Asset 3:
				negativeSignedProbe(3, releaseIndex1, assetIndex2),
				negativeProvenanceProbe(3, releaseIndex1, assetIndex2),
				// Release 3, Asset 1:
				negativeSignedProbe(3, releaseIndex2, assetIndex0),
				positiveProvenanceProbe(3, releaseIndex2, assetIndex0),
				// Release 3, Asset 2:
				negativeSignedProbe(3, releaseIndex2, assetIndex1),
				positiveProvenanceProbe(3, releaseIndex2, assetIndex1),
				// Release 3, Asset 3:
				negativeSignedProbe(3, releaseIndex2, assetIndex2),
				negativeProvenanceProbe(3, releaseIndex2, assetIndex2),
			},
			result: scut.TestReturn{
				Score:        6,
				NumberOfInfo: 3,
				NumberOfWarn: 13,
			},
		},
		{
			name: "5 releases. Two releases have one signed each, and two releases have one provenance each.",
			findings: []finding.Finding{
				// Release 1:
				// Release 1, Asset 1:
				negativeSignedProbe(5, releaseIndex0, assetIndex0),
				negativeProvenanceProbe(5, releaseIndex0, assetIndex0),
				{
					Probe:   "releasesAreSigned",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"totalReleases": 5,
						"releaseIndex":  0,
						"assetIndex":    1,
					},
				},
				negativeProvenanceProbe(5, releaseIndex0, assetIndex1),
				// Release 2:
				// Release 2, Asset 1:
				{
					Probe:   "releasesAreSigned",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"totalReleases": 5,
						"releaseIndex":  1,
						"assetIndex":    0,
					},
				},
				negativeProvenanceProbe(5, releaseIndex1, assetIndex0),
				// Release 2, Asset 2:
				negativeSignedProbe(5, releaseIndex1, assetIndex1),
				negativeProvenanceProbe(5, releaseIndex1, assetIndex1),
				// Release 2, Asset 3:
				negativeSignedProbe(5, releaseIndex1, assetIndex2),
				negativeProvenanceProbe(5, releaseIndex1, assetIndex2),
				// Release 3, Asset 1:
				negativeSignedProbe(5, releaseIndex2, assetIndex0),
				positiveProvenanceProbe(5, releaseIndex2, assetIndex0),
				// Release 3, Asset 2:
				negativeSignedProbe(5, releaseIndex2, assetIndex1),
				negativeProvenanceProbe(5, releaseIndex2, assetIndex1),
				// Release 3, Asset 3:
				negativeSignedProbe(5, releaseIndex2, assetIndex2),
				negativeProvenanceProbe(5, releaseIndex2, assetIndex2),
				// Release 4, Asset 1:
				negativeSignedProbe(5, releaseIndex3, assetIndex0),
				positiveProvenanceProbe(5, releaseIndex3, assetIndex0),
				// Release 4, Asset 2:
				negativeSignedProbe(5, releaseIndex3, assetIndex1),
				negativeProvenanceProbe(5, releaseIndex3, assetIndex1),
				// Release 4, Asset 3:
				negativeSignedProbe(5, releaseIndex3, assetIndex2),
				negativeProvenanceProbe(5, releaseIndex3, assetIndex2),
				// Release 5, Asset 1:
				negativeSignedProbe(5, releaseIndex4, assetIndex0),
				negativeProvenanceProbe(5, releaseIndex4, assetIndex0),
				// Release 5, Asset 2:
				negativeSignedProbe(5, releaseIndex4, assetIndex1),
				negativeProvenanceProbe(5, releaseIndex4, assetIndex1),
				// Release 5, Asset 3:
				negativeSignedProbe(5, releaseIndex4, assetIndex2),
				negativeProvenanceProbe(5, releaseIndex4, assetIndex2),
				// Release 5, Asset 4:
				negativeSignedProbe(5, releaseIndex4, assetIndex3),
				negativeProvenanceProbe(5, releaseIndex4, assetIndex3),
			},
			result: scut.TestReturn{
				Score:        7,
				NumberOfInfo: 4,
				NumberOfWarn: 26,
			},
		},
		{
			name: "5 releases. All have one signed artifact.",
			findings: []finding.Finding{
				// Release 1:
				// Release 1, Asset 1:
				negativeSignedProbe(5, releaseIndex0, assetIndex0),
				negativeProvenanceProbe(5, releaseIndex0, assetIndex0),
				{
					Probe:   "releasesAreSigned",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"totalReleases": 5,
						"releaseIndex":  0,
						"assetIndex":    1,
					},
				},
				negativeProvenanceProbe(5, releaseIndex0, assetIndex1),
				// Release 2:
				// Release 2, Asset 1:
				{
					Probe:   "releasesAreSigned",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"totalReleases": 5,
						"releaseIndex":  1,
						"assetIndex":    0,
					},
				},
				negativeProvenanceProbe(5, releaseIndex1, assetIndex0),
				// Release 2, Asset 2:
				negativeSignedProbe(5, releaseIndex1, assetIndex1),
				negativeProvenanceProbe(5, releaseIndex1, assetIndex1),
				// Release 2, Asset 3:
				negativeSignedProbe(5, releaseIndex1, assetIndex2),
				negativeProvenanceProbe(5, releaseIndex1, assetIndex2),
				// Release 3, Asset 1:
				{
					Probe:   "releasesAreSigned",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"totalReleases": 5,
						"releaseIndex":  2,
						"assetIndex":    0,
					},
				},
				positiveProvenanceProbe(5, releaseIndex2, assetIndex0),
				// Release 3, Asset 2:
				negativeSignedProbe(5, releaseIndex2, assetIndex1),
				negativeProvenanceProbe(5, releaseIndex2, assetIndex1),
				// Release 3, Asset 3:
				negativeSignedProbe(5, releaseIndex2, assetIndex2),
				negativeProvenanceProbe(5, releaseIndex2, assetIndex2),
				// Release 4, Asset 1:
				{
					Probe:   "releasesAreSigned",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"totalReleases": 5,
						"releaseIndex":  3,
						"assetIndex":    0,
					},
				},
				positiveProvenanceProbe(5, releaseIndex3, assetIndex0),
				// Release 4, Asset 2:
				negativeSignedProbe(5, releaseIndex3, assetIndex1),
				negativeProvenanceProbe(5, releaseIndex3, assetIndex1),
				// Release 4, Asset 3:
				negativeSignedProbe(5, releaseIndex3, assetIndex2),
				negativeProvenanceProbe(5, releaseIndex3, assetIndex2),
				// Release 5, Asset 1:
				{
					Probe:   "releasesAreSigned",
					Outcome: finding.OutcomePositive,
					Values: map[string]int{
						"totalReleases": 5,
						"releaseIndex":  4,
						"assetIndex":    0,
					},
				},
				negativeProvenanceProbe(5, releaseIndex4, assetIndex0),
				// Release 5, Asset 2:
				negativeSignedProbe(5, releaseIndex4, assetIndex1),
				negativeProvenanceProbe(5, releaseIndex4, assetIndex1),
				// Release 5, Asset 3:
				negativeSignedProbe(5, releaseIndex4, assetIndex2),
				negativeProvenanceProbe(5, releaseIndex4, assetIndex2),
				// Release 5, Asset 4:
				negativeSignedProbe(5, releaseIndex4, assetIndex3),
				negativeProvenanceProbe(5, releaseIndex4, assetIndex3),
			},
			result: scut.TestReturn{
				Score:        8,
				NumberOfInfo: 7,
				NumberOfWarn: 23,
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dl := &scut.TestDetailLogger{}
			data := &checker.SignedReleasesData{Releases: tc.releases}
			actualResult := SignedReleases("Signed-Releases", dl, data)

			if !cmp.Equal(tc.expectedResult, actualResult,
				cmpopts.IgnoreFields(checker.CheckResult{}, "Error")) {
				t.Errorf("SignedReleases() mismatch (-want +got):\n%s", cmp.Diff(tc.expectedResult, actualResult,
					cmpopts.IgnoreFields(checker.CheckResult{}, "Error")))
			}
		})
	}
}
