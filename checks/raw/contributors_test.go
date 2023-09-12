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
package raw

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
)

func TestCompanyContains(t *testing.T) {
	t.Parallel()
	testCases := []struct { //nolint:govet
		name     string
		cs       []string
		company  string
		expected bool
	}{
		{
			name:     "company is present",
			cs:       []string{"Google", "Facebook", "OpenAI"},
			company:  "OpenAI",
			expected: true,
		},
		{
			name:     "company is not present",
			cs:       []string{"Google", "Facebook", "OpenAI"},
			company:  "Microsoft",
			expected: false,
		},
		{
			name:     "empty slice",
			cs:       []string{},
			company:  "Microsoft",
			expected: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := companyContains(tc.cs, tc.company)
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestOrgContains(t *testing.T) {
	t.Parallel()
	testCases := []struct { //nolint:govet
		name     string
		os       []clients.User
		login    string
		expected bool
	}{
		{
			name:     "user is present",
			os:       []clients.User{{Login: "alice"}, {Login: "bob"}, {Login: "charlie"}},
			login:    "bob",
			expected: true,
		},
		{
			name:     "user is not present",
			os:       []clients.User{{Login: "alice"}, {Login: "bob"}, {Login: "charlie"}},
			login:    "david",
			expected: false,
		},
		{
			name:     "empty slice",
			os:       []clients.User{},
			login:    "alice",
			expected: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := orgContains(tc.os, tc.login)
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestContributors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
	contributors := []clients.User{
		{
			Login:            "user1",
			NumContributions: 5,
			Companies:        []string{"Company1", "Company2"},
			Organizations: []clients.User{
				{Login: "org1"},
				{Login: "org2"},
			},
		},
		{
			Login:            "user2",
			NumContributions: 3,
			Companies:        []string{"Company3", "Company4"},
			Organizations: []clients.User{
				{Login: "org3"},
				{Login: "org4"},
			},
		},
	}

	mockRepoClient.EXPECT().ListContributors().Return(contributors, nil)
	req := &checker.CheckRequest{
		RepoClient: mockRepoClient,
	}
	data, err := Contributors(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedUsers := []clients.User{
		{
			Login:            "user1",
			NumContributions: 5,
			Companies:        []string{"company1", "company2"},
			Organizations: []clients.User{
				{Login: "org1"},
				{Login: "org2"},
			},
		},
		{
			Login:            "user2",
			NumContributions: 3,
			Companies:        []string{"company3", "company4"},
			Organizations: []clients.User{
				{Login: "org3"},
				{Login: "org4"},
			},
		},
	}

	if diff := cmp.Diff(expectedUsers, data.Users); diff != "" {
		t.Errorf("unexpected contributors data (-want +got):\n%s", diff)
	}
}
