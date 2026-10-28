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
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/ossf/scorecard/v5/clients"
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

			repoURL := Repo{
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

func Test_issuesHandler_AuthorAssociation_MultiIssue(t *testing.T) {
	t.Parallel()

	twoIssuesJSON := `[
  {
    "id": 1,
    "created_at": "2023-07-26T14:22:52.000Z",
    "author": { "id": 1355794, "username": "maintainer1" }
  },
  {
    "id": 2,
    "created_at": "2023-07-27T14:22:52.000Z",
    "author": { "id": 999999, "username": "external_user" }
  }
]`

	membersJSON := `[
  {
    "id": 1355794,
    "username": "maintainer1",
    "access_level": 40
  }
]`

	httpClient := &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			body := ""
			if strings.HasSuffix(r.URL.Path, "/issues") {
				body = twoIssuesJSON
			} else if strings.HasSuffix(r.URL.Path, "/members/all") {
				body = membersJSON
			}
			return &http.Response{
				Status:     "200 OK",
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
			}, nil
		}),
	}
	client, err := gitlab.NewClient("", gitlab.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("gitlab.NewClient error: %v", err)
	}

	handler := &issuesHandler{
		glClient: client,
	}
	repoURL := Repo{
		owner:     "ossf-tests",
		commitSHA: clients.HeadSHA,
	}
	handler.init(&repoURL)

	issues, err := handler.listIssues()
	if err != nil {
		t.Fatalf("listIssues error: %v", err)
	}

	if len(issues) != 2 {
		t.Fatalf("expected 2 issues, got %d", len(issues))
	}

	// Verify issue 1 is Maintainer
	if *issues[0].AuthorAssociation != clients.RepoAssociationMaintainer {
		t.Errorf("issue 0: got %v, want RepoAssociationMaintainer", *issues[0].AuthorAssociation)
	}

	// Verify issue 2 is None (no leak from issue 1)
	if *issues[1].AuthorAssociation != clients.RepoAssociationNone {
		t.Errorf("issue 1: got %v, want RepoAssociationNone", *issues[1].AuthorAssociation)
	}

	// Verify pointer aliasing does not occur: pointers must be distinct addresses
	if issues[0].AuthorAssociation == issues[1].AuthorAssociation {
		t.Errorf("expected distinct pointer addresses, got aliased pointer %p", issues[0].AuthorAssociation)
	}
}

func Test_issuesHandler_setup_project_members_errors(t *testing.T) {
	t.Parallel()

	issuesJSON := `[
  {
    "id": 1,
    "created_at": "2023-07-26T14:22:52.000Z",
    "author": { "id": 1355794, "username": "maintainer1" }
  }
]`

	tests := []struct {
		name          string
		memberHandler func(r *http.Request) (*http.Response, error)
		wantErrSubstr string
	}{
		{
			name: "network error with nil response",
			memberHandler: func(r *http.Request) (*http.Response, error) {
				return nil, errors.New("network connection timeout")
			},
			wantErrSubstr: "unable to find access tokens associated with the project id",
		},
		{
			name: "unauthorized response 401",
			memberHandler: func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					Status:     "401 Unauthorized",
					StatusCode: http.StatusUnauthorized,
					Body:       io.NopCloser(strings.NewReader(`{"message":"401 Unauthorized"}`)),
				}, nil
			},
			wantErrSubstr: "insufficient permissions to check issue author associations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			httpClient := &http.Client{
				Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
					if strings.HasSuffix(r.URL.Path, "/issues") {
						return &http.Response{
							Status:     "200 OK",
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(strings.NewReader(issuesJSON)),
						}, nil
					}
					return tt.memberHandler(r)
				}),
			}
			client, err := gitlab.NewClient("", gitlab.WithHTTPClient(httpClient))
			if err != nil {
				t.Fatalf("gitlab.NewClient error: %v", err)
			}

			handler := &issuesHandler{
				glClient: client,
			}
			repoURL := Repo{
				owner:     "ossf-tests",
				commitSHA: clients.HeadSHA,
			}
			handler.init(&repoURL)

			_, err = handler.listIssues()
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErrSubstr) {
				t.Errorf("error %q does not contain expected substring %q", err.Error(), tt.wantErrSubstr)
			}
		})
	}
}

type roundTripperFunc func(r *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
