// Copyright 2023 OpenSSF Scorecard Authors
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
package evaluation

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/utests"
)

func TestContributors(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		raw      *checker.ContributorsData
		expected checker.CheckResult
	}{
		{
			name: "No data",
			raw:  nil,
			expected: checker.CheckResult{
				Version: 2,
				Score:   -1,
				Reason:  "internal error: empty raw data",
			},
		},
		{
			name: "No contributors",
			raw: &checker.ContributorsData{
				Users: []clients.User{},
			},
			expected: checker.CheckResult{
				Version: 2,
				Score:   0,
				Reason:  "0 different organizations found -- score normalized to 0",
			},
		},
		{
			name: "Contributors with orgs and number of contributions is greater than 5 with companies",
			raw: &checker.ContributorsData{
				Users: []clients.User{
					{
						NumContributions: 10,
						Organizations: []clients.User{
							{
								Login: "org1",
							},
						},
						Companies: []string{"company1"},
					},
					{
						NumContributions: 10,
						Organizations: []clients.User{
							{
								Login: "org2",
							},
						},
					},
					{
						NumContributions: 10,
						Organizations: []clients.User{
							{
								Login: "org3",
							},
						},
					},
					{
						NumContributions: 1,
						Organizations: []clients.User{
							{
								Login: "org2",
							},
						},
					},
				},
			},
			expected: checker.CheckResult{
				Version: 2,
				Score:   10,
				Reason:  "4 different organizations found -- score normalized to 10",
			},
		},
		{
			name: "Contributors with orgs and number of contributions is greater than 5 without companies",
			raw: &checker.ContributorsData{
				Users: []clients.User{
					{
						NumContributions: 10,
						Organizations: []clients.User{
							{
								Login: "org1",
							},
						},
					},
					{
						NumContributions: 10,
						Organizations: []clients.User{
							{
								Login: "org2",
							},
						},
					},
					{
						NumContributions: 10,
						Organizations: []clients.User{
							{
								Login: "org3",
							},
						},
					},
					{
						NumContributions: 1,
						Organizations: []clients.User{
							{
								Login: "org2",
							},
						},
					},
				},
			},
			expected: checker.CheckResult{
				Version: 2,
				Score:   10,
				Reason:  "3 different organizations found -- score normalized to 10",
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := Contributors("", &utests.TestDetailLogger{}, tc.raw)
			if !cmp.Equal(result, tc.expected, cmpopts.IgnoreFields(checker.CheckResult{}, "Error")) { //nolint:govet
				t.Errorf("expected %v, got %v", tc.expected, cmp.Diff(tc.expected, result, cmpopts.IgnoreFields(checker.CheckResult{}, "Error"))) //nolint:lll
			}
		})
	}
}
