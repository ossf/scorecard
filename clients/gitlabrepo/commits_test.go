// Copyright 2023 OpenSSF Scorecard Authors
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

package gitlabrepo

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/ossf/scorecard/v5/clients"
)

func TestParsingEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{
			name:     "Perfect Match Email Parser",
			email:    "john.doe@nowhere.com",
			expected: "john doe",
		},
		{
			name:     "Valid Email Not Formatted as expected",
			email:    "johndoe@nowhere.com",
			expected: "johndoe@nowhere com",
		},
		{
			name:     "Invalid email format",
			email:    "johndoe@nowherecom",
			expected: "johndoe@nowherecom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := parseEmailToName(tt.email)

			if tt.expected != result {
				t.Errorf("Parser didn't work as expected: %s != %s", tt.expected, result)
			}
		})
	}
}

func TestListRawCommits(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		commitsPath string
		commitDepth int
		want        int
		wantErr     bool
	}{
		{
			name:        "commits to non-default depth",
			commitsPath: "./testdata/valid-commits",
			commitDepth: 17,
			want:        17,
			wantErr:     false,
		},
		{
			name:        "commits to default depth",
			commitsPath: "./testdata/valid-commits",
			commitDepth: 30,
			want:        30,
			wantErr:     false,
		},
		{
			name:        "request more commits than exist in the repo",
			commitsPath: "./testdata/valid-commits",
			commitDepth: 60,
			want:        32,
			wantErr:     false,
		},
		{
			name:        "failure fetching commits",
			commitsPath: "./testdata/invalid-commits",
			commitDepth: 30,
			want:        0,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			httpClient := &http.Client{
				Transport: suffixStubTripper{
					responsePaths: map[string]string{
						"commits": tt.commitsPath, // corresponds to projects/<id>/repository/commits
					},
				},
			}
			client, err := gitlab.NewClient("", gitlab.WithHTTPClient(httpClient))
			if err != nil {
				t.Fatalf("gitlab.NewClient error: %v", err)
			}
			handler := &commitsHandler{
				glClient: client,
			}

			repoURL := Repo{
				owner:     "ossf-tests",
				commitSHA: clients.HeadSHA,
			}

			handler.init(&repoURL, tt.commitDepth)
			commits, err := handler.listRawCommits()
			if (err != nil) != tt.wantErr {
				t.Fatalf("listIssues error: %v, wantedErr: %t", err, tt.wantErr)
			}
			if !cmp.Equal(len(commits), tt.want) {
				t.Errorf("listCommits() = %v, want %v", len(commits), cmp.Diff(len(commits), tt.want))
			}
		})
	}
}
