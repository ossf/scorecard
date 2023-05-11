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

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	scut "github.com/ossf/scorecard/v4/utests"
)

func TestSignedReleases(t *testing.T) {
	tests := []struct {
		name           string
		releases       []clients.Release
		expectedResult checker.CheckResult
	}{
		{
			name: "Full score",
			releases: []clients.Release{
				{
					TagName: "v1.0",
					Assets: []clients.ReleaseAsset{
						{Name: "binary.tar.gz"},
						{Name: "binary.tar.gz.sig"},
						{Name: "binary.tar.gz.intoto.jsonl"},
					},
				},
			},
			expectedResult: checker.CheckResult{
				Name:    "Signed-Releases",
				Version: 2,
				Score:   10,
				Reason:  "1 out of 1 artifacts are signed or have provenance",
			},
		},
		{
			name: "Partial score",
			releases: []clients.Release{
				{
					TagName: "v1.0",
					Assets: []clients.ReleaseAsset{
						{Name: "binary.tar.gz"},
						{Name: "binary.tar.gz.sig"},
					},
				},
			},
			expectedResult: checker.CheckResult{
				Name:    "Signed-Releases",
				Version: 2,
				Score:   8,
				Reason:  "1 out of 1 artifacts are signed or have provenance",
			},
		},
		{
			name: "No score",
			releases: []clients.Release{
				{
					TagName: "v1.0",
					Assets: []clients.ReleaseAsset{
						{Name: "binary.tar.gz"},
					},
				},
			},
			expectedResult: checker.CheckResult{
				Name:    "Signed-Releases",
				Version: 2,
				Score:   0,
				Reason:  "0 out of 1 artifacts are signed or have provenance",
			},
		},
		{
			name:           "No releases",
			releases:       []clients.Release{},
			expectedResult: checker.CreateInconclusiveResult("Signed-Releases", "no releases found"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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
