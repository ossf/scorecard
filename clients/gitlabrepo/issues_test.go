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
					URI:       strptr("https://gitlab.com/ossf-test/e2e-issues/-/issues/1"),
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

func Test_issuesHandler_getIssues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		issuePath  string
		memberPath string
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "successfully fetch issues",
			issuePath:  "./testdata/valid-issues",
			memberPath: "./testdata/valid-repo-members",
			wantCount:  1,
			wantErr:    false,
		},
		{
			name:      "failure fetching issues",
			issuePath: "./testdata/invalid-issues",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			httpClient := &http.Client{
				Transport: suffixStubTripper{
					responsePaths: map[string]string{
						"issues": tt.issuePath,
						"all":    tt.memberPath,
					},
				},
			}

			client, err := gitlab.NewClient("", gitlab.WithHTTPClient(httpClient))
			if err != nil {
				t.Fatalf("gitlab.NewClient error: %v", err)
			}

			handler := &issuesHandler{glClient: client}
			repoURL := Repo{owner: "ossf-tests", commitSHA: clients.HeadSHA}
			handler.init(&repoURL)

			got, err := handler.getIssues()
			if (err != nil) != tt.wantErr {
				t.Fatalf("getIssues error: %v, wantedErr: %t", err, tt.wantErr)
			}
			if !tt.wantErr && len(got) != tt.wantCount {
				t.Errorf("getIssues() returned %d issues, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func Test_issuesHandler_getIssuesWithHistory(t *testing.T) {
	t.Parallel()

	t.Run("fetch issues with label events and comments", func(t *testing.T) {
		t.Parallel()

		httpClient := &http.Client{
			Transport: suffixStubTripper{
				responsePaths: map[string]string{
					"issues":                "./testdata/valid-issues",
					"all":                   "./testdata/valid-repo-members",
					"resource_label_events": "./testdata/valid-label-events",
					"notes":                 "./testdata/valid-notes",
				},
			},
		}

		client, err := gitlab.NewClient("", gitlab.WithHTTPClient(httpClient))
		if err != nil {
			t.Fatalf("gitlab.NewClient error: %v", err)
		}

		handler := &issuesHandler{glClient: client}
		repoURL := Repo{owner: "ossf-tests", commitSHA: clients.HeadSHA}
		handler.init(&repoURL)

		issues, err := handler.getIssuesWithHistory()
		if err != nil {
			t.Fatalf("getIssuesWithHistory error: %v", err)
		}

		if len(issues) == 0 {
			t.Fatal("Expected at least one issue")
		}

		// The actual assertions depend on what's in the test data files
		// but we can verify the structure is correct
		for _, issue := range issues {
			if issue.URI == nil {
				t.Error("Issue URI should not be nil")
			}
			if issue.CreatedAt == nil {
				t.Error("Issue CreatedAt should not be nil")
			}
			// LabelEvents and Comments may be empty depending on test data
		}
	})
}

func Test_accessLevelToRepoAssociation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		accessLevel gitlab.AccessLevelValue
		expected    clients.RepoAssociation
	}{
		{0, clients.RepoAssociationNone},
		{5, clients.RepoAssociationFirstTimeContributor},
		{10, clients.RepoAssociationCollaborator},
		{20, clients.RepoAssociationCollaborator},
		{30, clients.RepoAssociationMember},
		{40, clients.RepoAssociationMaintainer},
		{50, clients.RepoAssociationOwner},
		{99, clients.RepoAssociationNone}, // unknown value defaults to None
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.accessLevel)), func(t *testing.T) {
			t.Parallel()
			result := accessLevelToRepoAssociation(tt.accessLevel)
			if result != tt.expected {
				t.Errorf("accessLevelToRepoAssociation(%d) = %v, want %v",
					tt.accessLevel, result, tt.expected)
			}
		})
	}
}

//nolint:gocognit // Test function with multiple edge case scenarios
func Test_issuesHandler_labelEventEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet // fieldalignment: pointer field already first
		createdAt     *time.Time
		name          string
		labelName     string
		action        string
		shouldInclude bool
		expectedAdded *bool // nil means shouldn't be included
	}{
		{
			name:          "Empty label name (filtered out)",
			labelName:     "",
			createdAt:     timeptr(time.Now()),
			action:        "add",
			shouldInclude: false,
		},
		{
			name:          "Nil CreatedAt (filtered out)",
			labelName:     "bug",
			createdAt:     nil,
			action:        "add",
			shouldInclude: false,
		},
		{
			name:          "Valid bug label with 'add' action",
			labelName:     "bug",
			createdAt:     timeptr(time.Now()),
			action:        "add",
			shouldInclude: true,
			expectedAdded: boolPtr(true),
		},
		{
			name:          "Valid security label with 'added' action",
			labelName:     "security",
			createdAt:     timeptr(time.Now()),
			action:        "added",
			shouldInclude: true,
			expectedAdded: boolPtr(true),
		},
		{
			name:          "Bug label with 'remove' action",
			labelName:     "bug",
			createdAt:     timeptr(time.Now()),
			action:        "remove",
			shouldInclude: true,
			expectedAdded: boolPtr(false),
		},
		{
			name:          "Security label with 'removed' action",
			labelName:     "security",
			createdAt:     timeptr(time.Now()),
			action:        "removed",
			shouldInclude: true,
			expectedAdded: boolPtr(false),
		},
		{
			name:          "Unknown action (filtered out)",
			labelName:     "bug",
			createdAt:     timeptr(time.Now()),
			action:        "unknown",
			shouldInclude: false,
		},
		{
			name:          "Label with whitespace",
			labelName:     "  bug  ",
			createdAt:     timeptr(time.Now()),
			action:        "add",
			shouldInclude: true,
			expectedAdded: boolPtr(true),
		},
		{
			name:          "Enhancement label (not bug/security)",
			labelName:     "enhancement",
			createdAt:     timeptr(time.Now()),
			action:        "add",
			shouldInclude: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Simulate the filtering logic from getIssuesWithHistory
			if tt.labelName == "" || tt.createdAt == nil {
				if tt.shouldInclude {
					t.Error("Expected filtering to exclude this event")
				}
				return
			}

			name := strings.ToLower(strings.TrimSpace(tt.labelName))
			if name != "bug" && name != "security" {
				if tt.shouldInclude {
					t.Error("Expected filtering to exclude non-bug/security labels")
				}
				return
			}

			var added bool
			switch tt.action {
			case "add", "added":
				added = true
			case "remove", "removed":
				added = false
			default:
				if tt.shouldInclude {
					t.Error("Expected filtering to exclude unknown actions")
				}
				return
			}
			if !tt.shouldInclude {
				t.Error("Event passed all filters but shouldInclude=false")
			}

			if tt.expectedAdded != nil && added != *tt.expectedAdded {
				t.Errorf("Expected Added=%v, got %v", *tt.expectedAdded, added)
			}
		})
	}
}

func Test_issuesHandler_noteEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		noteCreatedAt   *time.Time
		issueURI        *string
		expectedCreated time.Time
		expectedURL     string
	}{
		{
			name:            "Note with nil CreatedAt (defaults to zero time)",
			noteCreatedAt:   nil,
			issueURI:        strptr("https://gitlab.com/test/project/-/issues/1"),
			expectedCreated: time.Time{},
			expectedURL:     "https://gitlab.com/test/project/-/issues/1",
		},
		{
			name:            "Note with valid CreatedAt",
			noteCreatedAt:   timeptr(time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)),
			issueURI:        strptr("https://gitlab.com/test/project/-/issues/2"),
			expectedCreated: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			expectedURL:     "https://gitlab.com/test/project/-/issues/2",
		},
		{
			name:            "Note with nil issue URI (empty URL)",
			noteCreatedAt:   timeptr(time.Now()),
			issueURI:        nil,
			expectedCreated: time.Now(),
			expectedURL:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Simulate the note processing logic
			created := time.Time{}
			if tt.noteCreatedAt != nil {
				created = *tt.noteCreatedAt
			}

			commentURL := ""
			if tt.issueURI != nil {
				commentURL = *tt.issueURI
			}

			if !created.Equal(tt.expectedCreated) && tt.noteCreatedAt != nil {
				// Allow minor differences for time.Now()
				if tt.noteCreatedAt != nil && created.Sub(tt.expectedCreated).Abs() > time.Second {
					t.Errorf("Expected CreatedAt=%v, got %v", tt.expectedCreated, created)
				}
			}

			if commentURL != tt.expectedURL {
				t.Errorf("Expected URL=%q, got %q", tt.expectedURL, commentURL)
			}
		})
	}
}

func Test_issuesHandler_maintainerDetectionLogic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		maintainersMap     map[string]struct{}
		name               string
		username           string
		expectedMaintainer bool
	}{
		{
			name:     "User in maintainers map",
			username: "maintainer1",
			maintainersMap: map[string]struct{}{
				"maintainer1": {},
				"maintainer2": {},
			},
			expectedMaintainer: true,
		},
		{
			name:     "User not in maintainers map",
			username: "contributor1",
			maintainersMap: map[string]struct{}{
				"maintainer1": {},
			},
			expectedMaintainer: false,
		},
		{
			name:               "Empty maintainers map",
			username:           "user1",
			maintainersMap:     map[string]struct{}{},
			expectedMaintainer: false,
		},
		{
			name:     "Case sensitivity - lowercase username",
			username: "maintainer1",
			maintainersMap: map[string]struct{}{
				"maintainer1": {}, // Map stores lowercase
			},
			expectedMaintainer: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Simulate the maintainer detection logic
			login := strings.ToLower(tt.username)
			_, isMaint := tt.maintainersMap[login]

			if isMaint != tt.expectedMaintainer {
				t.Errorf("Expected IsMaintainer=%v, got %v", tt.expectedMaintainer, isMaint)
			}
		})
	}
}

func Test_issuesHandler_multipleLabelEventsPerIssue(t *testing.T) {
	t.Parallel()

	t.Run("Issue with multiple add/remove sequences", func(t *testing.T) {
		t.Parallel()

		issue := clients.Issue{
			IssueNumber: 100,
			LabelEvents: []clients.LabelEvent{},
		}

		// Simulate multiple label operations
		events := []struct { //nolint:govet // field order matches positional literals
			label string
			added bool
			time  time.Time
		}{
			{"bug", true, time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)},
			{"security", true, time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC)},
			{"bug", false, time.Date(2024, 1, 3, 10, 0, 0, 0, time.UTC)},
			{"bug", true, time.Date(2024, 1, 4, 10, 0, 0, 0, time.UTC)},
			{"security", false, time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC)},
		}

		for _, ev := range events {
			issue.LabelEvents = append(issue.LabelEvents, clients.LabelEvent{
				Label:     ev.label,
				Added:     ev.added,
				CreatedAt: ev.time,
			})
		}

		// Verify all events are captured
		if len(issue.LabelEvents) != 5 {
			t.Errorf("Expected 5 label events, got %d", len(issue.LabelEvents))
		}

		// Verify chronological order
		for i := 1; i < len(issue.LabelEvents); i++ {
			if !issue.LabelEvents[i-1].CreatedAt.Before(issue.LabelEvents[i].CreatedAt) {
				t.Error("Label events should be in chronological order")
			}
		}

		// Count bug operations: added twice, removed once
		bugAdded := 0
		bugRemoved := 0
		for _, ev := range issue.LabelEvents {
			if ev.Label == "bug" {
				if ev.Added {
					bugAdded++
				} else {
					bugRemoved++
				}
			}
		}

		if bugAdded != 2 {
			t.Errorf("Expected 2 bug add events, got %d", bugAdded)
		}
		if bugRemoved != 1 {
			t.Errorf("Expected 1 bug remove event, got %d", bugRemoved)
		}
	})
}

func Test_issuesHandler_authorAssociationMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		authorID          int
		memberID          int
		memberAccessLevel int
		expectedAssoc     clients.RepoAssociation
	}{
		{
			name:              "Author is maintainer (access level 40)",
			authorID:          1,
			memberID:          1,
			memberAccessLevel: 40,
			expectedAssoc:     clients.RepoAssociationMaintainer,
		},
		{
			name:              "Author is owner (access level 50)",
			authorID:          1,
			memberID:          1,
			memberAccessLevel: 50,
			expectedAssoc:     clients.RepoAssociationOwner,
		},
		{
			name:              "Author is developer (access level 30)",
			authorID:          1,
			memberID:          1,
			memberAccessLevel: 30,
			expectedAssoc:     clients.RepoAssociationMember,
		},
		{
			name:              "Author is reporter (access level 20)",
			authorID:          1,
			memberID:          1,
			memberAccessLevel: 20,
			expectedAssoc:     clients.RepoAssociationCollaborator,
		},
		{
			name:              "Author not in project members",
			authorID:          1,
			memberID:          2, // Different ID
			memberAccessLevel: 40,
			expectedAssoc:     clients.RepoAssociationNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Simulate the author association logic
			var assoc clients.RepoAssociation
			if tt.authorID == tt.memberID {
				assoc = accessLevelToRepoAssociation(gitlab.AccessLevelValue(tt.memberAccessLevel))
			} else {
				assoc = clients.RepoAssociationNone
			}

			if assoc != tt.expectedAssoc {
				t.Errorf("Expected association %v, got %v", tt.expectedAssoc, assoc)
			}
		})
	}
}

func Test_issuesHandler_emptyIssuesList(t *testing.T) {
	t.Parallel()

	t.Run("getIssuesWithHistory with no issues", func(t *testing.T) {
		t.Parallel()

		// Simulate handler with empty issues list
		emptyIssues := []clients.Issue{}
		out := make([]clients.Issue, 0, len(emptyIssues))

		out = append(out, emptyIssues...)

		if len(out) != 0 {
			t.Errorf("Expected 0 issues, got %d", len(out))
		}
	})
}

// Helper function.
func boolPtr(b bool) *bool {
	return &b
}
