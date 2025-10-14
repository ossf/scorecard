// Copyright 2026 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package raw

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
)

func TestInactiveMaintainers(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		mockError      error
		expectedError  error
		expectedResult checker.InactiveMaintainersData
		mockActivity   map[string]bool
		name           string
	}{
		{
			name: "All maintainers active",
			mockActivity: map[string]bool{
				"maintainer1": true,
				"maintainer2": true,
			},
			mockError: nil,
			expectedResult: checker.InactiveMaintainersData{
				MaintainerActivity: map[string]bool{
					"maintainer1": true,
					"maintainer2": true,
				},
			},
			expectedError: nil,
		},
		{
			name: "All maintainers inactive",
			mockActivity: map[string]bool{
				"maintainer1": false,
				"maintainer2": false,
			},
			mockError: nil,
			expectedResult: checker.InactiveMaintainersData{
				MaintainerActivity: map[string]bool{
					"maintainer1": false,
					"maintainer2": false,
				},
			},
			expectedError: nil,
		},
		{
			name: "Mixed active and inactive",
			mockActivity: map[string]bool{
				"active-maintainer":   true,
				"inactive-maintainer": false,
			},
			mockError: nil,
			expectedResult: checker.InactiveMaintainersData{
				MaintainerActivity: map[string]bool{
					"active-maintainer":   true,
					"inactive-maintainer": false,
				},
			},
			expectedError: nil,
		},
		{
			name:           "No maintainers",
			mockActivity:   map[string]bool{},
			mockError:      nil,
			expectedResult: checker.InactiveMaintainersData{MaintainerActivity: map[string]bool{}},
			expectedError:  nil,
		},
		{
			name:           "Error from client",
			mockActivity:   nil,
			mockError:      errors.New("failed to fetch maintainer activity"),
			expectedResult: checker.InactiveMaintainersData{},
			expectedError:  errors.New("GetMaintainerActivity: failed to fetch maintainer activity"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mockrepo.NewMockRepoClient(ctrl)
			mockClient.EXPECT().
				GetMaintainerActivity(gomock.Any()).
				Return(tc.mockActivity, tc.mockError).
				Times(1)

			req := &checker.CheckRequest{
				RepoClient: mockClient,
			}

			result, err := InactiveMaintainers(req)

			if tc.expectedError != nil {
				if err == nil || err.Error() != tc.expectedError.Error() {
					t.Errorf("expected error %v, got %v", tc.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if diff := cmp.Diff(tc.expectedResult, result); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
