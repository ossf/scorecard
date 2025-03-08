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

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
)

func TestCleanCompaniesOrg(t *testing.T) {
	t.Parallel()
	testCases := []struct { //nolint:govet
		name     string
		user     clients.User
		expected clients.User
	}{
		{
			name: "removes duplicate orgs and companies",
			user: clients.User{
				Login:            "user1",
				NumContributions: 5,
				Companies:        []string{"Company1", "Company1"},
				Organizations: []clients.User{
					{Login: "org1"},
					{Login: "org1"},
				},
			},
			expected: clients.User{

				Login:            "user1",
				NumContributions: 5,
				Companies:        []string{"company1"},
				Organizations: []clients.User{
					{Login: "org1"},
				},
			},
		},
		{
			name: "normalizes company names",
			user: clients.User{
				Login:            "user1",
				NumContributions: 5,
				Companies:        []string{"Company1", "@Company2", "Company3 inc.", "Company4 llc", "Company,5", "COMPANY6"},
				Organizations: []clients.User{
					{Login: "org1"},
				},
			},
			expected: clients.User{
				Login:            "user1",
				NumContributions: 5,
				Companies:        []string{"company1", "company2", "company3", "company4", "company5", "company6"},
				Organizations: []clients.User{
					{Login: "org1"},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := cleanCompaniesOrgs(&tc.user)
			if diff := cmp.Diff(result.Companies, tc.expected.Companies); diff != "" {
				t.Errorf("unexpected companies data (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(result.Organizations, tc.expected.Organizations); diff != "" {
				t.Errorf("unexpected organizations data (-want +got):\n%s", diff)
			}
		})
	}
}
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
	t.Parallel()
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
	codeOwners := []clients.User{
		{Login: "user1"},
		{Login: "user2"},
	}

	mockRepoClient.EXPECT().ListContributors().Return(contributors, nil)
	mockRepoClient.EXPECT().ListCodeOwners().Return(codeOwners, nil)
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
	expectedCodeOwners := []clients.User{
		{Login: "user1"},
		{Login: "user2"},
	}

	if diff := cmp.Diff(expectedUsers, data.Contributors); diff != "" {
		t.Errorf("unexpected contributors data (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(expectedCodeOwners, data.CodeOwners); diff != "" {
		t.Errorf("unexpected code owners data (-want +got):\n%s", diff)
	}
}
