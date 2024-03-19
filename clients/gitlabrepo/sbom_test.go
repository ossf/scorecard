// Copyright 2024 OpenSSF Scorecard Authors
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
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/xanzy/go-gitlab"

	"github.com/ossf/scorecard/v4/clients"
)

func Test_listSboms(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		issuePath  string
		memberPath string
		want       []clients.Issue
		wantErr    bool
	}{
		{
			name:       "issue with maintainer as author",
			issuePath:  "./testdata/valid-issues",
			memberPath: "./testdata/valid-repo-members",
			want: []clients.Issue{
				{
					URI:       strptr("131356518"),
					CreatedAt: timeptr(time.Date(2023, time.July, 26, 14, 22, 52, 0, time.UTC)),
					Author: &clients.User{
						ID: 1355794,
					},
					AuthorAssociation: associationptr(clients.RepoAssociationMaintainer),
				},
			},
			wantErr: false,
		},
		{
			name:      "failure fetching issues",
			issuePath: "./testdata/invalid-issues",
			want:      nil,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			httpClient := &http.Client{
				Transport: suffixStubTripper{
					responsePaths: map[string]string{
						"issues": tt.issuePath,  // corresponds to projects/<id>/issues
						"all":    tt.memberPath, // corresponds to projects/<id>/members/all
					},
				},
			}
			client, err := gitlab.NewClient("", gitlab.WithHTTPClient(httpClient))
			if err != nil {
				t.Fatalf("gitlab.NewClient error: %v", err)
			}
			handler := &issuesHandler{
				glClient: client,
			}

			repoURL := repoURL{
				owner:     "ossf-tests",
				commitSHA: clients.HeadSHA,
			}
			handler.init(&repoURL)
			got, err := handler.listIssues()
			if (err != nil) != tt.wantErr {
				t.Fatalf("listIssues error: %v, wantedErr: %t", err, tt.wantErr)
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("listIssues() = %v, want %v", got, cmp.Diff(got, tt.want))
			}
		})
	}
}
