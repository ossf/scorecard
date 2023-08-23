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
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/xanzy/go-gitlab"

	"github.com/ossf/scorecard/v4/clients"
)

// suffix may not be the best term, but maps the final part of a path to a response file.
// this is helpful when multiple API calls need to be made.
// e.g. a call to /foo/bar/some/endpoint would have "endpoint" as a suffix.
type suffixStubTripper struct {
	// key is suffix, value is response file.
	responsePaths map[string]string
}

func (s suffixStubTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	pathParts := strings.Split(r.URL.Path, "/")
	suffix := pathParts[len(pathParts)-1]
	f, err := os.Open(s.responsePaths[suffix])
	if err != nil {
		//nolint:wrapcheck
		return nil, err
	}
	return &http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Body:       f,
	}, nil
}

func associationptr(r clients.RepoAssociation) *clients.RepoAssociation {
	return &r
}

func timeptr(t time.Time) *time.Time {
	return &t
}

func Test_listIssues(t *testing.T) {
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
