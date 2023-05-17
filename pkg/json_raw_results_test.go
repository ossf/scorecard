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
package pkg

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
)

func TestAsPointer(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet
		name     string
		input    string
		expected *string
	}{
		{
			name:     "test_empty_string",
			input:    "",
			expected: asPointer(""),
		},
		{
			name:     "test_non_empty_string",
			input:    "test",
			expected: asPointer("test"),
		},
		{
			name:     "test_number_string",
			input:    "123",
			expected: asPointer("123"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := asPointer(tt.input)
			if *result != *tt.expected {
				t.Errorf("asPointer() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestJsonScorecardRawResult_AddPackagingRawResults(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet
		name      string
		input     *checker.PackagingData
		wantError bool
	}{
		{
			name: "test_with_nil_file_field",
			input: &checker.PackagingData{
				Packages: []checker.Package{
					{File: nil},
				},
			},
			wantError: true,
		},
		{
			name: "test_with_empty_package_data",
			input: &checker.PackagingData{
				Packages: []checker.Package{},
			},
			wantError: false,
		},
		{
			name: "test_with_valid_package_data",
			input: &checker.PackagingData{
				Packages: []checker.Package{
					{
						File: &checker.File{
							Path:    "testPath",
							Offset:  0,
							Snippet: "testSnippet",
						},
						Runs: []checker.Run{
							{URL: "testUrl"},
						},
					},
				},
			},
			wantError: false,
		},
		{
			name: "test_with_package_with_msg",
			input: &checker.PackagingData{
				Packages: []checker.Package{
					{
						Msg: asPointer("testMsg"),
					},
				},
			},
		},
		{
			name: "test_with_package_with_file_nil",
			input: &checker.PackagingData{
				Packages: []checker.Package{
					{
						File: nil,
					},
				},
			},
			wantError: true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			r := &jsonScorecardRawResult{}
			err := r.addPackagingRawResults(test.input)
			if (err != nil) != test.wantError {
				t.Errorf("addPackagingRawResults() error = %v, wantError %v", err, test.wantError)
			}
		})
	}
}

func TestJsonScorecardRawResult_AddTokenPermissionsRawResults(t *testing.T) {
	t.Parallel()
	loc := checker.PermissionLocation("testLocationType")
	tests := []struct { //nolint:govet
		name      string
		input     *checker.TokenPermissionsData
		wantError bool
	}{
		{
			name: "test_with_nil_location_type",
			input: &checker.TokenPermissionsData{
				TokenPermissions: []checker.TokenPermission{
					{
						LocationType: nil,
						Type:         checker.PermissionLevelUndeclared,
					},
				},
			},
			wantError: true,
		},
		{
			name: "test_with_debug_message",
			input: &checker.TokenPermissionsData{
				TokenPermissions: []checker.TokenPermission{
					{
						LocationType: &loc,
						Type:         checker.PermissionLevelRead,
					},
				},
			},
		},
		{
			name: "test_with_nil_job_and_file",
			input: &checker.TokenPermissionsData{
				TokenPermissions: []checker.TokenPermission{
					{
						LocationType: &loc,
						Type:         checker.PermissionLevelUndeclared,
					},
				},
			},
			wantError: false,
		},
		{
			name: "test_with_valid_data",
			input: &checker.TokenPermissionsData{
				TokenPermissions: []checker.TokenPermission{
					{
						LocationType: &loc,
						Type:         checker.PermissionLevelUndeclared,
						Job: &checker.WorkflowJob{
							Name: asPointer("testJobName"),
							ID:   asPointer("testJobID"),
						},
						File: &checker.File{
							Path:    "testPath",
							Offset:  0,
							Snippet: "testSnippet",
						},
					},
				},
			},
			wantError: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			r := &jsonScorecardRawResult{}
			err := r.addTokenPermissionsRawResults(test.input)
			if (err != nil) != test.wantError {
				t.Errorf("addTokenPermissionsRawResults() error = %v, wantError %v", err, test.wantError)
			}
		})
	}
}

func TestJsonScorecardRawResult_AddDependencyPinningRawResults(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet
		name      string
		input     *checker.PinningDependenciesData
		wantError bool
	}{
		{
			name: "test_with_nil_location",
			input: &checker.PinningDependenciesData{
				Dependencies: []checker.Dependency{
					{Location: nil},
				},
			},
			wantError: false,
		},
		{
			name: "test_with_valid_data",
			input: &checker.PinningDependenciesData{
				Dependencies: []checker.Dependency{
					{
						Location: &checker.File{
							Path:      "testPath",
							Offset:    0,
							EndOffset: 5,
							Snippet:   "testSnippet",
						},
						Name:     asPointer("testDependency"),
						PinnedAt: asPointer("testPinnedAt"),
						Type:     checker.DependencyUseTypeGHAction,
					},
				},
			},
			wantError: false,
		},
		{
			name: "test_with_nil_location",
			input: &checker.PinningDependenciesData{
				Dependencies: []checker.Dependency{
					{
						Location: nil,
						Name:     asPointer("testDependency"),
						PinnedAt: asPointer("testPinnedAt"),
						Type:     checker.DependencyUseTypeGHAction,
					},
				},
			},
			wantError: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			r := &jsonScorecardRawResult{}
			err := r.addDependencyPinningRawResults(test.input)
			if (err != nil) != test.wantError {
				t.Errorf("addDependencyPinningRawResults() error = %v, wantError %v", err, test.wantError)
			}
		})
	}
}

func TestJsonScorecardRawResult_AddDangerousWorkflowRawResults(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet
		name      string
		input     *checker.DangerousWorkflowData
		wantError bool
	}{
		{
			name: "test_with_valid_data",
			input: &checker.DangerousWorkflowData{
				Workflows: []checker.DangerousWorkflow{
					{
						File: checker.File{
							Path:      "testPath",
							Offset:    0,
							EndOffset: 5,
							Snippet:   "testSnippet",
						},
						Type: checker.DangerousWorkflowScriptInjection,
						Job: &checker.WorkflowJob{
							Name: asPointer("testJob"),
							ID:   asPointer("testID"),
						},
					},
				},
			},
			wantError: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			r := &jsonScorecardRawResult{}
			err := r.addDangerousWorkflowRawResults(test.input)
			if (err != nil) != test.wantError {
				t.Errorf("addDangerousWorkflowRawResults() error = %v, wantError %v", err, test.wantError)
			}
		})
	}
}

func TestJsonScorecardRawResult_AddContributorsRawResults(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet
		name      string
		input     *checker.ContributorsData
		wantError bool
	}{
		{
			name: "test_with_valid_data",
			input: &checker.ContributorsData{
				Users: []clients.User{
					{
						Login:            "testLogin",
						NumContributions: 5,
						Organizations: []clients.User{
							{Login: "testOrg"},
						},
						Companies: []string{"testCompany"},
					},
				},
			},
			wantError: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			r := &jsonScorecardRawResult{}
			err := r.addContributorsRawResults(test.input)
			if (err != nil) != test.wantError {
				t.Errorf("addContributorsRawResults() error = %v, wantError %v", err, test.wantError)
			}
		})
	}
}

func TestJsonScorecardRawResult_AddSignedReleasesRawResults(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet
		name      string
		input     *checker.SignedReleasesData
		wantError bool
	}{
		{
			name: "test_with_valid_data",
			input: &checker.SignedReleasesData{
				Releases: []clients.Release{
					{
						TagName: "v1.0",
						URL:     "https://example.com/v1.0",
						Assets: []clients.ReleaseAsset{
							{
								Name: "asset1",
								URL:  "https://example.com/v1.0/asset1",
							},
						},
					},
				},
			},
			wantError: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			r := &jsonScorecardRawResult{}
			err := r.addSignedReleasesRawResults(test.input)
			if (err != nil) != test.wantError {
				t.Errorf("addSignedReleasesRawResults() error = %v, wantError %v", err, test.wantError)
			}
		})
	}
}

func TestJsonScorecardRawResult_AddMaintainedRawResults(t *testing.T) {
	t.Parallel()
	c := clients.RepoAssociationNone
	tests := []struct { //nolint:govet
		name      string
		input     *checker.MaintainedData
		wantError bool
	}{
		{
			name: "test_with_nil_archived_status",
			input: &checker.MaintainedData{
				CreatedAt: time.Now(),
				Issues:    []clients.Issue{},
			},
			wantError: false,
		},
		{
			name: "test_with_valid_archived_status",
			input: &checker.MaintainedData{
				CreatedAt: time.Now(),
				Issues: []clients.Issue{
					{
						URI: asPointer("testUrl"),
						Author: &clients.User{
							Login: "testLogin",
						},
						AuthorAssociation: &c,
						Comments: []clients.IssueComment{
							{
								Author: &clients.User{
									Login: "testLogin",
								},
								AuthorAssociation: &c,
							},
						},
					},
				},
			},
			wantError: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			r := &jsonScorecardRawResult{}
			err := r.addMaintainedRawResults(test.input)
			if (err != nil) != test.wantError {
				t.Errorf("addMaintainedRawResults() error = %v, wantError %v", err, test.wantError)
			}
		})
	}
}

func TestSetDefaultCommitData(t *testing.T) {
	// Define some test data.
	changesets := []checker.Changeset{
		{
			ReviewPlatform: "GitHub",
			RevisionID:     "abc123",
			Commits: []clients.Commit{
				{
					CommittedDate: time.Now(),
					Message:       "Initial commit",
					SHA:           "def456",
					Committer: clients.User{
						Login: "johndoe",
					},
				},
			},
			Reviews: []clients.Review{
				{
					State: "approved",
					Author: &clients.User{
						Login: "janedoe",
						IsBot: false,
					},
				},
			},
			Author: clients.User{
				Login: "johndoe",
			},
		},
	}

	// Create a new jsonScorecardRawResult.
	r := &jsonScorecardRawResult{}

	// Call setDefaultCommitData with the test data.
	err := r.setDefaultCommitData(changesets)
	if err != nil {
		t.Fatalf("setDefaultCommitData() returned an error: %v", err)
	}

	// Define the expected results.
	expected := []jsonDefaultBranchChangeset{
		{
			RevisionID:     "abc123",
			ReviewPlatform: "GitHub",
			Commits: []jsonCommit{
				{
					Committer: jsonUser{
						Login: "johndoe",
					},
					Message: "Initial commit",
					SHA:     "def456",
				},
			},
			Reviews: []jsonReview{
				{
					State: "approved",
					Reviewer: jsonUser{
						Login: "janedoe",
						IsBot: false,
					},
				},
			},
			Authors: []jsonUser{
				{
					Login: "johndoe",
				},
			},
		},
	}

	// Compare the actual results with the expected results.
	if diff := cmp.Diff(r.Results.DefaultBranchChangesets, expected); diff != "" {
		t.Errorf("setDefaultCommitData() mismatch (-want +got):\n%s", diff)
	}
}

func TestJsonScorecardRawResult_AddOssfBestPracticesRawResults(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet
		name      string
		input     *checker.CIIBestPracticesData
		wantError bool
	}{
		{
			name: "test_with_valid_badge",
			input: &checker.CIIBestPracticesData{
				Badge: clients.Gold,
			},
			wantError: false,
		},
		{
			name: "test_with_nil_badge",
			input: &checker.CIIBestPracticesData{
				Badge: clients.Silver,
			},
			wantError: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			r := &jsonScorecardRawResult{}
			err := r.addOssfBestPracticesRawResults(test.input)
			if (err != nil) != test.wantError {
				t.Errorf("addOssfBestPracticesRawResults() error = %v, wantError %v", err, test.wantError)
			}
			if r.Results.OssfBestPractices.Badge != test.input.Badge.String() {
				t.Errorf("addOssfBestPracticesRawResults() badge = %v, want %v", r.Results.OssfBestPractices.Badge, test.input.Badge.String()) //nolint:lll
			}
		})
	}
}

func TestJsonScorecardRawResult_AddCodeReviewRawResults(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet
		name      string
		input     *checker.CodeReviewData
		wantError bool
	}{
		{
			name: "test_with_valid_changesets",
			input: &checker.CodeReviewData{
				DefaultBranchChangesets: []checker.Changeset{
					{
						ReviewPlatform: "GitHub",
						RevisionID:     "123",
						Commits: []clients.Commit{
							{
								CommittedDate: time.Now(),
								Message:       "test commit",
								SHA:           "abc123",
								Committer: clients.User{
									Login: "testuser",
								},
							},
						},
						Reviews: []clients.Review{
							{
								State: "approved",
								Author: &clients.User{
									Login: "testuser",
								},
							},
						},
						Author: clients.User{
							Login: "testuser",
						},
					},
				},
			},
			wantError: false,
		},
		{
			name: "test_with_nil_changesets",
			input: &checker.CodeReviewData{
				DefaultBranchChangesets: nil,
			},
			wantError: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			r := &jsonScorecardRawResult{}
			err := r.addCodeReviewRawResults(test.input)
			if (err != nil) != test.wantError {
				t.Errorf("addCodeReviewRawResults() error = %v, wantError %v", err, test.wantError)
			}
			if len(r.Results.DefaultBranchChangesets) != len(test.input.DefaultBranchChangesets) {
				t.Errorf("addCodeReviewRawResults() changesets length = %v, want %v", len(r.Results.DefaultBranchChangesets), len(test.input.DefaultBranchChangesets)) //nolint:lll
			}
		})
	}
}

func TestAddCodeReviewRawResults(t *testing.T) {
	r := &jsonScorecardRawResult{}
	cr := &checker.CodeReviewData{
		DefaultBranchChangesets: []checker.Changeset{
			{
				RevisionID:     "abc123",
				ReviewPlatform: "github",
				Commits: []clients.Commit{
					{
						Committer: clients.User{
							Login: "johndoe",
						},
						Message: "Fix bug",
						SHA:     "def456",
					},
				},
				Reviews: []clients.Review{
					{
						State: "approved",
						Author: &clients.User{
							Login: "janedoe",
							IsBot: false,
						},
					},
				},
				Author: clients.User{
					Login: "johndoe",
				},
			},
		},
	}

	err := r.addCodeReviewRawResults(cr)
	if err != nil {
		t.Errorf("addCodeReviewRawResults returned an error: %v", err)
	}

	expected := []jsonDefaultBranchChangeset{
		{
			RevisionID:     "abc123",
			ReviewPlatform: "github",
			Commits: []jsonCommit{
				{
					Committer: jsonUser{
						Login: "johndoe",
					},
					Message: "Fix bug",
					SHA:     "def456",
				},
			},
			Reviews: []jsonReview{
				{
					State: "approved",
					Reviewer: jsonUser{
						Login: "janedoe",
						IsBot: false,
					},
				},
			},
			Authors: []jsonUser{
				{
					Login: "johndoe",
				},
			},
		},
	}

	if !reflect.DeepEqual(r.Results.DefaultBranchChangesets, expected) {
		t.Errorf("addCodeReviewRawResults did not produce the expected output. Got: %v, Expected: %v", r.Results.DefaultBranchChangesets, expected) //nolint:lll
	}
}

func TestAddLicenseRawResults(t *testing.T) {
	// Create a new jsonScorecardRawResult instance
	r := &jsonScorecardRawResult{}

	// Create a new LicenseData instance
	ld := &checker.LicenseData{
		LicenseFiles: []checker.LicenseFile{
			{
				File: checker.File{
					Path: "LICENSE",
				},
				LicenseInformation: checker.License{
					Name:        "MIT License",
					SpdxID:      "MIT",
					Attribution: checker.LicenseAttributionTypeOther,
					Approved:    true,
				},
			},
		},
	}

	// Call the addLicenseRawResults function
	err := r.addLicenseRawResults(ld)
	// Check if there was an error
	if err != nil {
		t.Errorf("addLicenseRawResults returned an error: %v", err)
	}

	// Check if the Licenses field was populated correctly
	expected := []jsonLicense{
		{
			License: jsonLicenseInfo{
				File:        "LICENSE",
				Name:        "MIT License",
				SpdxID:      "MIT",
				Attribution: "other",
				Approved:    "true",
			},
		},
	}

	if len(r.Results.Licenses) != len(expected) {
		t.Errorf("addLicenseRawResults did not populate the Licenses field correctly")
	}

	for i, license := range r.Results.Licenses {
		if license.License.File != expected[i].License.File {
			t.Errorf("addLicenseRawResults did not populate the Licenses field correctly")
		}
		if license.License.Name != expected[i].License.Name {
			t.Errorf("addLicenseRawResults did not populate the Licenses field correctly")
		}
		if license.License.SpdxID != expected[i].License.SpdxID {
			t.Errorf("addLicenseRawResults did not populate the Licenses field correctly")
		}
		if license.License.Attribution != expected[i].License.Attribution {
			t.Errorf("addLicenseRawResults did not populate the Licenses field correctly")
		}
		if license.License.Approved != expected[i].License.Approved {
			t.Errorf("addLicenseRawResults did not populate the Licenses field correctly")
		}
	}
}

func TestAddBinaryArtifactRawResults(t *testing.T) {
	r := &jsonScorecardRawResult{}
	ba := &checker.BinaryArtifactData{
		Files: []checker.File{
			{
				Path: "path/to/file1",
			},
			{
				Path: "path/to/file2",
			},
		},
	}

	err := r.addBinaryArtifactRawResults(ba)
	if err != nil {
		t.Errorf("addBinaryArtifactRawResults returned an error: %v", err)
	}

	expected := []jsonFile{
		{
			Path: "path/to/file1",
		},
		{
			Path: "path/to/file2",
		},
	}

	if len(r.Results.Binaries) != len(expected) {
		t.Errorf("addBinaryArtifactRawResults did not add the correct number of files. Expected %d, got %d", len(expected), len(r.Results.Binaries)) //nolint:lll
	}

	for i, file := range r.Results.Binaries {
		if file.Path != expected[i].Path {
			t.Errorf("addBinaryArtifactRawResults did not add the correct file. Expected %s, got %s", expected[i].Path, file.Path) //nolint:lll
		}
	}
}

func TestAddSecurityPolicyRawResults(t *testing.T) {
	r := &jsonScorecardRawResult{}
	sp := &checker.SecurityPolicyData{
		PolicyFiles: []checker.SecurityPolicyFile{
			{
				File: checker.File{
					Path:     "path/to/policy1",
					FileSize: 100,
				},
				Information: []checker.SecurityPolicyInformation{
					{
						InformationType: checker.SecurityPolicyInformationType("type1"),
					},
					{
						InformationType: checker.SecurityPolicyInformationType("type2"),
					},
				},
			},
			{
				File: checker.File{
					Path:     "path/to/policy2",
					FileSize: 200,
				},
				Information: []checker.SecurityPolicyInformation{
					{
						InformationType: checker.SecurityPolicyInformationType("type3"),
					},
				},
			},
		},
	}

	err := r.addSecurityPolicyRawResults(sp)
	if err != nil {
		t.Errorf("addSecurityPolicyRawResults returned an error: %v", err)
	}

	expected := []jsonSecurityFile{
		{
			Path:          "path/to/policy1",
			ContentLength: 100,
			Hits: []jsonSecurityPolicyHits{
				{
					Type: "type1",
				},
				{
					Type: "type2",
				},
			},
		},
		{
			Path:          "path/to/policy2",
			ContentLength: 200,
			Hits: []jsonSecurityPolicyHits{
				{
					Type: "type3",
				},
			},
		},
	}

	if len(r.Results.SecurityPolicies) != len(expected) {
		t.Errorf("addSecurityPolicyRawResults did not add the correct number of policies. Expected %d, got %d", len(expected), len(r.Results.SecurityPolicies)) //nolint:lll
	}

	for i, policy := range r.Results.SecurityPolicies {
		if policy.Path != expected[i].Path {
			t.Errorf("addSecurityPolicyRawResults did not add the correct policy. Expected %s, got %s", expected[i].Path, policy.Path) //nolint:lll
		}

		if policy.ContentLength != expected[i].ContentLength {
			t.Errorf("addSecurityPolicyRawResults did not add the correct content length. Expected %d, got %d", expected[i].ContentLength, policy.ContentLength) //nolint:lll
		}

		if len(policy.Hits) != len(expected[i].Hits) {
			t.Errorf("addSecurityPolicyRawResults did not add the correct number of hits. Expected %d, got %d", len(expected[i].Hits), len(policy.Hits)) //nolint:lll
		}

		for j, hit := range policy.Hits {
			if hit.Type != expected[i].Hits[j].Type {
				t.Errorf("addSecurityPolicyRawResults did not add the correct hit type. Expected %s, got %s", expected[i].Hits[j].Type, hit.Type) //nolint:lll
			}
		}
	}
}

func TestAddVulnerabilitiesRawResults(t *testing.T) {
	r := &jsonScorecardRawResult{}
	vd := &checker.VulnerabilitiesData{
		Vulnerabilities: []clients.Vulnerability{
			{
				ID: "CVE-2021-1234",
			},
			{
				ID: "CVE-2021-5678",
			},
		},
	}

	err := r.addVulnerbilitiesRawResults(vd)
	if err != nil {
		t.Errorf("addVulnerbilitiesRawResults returned an error: %v", err)
	}

	expected := []jsonDatabaseVulnerability{
		{
			ID: "CVE-2021-1234",
		},
		{
			ID: "CVE-2021-5678",
		},
	}

	if len(r.Results.DatabaseVulnerabilities) != len(expected) {
		t.Errorf("addVulnerbilitiesRawResults did not add the correct number of vulnerabilities. Expected %d, got %d", len(expected), len(r.Results.DatabaseVulnerabilities)) //nolint:lll
	}

	for i, vuln := range r.Results.DatabaseVulnerabilities {
		if vuln.ID != expected[i].ID {
			t.Errorf("addVulnerbilitiesRawResults did not add the correct vulnerability. Expected %s, got %s", expected[i].ID, vuln.ID) //nolint:lll
		}
	}
}

func TestAddFuzzingRawResults(t *testing.T) {
	r := &jsonScorecardRawResult{}
	fd := &checker.FuzzingData{
		Fuzzers: []checker.Tool{
			{
				Name: "fuzzer1",
				URL:  asPointer("https://example.com/fuzzer1"),
				Desc: asPointer("Fuzzer 1 description"),
				Files: []checker.File{
					{
						Path: "path/to/fuzzer1/file1",
					},
					{
						Path: "path/to/fuzzer1/file2",
					},
				},
			},
			{
				Name: "fuzzer2",
				URL:  asPointer("https://example.com/fuzzer2"),
				Desc: asPointer("Fuzzer 2 description"),
				Files: []checker.File{
					{
						Path: "path/to/fuzzer2/file1",
					},
				},
			},
		},
	}

	err := r.addFuzzingRawResults(fd)
	if err != nil {
		t.Errorf("addFuzzingRawResults returned an error: %v", err)
	}

	expectedFuzzers := []jsonTool{
		{
			Name: "fuzzer1",
			URL:  asPointer("https://example.com/fuzzer1"),
			Desc: asPointer("Fuzzer 1 description"),
			Files: []jsonFile{
				{
					Path: "path/to/fuzzer1/file1",
				},
				{
					Path: "path/to/fuzzer1/file2",
				},
			},
		},
		{
			Name: "fuzzer2",
			URL:  asPointer("https://example.com/fuzzer2"),
			Desc: asPointer("Fuzzer 2 description"),
			Files: []jsonFile{
				{
					Path: "path/to/fuzzer2/file1",
				},
			},
		},
	}

	if len(r.Results.Fuzzers) != len(expectedFuzzers) {
		t.Errorf("addFuzzingRawResults did not add the correct number of fuzzers. Expected %d, got %d", len(expectedFuzzers), len(r.Results.Fuzzers)) //nolint:lll
	}
	for i, fuzzer := range r.Results.Fuzzers {
		if fuzzer.Name != expectedFuzzers[i].Name {
			t.Errorf("addFuzzingRawResults did not add the correct fuzzer name. Expected %s, got %s", expectedFuzzers[i].Name, fuzzer.Name) //nolint:lll
		}
		if *fuzzer.URL != *expectedFuzzers[i].URL {
			t.Errorf("addFuzzingRawResults did not add the correct fuzzer URL. Expected %s, got %s", *expectedFuzzers[i].URL, *fuzzer.URL) //nolint:lll
		}
		if *fuzzer.Desc != *expectedFuzzers[i].Desc {
			t.Errorf("addFuzzingRawResults did not add the correct fuzzer description. Expected %s, got %s", *expectedFuzzers[i].Desc, *fuzzer.Desc) //nolint:lll
		}
		if len(fuzzer.Files) != len(expectedFuzzers[i].Files) {
			t.Errorf("addFuzzingRawResults did not add the correct number of files for fuzzer %s. Expected %d, got %d", fuzzer.Name, len(expectedFuzzers[i].Files), len(fuzzer.Files)) //nolint:lll
		}
		for j, file := range fuzzer.Files {
			if file.Path != expectedFuzzers[i].Files[j].Path {
				t.Errorf("addFuzzingRawResults did not add the correct file path for fuzzer %s. Expected %s, got %s", fuzzer.Name, expectedFuzzers[i].Files[j].Path, file.Path) //nolint:lll
			}
		}
	}
}

func TestJsonScorecardRawResult(t *testing.T) {
	// create a new instance of jsonScorecardRawResult
	r := &jsonScorecardRawResult{}

	// create some test data for each of the add*RawResults functions
	vd := &checker.VulnerabilitiesData{
		Vulnerabilities: []clients.Vulnerability{
			{ID: "CVE-2021-1234"},
			{ID: "CVE-2021-5678"},
		},
	}
	ba := &checker.BinaryArtifactData{
		Files: []checker.File{
			{Path: "binaries/foo"},
			{Path: "binaries/bar"},
		},
	}
	sp := &checker.SecurityPolicyData{
		PolicyFiles: []checker.SecurityPolicyFile{
			{
				File: checker.File{
					Path:     "policies/baz",
					FileSize: 1024,
				},
				Information: []checker.SecurityPolicyInformation{
					{
						InformationType: checker.SecurityPolicyInformationTypeEmail,
						InformationValue: checker.SecurityPolicyValueType{
							Match:      "match",
							LineNumber: 42,
							Offset:     0,
						},
					},
				},
			},
		},
	}
	fd := &checker.FuzzingData{
		Fuzzers: []checker.Tool{
			{
				Name: "fuzzer1",
				URL:  asPointer("https://example.com/fuzzer1"),
				Desc: asPointer("fuzzer1 description"),
				Files: []checker.File{
					{Path: "fuzzers/fuzzer1/foo"},
					{Path: "fuzzers/fuzzer1/bar"},
				},
			},
			{
				Name: "fuzzer2",
				URL:  asPointer("https://example.com/fuzzer2"),
				Desc: asPointer("fuzzer2 description"),
				Files: []checker.File{
					{Path: "fuzzers/fuzzer2/foo"},
					{Path: "fuzzers/fuzzer2/bar"},
				},
			},
		},
	}
	bp := &checker.BranchProtectionsData{
		Branches: []clients.BranchRef{
			{
				Name:      stringPtr("main"),
				Protected: boolPtr(true),
				BranchProtectionRule: clients.BranchProtectionRule{
					AllowDeletions:   boolPtr(true),
					AllowForcePushes: boolPtr(false),
					RequiredPullRequestReviews: clients.PullRequestReviewRule{
						RequireCodeOwnerReviews:      boolPtr(true),
						DismissStaleReviews:          boolPtr(true),
						RequiredApprovingReviewCount: intPtr(2),
					},
					RequireLinearHistory: boolPtr(true),
					EnforceAdmins:        boolPtr(true),
					CheckRules: clients.StatusChecksRule{
						RequiresStatusChecks: boolPtr(true),
						Contexts:             []string{"ci"},
						UpToDateBeforeMerge:  boolPtr(true),
					},
				},
			},
			{
				Name:      stringPtr("dev"),
				Protected: boolPtr(false),
			},
		},
	}

	// test addVulnerbilitiesRawResults
	err := r.addVulnerbilitiesRawResults(vd)
	if err != nil {
		t.Errorf("addVulnerbilitiesRawResults returned an error: %v", err)
	}
	expectedVulnerabilities := []jsonDatabaseVulnerability{
		{ID: "CVE-2021-1234"},
		{ID: "CVE-2021-5678"},
	}
	if cmp.Diff(r.Results.DatabaseVulnerabilities, expectedVulnerabilities) != "" {
		t.Errorf("addVulnerbilitiesRawResults did not produce the expected results %v", cmp.Diff(r.Results.DatabaseVulnerabilities, expectedVulnerabilities)) //nolint:lll
	}

	// test addBinaryArtifactRawResults
	err = r.addBinaryArtifactRawResults(ba)
	if err != nil {
		t.Errorf("addBinaryArtifactRawResults returned an error: %v", err)
	}
	expectedBinaries := []jsonFile{
		{Path: "binaries/foo"},
		{Path: "binaries/bar"},
	}
	if cmp.Diff(expectedBinaries, r.Results.Binaries) != "" {
		t.Errorf("addBinaryArtifactRawResults did not produce the expected results")
	}

	// test addSecurityPolicyRawResults
	err = r.addSecurityPolicyRawResults(sp)
	if err != nil {
		t.Errorf("addSecurityPolicyRawResults returned an error: %v", err)
	}
	expectedSecurityPolicies := []jsonSecurityFile{
		{
			Path:          "policies/baz",
			ContentLength: 1024,
			Hits: []jsonSecurityPolicyHits{
				{
					Type:       "emailAddress",
					Match:      "match",
					LineNumber: 42,
					Offset:     0,
				},
			},
		},
	}
	if cmp.Diff(expectedSecurityPolicies, r.Results.SecurityPolicies) != "" {
		t.Errorf("addSecurityPolicyRawResults did not produce the expected results %v", cmp.Diff(expectedSecurityPolicies, r.Results.SecurityPolicies)) //nolint:lll
	}

	// test addFuzzingRawResults
	err = r.addFuzzingRawResults(fd)
	if err != nil {
		t.Errorf("addFuzzingRawResults returned an error: %v", err)
	}
	expectedFuzzers := []jsonTool{
		{
			Name: "fuzzer1",
			URL:  asPointer("https://example.com/fuzzer1"),
			Desc: asPointer("fuzzer1 description"),
			Files: []jsonFile{
				{Path: "fuzzers/fuzzer1/foo"},
				{Path: "fuzzers/fuzzer1/bar"},
			},
		},
		{
			Name: "fuzzer2",
			URL:  asPointer("https://example.com/fuzzer1"),
			Desc: asPointer("fuzzer1 description"),
			Files: []jsonFile{
				{Path: "fuzzers/fuzzer2/foo"},
				{Path: "fuzzers/fuzzer2/bar"},
			},
		},
	}
	if cmp.Diff(expectedFuzzers, r.Results.Fuzzers, cmpopts.IgnoreFields(jsonTool{}, "URL", "Desc")) != "" {
		t.Errorf("addFuzzingRawResults did not produce the expected results %v", cmp.Diff(expectedFuzzers, r.Results.Fuzzers)) //nolint:lll
	}

	// test addBranchProtectionRawResults
	err = r.addBranchProtectionRawResults(bp)
	if err != nil {
		t.Errorf("addBranchProtectionRawResults returned an error: %v", err)
	}
}

func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int32) *int32 {
	return &i
}

//nolint:lll
func TestScorecardResult_AsRawJSON(t *testing.T) {
	type fields struct {
		Repo       RepoInfo
		Date       time.Time
		Scorecard  ScorecardInfo
		Checks     []checker.CheckResult
		RawResults checker.RawResults
		Metadata   []string
	}
	tests := []struct { //nolint:govet
		name       string
		fields     fields
		wantWriter string
		wantErr    bool
	}{
		{
			name: "happy path",
			fields: fields{
				Repo: RepoInfo{
					Name:      "bar",
					CommitSHA: "1234567890123456789012345678901234567890",
				},
			},
			wantWriter: `{"date":"0001-01-01","repo":{"name":"bar","commit":"1234567890123456789012345678901234567890"},"scorecard":{"version":"","commit":""},"metadata":null,"results":{"workflows":[],"permissions":{},"licenses":[],"issues":null,"openssfBestPracticesBadge":{"badge":"Unknown"},"databaseVulnerabilities":[],"binaries":[],"securityPolicies":[],"dependencyUpdateTools":[],"branchProtections":{"branches":[],"codeownersFiles":null},"Contributors":{"users":null},"defaultBranchChangesets":[],"archived":{"status":false},"createdAt":{"timestamp":"0001-01-01T00:00:00Z"},"fuzzers":[],"releases":[],"packages":[],"dependencyPinning":{"dependencies":null}}}
`, //nolint:lll
		},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			r := &ScorecardResult{
				Repo:       tt.fields.Repo,
				Date:       tt.fields.Date,
				Scorecard:  tt.fields.Scorecard,
				Checks:     tt.fields.Checks,
				RawResults: tt.fields.RawResults,
				Metadata:   tt.fields.Metadata,
			}
			writer := &bytes.Buffer{}
			err := r.AsRawJSON(writer)
			if (err != nil) != tt.wantErr {
				t.Errorf("AsRawJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotWriter := writer.String(); gotWriter != tt.wantWriter {
				t.Errorf(cmp.Diff(gotWriter, tt.wantWriter))
			}
		})
	}
}
