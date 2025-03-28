// Copyright 2022 OpenSSF Scorecard Authors
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
	"strings"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	scut "github.com/ossf/scorecard/v5/utests"
)

// TestContributors tests the contributors check.
//
//nolint:gocognit
func TestContributors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		contribErr     error
		ownErr         error
		name           string
		contrib        []clients.User
		own            []clients.User
		expectedDetail string
		expected       checker.CheckResult
	}{
		{
			ownErr:     nil,
			contribErr: nil,
			name:       "Two contributors without company",
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
			own: []clients.User{},
			expected: checker.CheckResult{
				Score: 0,
			},
		},
		{
			ownErr:     nil,
			contribErr: nil,
			name:       "Two owners with no contributors",
			contrib:    []clients.User{},
			own: []clients.User{
				{Login: "Login1"},
				{Login: "Login2"},
			},
			expected: checker.CheckResult{
				Score: 0,
			},
		},
		{
			contribErr: nil,
			ownErr:     nil,
			name:       "Valid contributors with enough contributions and companies",
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
			own: []clients.User{},
			expected: checker.CheckResult{
				Score: 5,
			},
			expectedDetail: "found contributions from: company1, company2, company3, company4, company5, org1, org2",
		},
		{
			contribErr: nil,
			ownErr:     nil,
			name:       "Enough contibutor owners",
			contrib: []clients.User{
				{Login: "login1"},
				{Login: "login2"},
				{Login: "login3"},
			},
			own: []clients.User{
				{Login: "login3"},
				{Login: "login1"},
				{Login: "login2"},
			},
			expected: checker.CheckResult{
				Score: 5,
			},
			expectedDetail: "found contributions from: login1, login2, login3",
		},
		{
			contribErr: nil,
			ownErr:     nil,
			name:       "Valid contributors with enough contributions, companies, and owners",
			contrib: []clients.User{
				{
					Login:            "Login1",
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
					Login:            "Login2",
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
					Login:            "Login3",
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
			own: []clients.User{
				{Login: "Login3"},
				{Login: "Login1"},
				{Login: "Login2"},
			},
			expected: checker.CheckResult{
				Score: 10,
			},
			expectedDetail: "found contributions from: Login1, Login2, Login3, company1, company2, company3, company4, company5, org1, org2",
		},
		{
			contribErr: nil,
			ownErr:     nil,
			name:       "No contributors",
			contrib:    []clients.User{},
			own:        []clients.User{},
			expected: checker.CheckResult{
				Score: 0,
			},
		},
		{
			contribErr: errors.New("error"),
			ownErr:     nil,
			name:       "Error getting contributors",
			contrib:    []clients.User{},
			own:        []clients.User{},
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
				if tt.contribErr != nil {
					return nil, tt.contribErr
				}
				return tt.contrib, nil
			})
			mockRepo.EXPECT().ListCodeOwners().DoAndReturn(func() ([]clients.User, error) {
				if tt.ownErr != nil {
					return nil, tt.ownErr
				}
				return tt.own, nil
			}).AnyTimes()

			req := checker.CheckRequest{
				RepoClient: mockRepo,
			}
			req.Dlogger = &scut.TestDetailLogger{}
			res := Contributors(&req)

			if tt.contribErr != nil || tt.ownErr != nil {
				if res.Error == nil {
					t.Errorf("Expected error %v, got nil", tt.contribErr)
				}
				// return as we don't need to check the rest of the fields.
				return
			}

			if res.Score != tt.expected.Score {
				t.Errorf("Expected score %d, got %d for %v", tt.expected.Score, res.Score, tt.name)
			}
			// make sure the output stays relatively stable
			if tt.expectedDetail != "" {
				details := req.Dlogger.Flush()
				if len(details) != 1 {
					t.Errorf("expected one check detail, got %d", len(details))
				}
				detail := details[0].Msg.Text
				if !strings.Contains(detail, tt.expectedDetail) {
					t.Errorf("expected %q but didn't find it: %q", tt.expectedDetail, detail)
				}
			}
		})
	}
}
