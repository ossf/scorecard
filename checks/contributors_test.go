// Copyright 2022 Security Scorecard Authors
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

	"github.com/golang/mock/gomock"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
	scut "github.com/ossf/scorecard/v4/utests"
)

// TestContributors tests the contributors check.
func TestContributors(t *testing.T) {
	t.Parallel()
	//fieldalignment lint issue. Ignoring it as it is not important for this test.
	//nolint
	tests := []struct {
		err      error
		name     string
		contrib  []clients.User
		expected checker.CheckResult
	}{
		{
			err:  nil,
			name: "Two contributors without company",
			contrib: []clients.User{
				{
					Organizations: []clients.User{
						{
							Login: "org1",
						},
						{
							Login: "org2",
						},
					},
				},
			},
			expected: checker.CheckResult{
				Score: 0,
			},
		},
		{
			err:  nil,
			name: "Valid contributors with enough contributors and companies",
			contrib: []clients.User{
				{

					Companies:        []string{"company1"},
					NumContributions: 10,
					Organizations: []clients.User{
						{
							Login: "org1",
						},
						{
							Login: "org2",
						},
					},
				},
				{
					Companies:        []string{"company2"},
					NumContributions: 10,
					Organizations: []clients.User{
						{
							Login: "org1",
						},
						{
							Login: "org2",
						},
					},
				},
				{
					Companies:        []string{"company3"},
					NumContributions: 10,
					Organizations: []clients.User{
						{
							Login: "org1",
						},
						{
							Login: "org2",
						},
					},
				},
				{
					Companies:        []string{"company4"},
					NumContributions: 10,
					Organizations: []clients.User{
						{
							Login: "org1",
						},
						{
							Login: "org2",
						},
					},
				},
				{
					Companies:        []string{"company5"},
					NumContributions: 10,
					Organizations: []clients.User{
						{
							Login: "org1",
						},
						{
							Login: "org2",
						},
					},
				},
				{
					Companies: []string{"company6"},
					Organizations: []clients.User{
						{
							Login: "org1",
						},
						{
							Login: "org2",
						},
					},
				},
			},
			expected: checker.CheckResult{
				Score: 10,
			},
		},
		{
			err:     nil,
			name:    "No contributors",
			contrib: []clients.User{},
			expected: checker.CheckResult{
				Score: 0,
			},
		},
		{
			err:     errors.New("error"),
			name:    "Error getting contributors",
			contrib: []clients.User{},
			expected: checker.CheckResult{
				Score: -1,
			},
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockRepo := mockrepo.NewMockRepoClient(ctrl)
			mockRepo.EXPECT().ListContributors().DoAndReturn(func() ([]clients.User, error) {
				if tt.err != nil {
					return nil, tt.err
				}
				return tt.contrib, nil
			})

			req := checker.CheckRequest{
				RepoClient: mockRepo,
			}
			req.Dlogger = &scut.TestDetailLogger{}
			res := Contributors(&req)

			if tt.err != nil {
				if res.Error == nil {
					t.Errorf("Expected error %v, got nil", tt.err)
				}
				// return as we don't need to check the rest of the fields.
				return
			}

			if res.Score != tt.expected.Score {
				t.Errorf("Expected score %d, got %d for %v", tt.expected.Score, res.Score, tt.name)
			}
			ctrl.Finish()
		})
	}
}
