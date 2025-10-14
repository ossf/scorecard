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

package checks

import (
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	scut "github.com/ossf/scorecard/v5/utests"
)

// TestInactiveMaintainers tests the inactive maintainers check.
func TestInactiveMaintainers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		activity        map[string]bool
		err             error
		expected        checker.CheckResult
		expectedDetails scut.TestReturn
	}{
		{
			name: "All maintainers active",
			activity: map[string]bool{
				"maintainer1": true,
				"maintainer2": true,
			},
			err: nil,
			expected: checker.CheckResult{
				Score: checker.MaxResultScore,
			},
			expectedDetails: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 2,
				NumberOfWarn: 0,
			},
		},
		{
			name: "All maintainers inactive",
			activity: map[string]bool{
				"maintainer1": false,
				"maintainer2": false,
			},
			err: nil,
			expected: checker.CheckResult{
				Score: checker.MinResultScore,
			},
			expectedDetails: scut.TestReturn{
				Score:        checker.MinResultScore,
				NumberOfWarn: 2,
				NumberOfInfo: 0,
			},
		},
		{
			name: "Mixed active and inactive",
			activity: map[string]bool{
				"active":   true,
				"inactive": false,
			},
			err: nil,
			expected: checker.CheckResult{
				Score: 5, // 50% active
			},
			expectedDetails: scut.TestReturn{
				Score:        5,
				NumberOfInfo: 1,
				NumberOfWarn: 1,
			},
		},
		{
			name:     "No maintainers",
			activity: map[string]bool{},
			err:      nil,
			expected: checker.CheckResult{
				Score: checker.InconclusiveResultScore,
			},
		},
		{
			name:     "Error fetching maintainer activity",
			activity: nil,
			err:      errors.New("failed to fetch maintainer activity"),
			expected: checker.CheckResult{
				Score: checker.InconclusiveResultScore,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mockrepo.NewMockRepoClient(ctrl)
			mockRepo.EXPECT().
				GetMaintainerActivity(gomock.Any()).
				Return(tt.activity, tt.err).
				AnyTimes()
			mockRepo.EXPECT().URI().Return("github.com/ossf/scorecard").AnyTimes()

			dl := scut.TestDetailLogger{}
			req := &checker.CheckRequest{
				RepoClient: mockRepo,
				Dlogger:    &dl,
			}

			result := InactiveMaintainers(req)

			if result.Score != tt.expected.Score {
				t.Errorf("expected score %d, got %d", tt.expected.Score, result.Score)
			}

			// If we have expected details, validate them
			if tt.expectedDetails.NumberOfInfo > 0 || tt.expectedDetails.NumberOfWarn > 0 {
				scut.ValidateTestReturn(t, tt.name, &tt.expectedDetails, &result, &dl)
			}

			// For error cases, check that the score is inconclusive
			if tt.err != nil && result.Score != checker.InconclusiveResultScore {
				t.Errorf("expected inconclusive score for error case, got %d", result.Score)
			}
		})
	}
}
