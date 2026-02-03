// Copyright 2026 OpenSSF Scorecard Authors
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

package githubrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v82/github"
)

// setupMockGitHubServer creates a test HTTP server that mocks GitHub API responses.
func setupMockGitHubServer(t *testing.T, handlers map[string]http.HandlerFunc) (*httptest.Server, *github.Client) {
	t.Helper()

	mux := http.NewServeMux()
	for pattern, handler := range handlers {
		mux.HandleFunc(pattern, handler)
	}

	server := httptest.NewServer(mux)

	client := github.NewClient(nil)
	baseURL, err := url.Parse(server.URL + "/")
	if err != nil {
		t.Fatalf("Failed to parse server URL: %v", err)
	}
	client.BaseURL = baseURL
	client.UploadURL = baseURL

	return server, client
}

// mustEncode is a test helper that encodes JSON and ignores errors (acceptable in test mocks).
func mustEncode(w http.ResponseWriter, v interface{}) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// In test mocks, encoding errors are not recoverable anyway
		panic(fmt.Sprintf("test mock encoding failed: %v", err))
	}
}

// TestCheckCommits_WithMockedAPI tests commit activity detection with mocked GitHub API.
func TestCheckCommits_WithMockedAPI(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cutoff := now.Add(-180 * 24 * time.Hour)
	recentDate := now.Add(-30 * 24 * time.Hour)

	tests := []struct {
		name           string
		commits        []*github.RepositoryCommit
		expectedActive []string
	}{
		{
			name: "recent commits mark users active",
			commits: []*github.RepositoryCommit{
				{
					Author: &github.User{
						Login: github.Ptr("user1"),
					},
					Commit: &github.Commit{
						Author: &github.CommitAuthor{
							Date: &github.Timestamp{Time: recentDate},
						},
					},
				},
				{
					Committer: &github.User{
						Login: github.Ptr("user2"),
					},
					Commit: &github.Commit{
						Committer: &github.CommitAuthor{
							Date: &github.Timestamp{Time: recentDate},
						},
					},
				},
			},
			expectedActive: []string{"user1", "user2"},
		},
		{
			name: "old commits not returned by API",
			// Since the API uses Since parameter, old commits shouldn't be returned
			// Empty commits array simulates API filtering
			commits:        []*github.RepositoryCommit{},
			expectedActive: []string{},
		},
		{
			name: "nil author/committer handled",
			commits: []*github.RepositoryCommit{
				{
					Commit: &github.Commit{
						Author: &github.CommitAuthor{
							Date: &github.Timestamp{Time: recentDate},
						},
					},
				},
			},
			expectedActive: []string{},
		},
		{
			name: "multiple recent commits",
			// API Since parameter means all returned commits are recent
			commits: []*github.RepositoryCommit{
				{
					Author: &github.User{
						Login: github.Ptr("activeuser1"),
					},
					Commit: &github.Commit{
						Author: &github.CommitAuthor{
							Date: &github.Timestamp{Time: recentDate},
						},
					},
				},
				{
					Author: &github.User{
						Login: github.Ptr("activeuser2"),
					},
					Commit: &github.Commit{
						Author: &github.CommitAuthor{
							Date: &github.Timestamp{Time: recentDate},
						},
					},
				},
			},
			expectedActive: []string{"activeuser1", "activeuser2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server, client := setupMockGitHubServer(t, map[string]http.HandlerFunc{
				"/repos/test/repo/commits": func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					mustEncode(w, tt.commits)
				},
			})
			defer server.Close()

			handler := &maintainerHandler{
				ctx:       context.Background(),
				ghClient:  client,
				cutoff:    cutoff,
				elevated:  map[string]struct{}{},
				active:    make(map[string]bool),
				setupOnce: nil,
			}

			// Add all expected active users as elevated
			for _, login := range tt.expectedActive {
				handler.elevated[strings.ToLower(login)] = struct{}{}
			}
			// Also add some users that might appear in test names but won't be active
			handler.elevated["inactiveuser"] = struct{}{}
			handler.elevated["activeuser1"] = struct{}{}
			handler.elevated["activeuser2"] = struct{}{}

			err := handler.checkCommits("test", "repo")
			if err != nil {
				t.Fatalf("checkCommits() error = %v", err)
			}

			// Verify expected users are marked active
			for _, login := range tt.expectedActive {
				if !handler.active[strings.ToLower(login)] {
					t.Errorf("Expected %s to be marked active", login)
				}
			}

			// Verify activity map size matches expectations
			if len(handler.active) != len(tt.expectedActive) {
				t.Errorf("Expected %d active users, got %d", len(tt.expectedActive), len(handler.active))
			}
		})
	}
}

// TestCheckCommits_Pagination tests that pagination works correctly.
func TestCheckCommits_Pagination(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cutoff := now.Add(-180 * 24 * time.Hour)
	recentDate := now.Add(-30 * 24 * time.Hour)

	pageCount := 0

	// Create server with handler
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	// Add handler after server is created so we can reference server.URL
	mux.HandleFunc("/repos/test/repo/commits", func(w http.ResponseWriter, r *http.Request) {
		pageCount++

		w.Header().Set("Content-Type", "application/json")

		// First page: return data with Link header to next page
		if r.URL.Query().Get("page") == "" || r.URL.Query().Get("page") == "1" {
			// Set Link header for pagination - must be absolute URL
			nextURL := server.URL + "/repos/test/repo/commits?page=2"
			w.Header().Set("Link", fmt.Sprintf(`<%s>; rel="next"`, nextURL))
			commits := []*github.RepositoryCommit{
				{
					Author: &github.User{Login: github.Ptr("user1")},
					Commit: &github.Commit{
						Author: &github.CommitAuthor{Date: &github.Timestamp{Time: recentDate}},
					},
				},
			}
			mustEncode(w, commits)
			return
		}

		// Second page: return data with no next link
		commits := []*github.RepositoryCommit{
			{
				Author: &github.User{Login: github.Ptr("user2")},
				Commit: &github.Commit{
					Author: &github.CommitAuthor{Date: &github.Timestamp{Time: recentDate}},
				},
			},
		}
		mustEncode(w, commits)
	})

	// Create GitHub client
	client := github.NewClient(nil)
	baseURL, urlErr := url.Parse(server.URL + "/")
	if urlErr != nil {
		t.Fatalf("Failed to parse server URL: %v", urlErr)
	}
	client.BaseURL = baseURL
	client.UploadURL = baseURL

	handler := &maintainerHandler{
		ctx:       context.Background(),
		ghClient:  client,
		cutoff:    cutoff,
		elevated:  map[string]struct{}{"user1": {}, "user2": {}},
		active:    make(map[string]bool),
		setupOnce: nil,
	}

	err := handler.checkCommits("test", "repo")
	if err != nil {
		t.Fatalf("checkCommits() error = %v", err)
	}

	// Should have made 2 requests (2 pages)
	if pageCount != 2 {
		t.Errorf("Expected 2 API calls for pagination, got %d", pageCount)
	}

	// Both users should be active
	if !handler.active["user1"] || !handler.active["user2"] {
		t.Error("Expected both user1 and user2 to be active from paginated results")
	}
}

// TestCheckCommits_APIError tests error handling.
func TestCheckCommits_APIError(t *testing.T) {
	t.Parallel()

	server, client := setupMockGitHubServer(t, map[string]http.HandlerFunc{
		"/repos/test/repo/commits": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			mustEncode(w, map[string]string{
				"message": "Internal server error",
			})
		},
	})
	defer server.Close()

	handler := &maintainerHandler{
		ctx:       context.Background(),
		ghClient:  client,
		cutoff:    time.Now().Add(-180 * 24 * time.Hour),
		elevated:  map[string]struct{}{"user1": {}},
		active:    make(map[string]bool),
		setupOnce: nil,
	}

	err := handler.checkCommits("test", "repo")
	if err == nil {
		t.Error("Expected error when API returns 500, got nil")
	}
	if !strings.Contains(err.Error(), "ListCommits") {
		t.Errorf("Expected error message to mention ListCommits, got: %v", err)
	}
}

// TestCheckMergedPRs_WithMockedAPI tests PR activity detection.
func TestCheckMergedPRs_WithMockedAPI(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cutoff := now.Add(-180 * 24 * time.Hour)
	recentDate := now.Add(-30 * 24 * time.Hour)
	oldDate := now.Add(-200 * 24 * time.Hour)

	tests := []struct {
		name           string
		prs            []*github.PullRequest
		expectedActive []string
	}{
		{
			name: "recently merged PRs mark users active",
			prs: []*github.PullRequest{
				{
					Number:   github.Ptr(1),
					MergedAt: &github.Timestamp{Time: recentDate},
					User: &github.User{
						Login: github.Ptr("prauthor"),
					},
					MergedBy: &github.User{
						Login: github.Ptr("merger"),
					},
				},
			},
			expectedActive: []string{"prauthor", "merger"},
		},
		{
			name: "old merged PRs ignored",
			prs: []*github.PullRequest{
				{
					Number:   github.Ptr(2),
					MergedAt: &github.Timestamp{Time: oldDate},
					User: &github.User{
						Login: github.Ptr("oldauthor"),
					},
				},
			},
			expectedActive: []string{},
		},
		{
			name: "unmerged PRs ignored",
			prs: []*github.PullRequest{
				{
					Number:   github.Ptr(3),
					MergedAt: nil,
					User: &github.User{
						Login: github.Ptr("openpr"),
					},
				},
			},
			expectedActive: []string{},
		},
		{
			name: "nil user/merger handled",
			prs: []*github.PullRequest{
				{
					Number:   github.Ptr(4),
					MergedAt: &github.Timestamp{Time: recentDate},
				},
			},
			expectedActive: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server, client := setupMockGitHubServer(t, map[string]http.HandlerFunc{
				"/repos/test/repo/pulls": func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					mustEncode(w, tt.prs)
				},
			})
			defer server.Close()

			handler := &maintainerHandler{
				ctx:       context.Background(),
				ghClient:  client,
				cutoff:    cutoff,
				elevated:  map[string]struct{}{},
				active:    make(map[string]bool),
				setupOnce: nil,
			}

			// Add all expected users as elevated
			for _, login := range tt.expectedActive {
				handler.elevated[strings.ToLower(login)] = struct{}{}
			}
			handler.elevated["oldauthor"] = struct{}{}
			handler.elevated["openpr"] = struct{}{}

			err := handler.checkMergedPRs("test", "repo")
			if err != nil {
				t.Fatalf("checkMergedPRs() error = %v", err)
			}

			// Verify expected users are marked active
			for _, login := range tt.expectedActive {
				if !handler.active[strings.ToLower(login)] {
					t.Errorf("Expected %s to be marked active", login)
				}
			}

			if len(handler.active) != len(tt.expectedActive) {
				t.Errorf("Expected %d active users, got %d", len(tt.expectedActive), len(handler.active))
			}
		})
	}
}

// TestCheckMergedPRs_EarlyTermination tests that we stop pagination when encountering old PRs.
func TestCheckMergedPRs_EarlyTermination(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cutoff := now.Add(-180 * 24 * time.Hour)
	oldDate := now.Add(-200 * 24 * time.Hour)

	callCount := 0
	server, client := setupMockGitHubServer(t, map[string]http.HandlerFunc{
		"/repos/test/repo/pulls": func(w http.ResponseWriter, r *http.Request) {
			callCount++

			w.Header().Set("Content-Type", "application/json")

			// First page returns old PR - should stop here
			prs := []*github.PullRequest{
				{
					Number:   github.Ptr(1),
					MergedAt: &github.Timestamp{Time: oldDate},
					User:     &github.User{Login: github.Ptr("olduser")},
				},
			}

			// Set next page header (but should not be followed)
			w.Header().Set("Link", `<https://api.github.com/repos/test/repo/pulls?page=2>; rel="next"`)
			mustEncode(w, prs)
		},
	})
	defer server.Close()

	handler := &maintainerHandler{
		ctx:       context.Background(),
		ghClient:  client,
		cutoff:    cutoff,
		elevated:  map[string]struct{}{"olduser": {}},
		active:    make(map[string]bool),
		setupOnce: nil,
	}

	err := handler.checkMergedPRs("test", "repo")
	if err != nil {
		t.Fatalf("checkMergedPRs() error = %v", err)
	}

	// Should only make 1 call due to early termination
	if callCount != 1 {
		t.Errorf("Expected 1 API call (early termination), got %d", callCount)
	}

	// No users should be active
	if len(handler.active) != 0 {
		t.Errorf("Expected 0 active users, got %d", len(handler.active))
	}
}

// TestCheckReviews_WithMockedAPI tests PR review activity detection.
//
//nolint:gocognit // Test function with table-driven tests
func TestCheckReviews_WithMockedAPI(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cutoff := now.Add(-180 * 24 * time.Hour)
	recentDate := now.Add(-30 * 24 * time.Hour)
	oldDate := now.Add(-200 * 24 * time.Hour)

	tests := []struct {
		name           string
		prs            []*github.PullRequest
		reviews        map[int][]*github.PullRequestReview // PR number -> reviews
		expectedActive []string
	}{
		{
			name: "recent reviews mark users active",
			prs: []*github.PullRequest{
				{
					Number:    github.Ptr(1),
					UpdatedAt: &github.Timestamp{Time: recentDate},
				},
			},
			reviews: map[int][]*github.PullRequestReview{
				1: {
					{
						User:        &github.User{Login: github.Ptr("reviewer1")},
						SubmittedAt: &github.Timestamp{Time: recentDate},
					},
					{
						User:        &github.User{Login: github.Ptr("reviewer2")},
						SubmittedAt: &github.Timestamp{Time: recentDate},
					},
				},
			},
			expectedActive: []string{"reviewer1", "reviewer2"},
		},
		{
			name: "old reviews ignored",
			prs: []*github.PullRequest{
				{
					Number:    github.Ptr(2),
					UpdatedAt: &github.Timestamp{Time: recentDate},
				},
			},
			reviews: map[int][]*github.PullRequestReview{
				2: {
					{
						User:        &github.User{Login: github.Ptr("oldreviewer")},
						SubmittedAt: &github.Timestamp{Time: oldDate},
					},
				},
			},
			expectedActive: []string{},
		},
		{
			name: "early termination when all active",
			prs: []*github.PullRequest{
				{
					Number:    github.Ptr(3),
					UpdatedAt: &github.Timestamp{Time: recentDate},
				},
				{
					Number:    github.Ptr(4),
					UpdatedAt: &github.Timestamp{Time: recentDate},
				},
			},
			reviews: map[int][]*github.PullRequestReview{
				3: {
					{
						User:        &github.User{Login: github.Ptr("user1")},
						SubmittedAt: &github.Timestamp{Time: recentDate},
					},
				},
				4: {
					// This should not be fetched due to early termination
					{
						User:        &github.User{Login: github.Ptr("user2")},
						SubmittedAt: &github.Timestamp{Time: recentDate},
					},
				},
			},
			expectedActive: []string{"user1"}, // Only user1, not user2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reviewCallCount := 0
			server, client := setupMockGitHubServer(t, map[string]http.HandlerFunc{
				"/repos/test/repo/pulls": func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					mustEncode(w, tt.prs)
				},
				"/repos/test/repo/pulls/": func(w http.ResponseWriter, r *http.Request) {
					// Extract PR number from path
					parts := strings.Split(r.URL.Path, "/")
					prNum := 0
					for i, part := range parts {
						if part == "pulls" && i+1 < len(parts) {
							if _, err := fmt.Sscanf(parts[i+1], "%d", &prNum); err != nil {
								t.Logf("Failed to parse PR number: %v", err)
							}
							break
						}
					}

					reviewCallCount++

					w.Header().Set("Content-Type", "application/json")
					reviews := tt.reviews[prNum]
					mustEncode(w, reviews)
				},
			})
			defer server.Close()

			handler := &maintainerHandler{
				ctx:       context.Background(),
				ghClient:  client,
				cutoff:    cutoff,
				elevated:  map[string]struct{}{},
				active:    make(map[string]bool),
				setupOnce: nil,
			}

			// For early termination test, only add user1 as elevated
			if tt.name == "early termination when all active" {
				handler.elevated["user1"] = struct{}{}
			} else {
				// Add all expected users as elevated
				for _, login := range tt.expectedActive {
					handler.elevated[strings.ToLower(login)] = struct{}{}
				}
				handler.elevated["oldreviewer"] = struct{}{}
			}

			err := handler.checkReviews("test", "repo")
			if err != nil {
				t.Fatalf("checkReviews() error = %v", err)
			}

			// Verify expected users are marked active
			for _, login := range tt.expectedActive {
				if !handler.active[strings.ToLower(login)] {
					t.Errorf("Expected %s to be marked active", login)
				}
			}

			// For early termination test, verify we didn't fetch all PRs' reviews
			if tt.name == "early termination when all active" && reviewCallCount > 1 {
				t.Errorf("Expected early termination, but made %d review API calls", reviewCallCount)
			}
		})
	}
}

// TestCheckIssueComments_WithMockedAPI tests issue comment activity detection.
func TestCheckIssueComments_WithMockedAPI(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cutoff := now.Add(-180 * 24 * time.Hour)
	recentDate := now.Add(-30 * 24 * time.Hour)

	tests := []struct {
		name           string
		comments       []*github.IssueComment
		expectedActive []string
	}{
		{
			name: "recent comments mark users active",
			comments: []*github.IssueComment{
				{
					User:      &github.User{Login: github.Ptr("commenter1")},
					CreatedAt: &github.Timestamp{Time: recentDate},
				},
				{
					User:      &github.User{Login: github.Ptr("commenter2")},
					CreatedAt: &github.Timestamp{Time: recentDate},
				},
			},
			expectedActive: []string{"commenter1", "commenter2"},
		},
		{
			name: "nil user handled",
			comments: []*github.IssueComment{
				{
					CreatedAt: &github.Timestamp{Time: recentDate},
				},
			},
			expectedActive: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server, client := setupMockGitHubServer(t, map[string]http.HandlerFunc{
				"/repos/test/repo/issues/comments": func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					mustEncode(w, tt.comments)
				},
			})
			defer server.Close()

			handler := &maintainerHandler{
				ctx:       context.Background(),
				ghClient:  client,
				cutoff:    cutoff,
				elevated:  map[string]struct{}{},
				active:    make(map[string]bool),
				setupOnce: nil,
			}

			for _, login := range tt.expectedActive {
				handler.elevated[strings.ToLower(login)] = struct{}{}
			}

			err := handler.checkIssueComments("test", "repo")
			if err != nil {
				t.Fatalf("checkIssueComments() error = %v", err)
			}

			for _, login := range tt.expectedActive {
				if !handler.active[strings.ToLower(login)] {
					t.Errorf("Expected %s to be marked active", login)
				}
			}
		})
	}
}

// TestCheckReleases_WithMockedAPI tests release activity detection.
func TestCheckReleases_WithMockedAPI(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cutoff := now.Add(-180 * 24 * time.Hour)
	recentDate := now.Add(-30 * 24 * time.Hour)
	oldDate := now.Add(-200 * 24 * time.Hour)

	tests := []struct {
		name           string
		releases       []*github.RepositoryRelease
		expectedActive []string
	}{
		{
			name: "recent releases mark authors active",
			releases: []*github.RepositoryRelease{
				{
					Author:    &github.User{Login: github.Ptr("releaser1")},
					CreatedAt: &github.Timestamp{Time: recentDate},
				},
				{
					Author:    &github.User{Login: github.Ptr("releaser2")},
					CreatedAt: &github.Timestamp{Time: recentDate},
				},
			},
			expectedActive: []string{"releaser1", "releaser2"},
		},
		{
			name: "old releases ignored",
			releases: []*github.RepositoryRelease{
				{
					Author:    &github.User{Login: github.Ptr("oldreleaser")},
					CreatedAt: &github.Timestamp{Time: oldDate},
				},
			},
			expectedActive: []string{},
		},
		{
			name: "nil author/date handled",
			releases: []*github.RepositoryRelease{
				{
					CreatedAt: &github.Timestamp{Time: recentDate},
				},
				{
					Author: &github.User{Login: github.Ptr("user")},
				},
			},
			expectedActive: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server, client := setupMockGitHubServer(t, map[string]http.HandlerFunc{
				"/repos/test/repo/releases": func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					mustEncode(w, tt.releases)
				},
			})
			defer server.Close()

			handler := &maintainerHandler{
				ctx:       context.Background(),
				ghClient:  client,
				cutoff:    cutoff,
				elevated:  map[string]struct{}{},
				active:    make(map[string]bool),
				setupOnce: nil,
			}

			for _, login := range tt.expectedActive {
				handler.elevated[strings.ToLower(login)] = struct{}{}
			}
			handler.elevated["oldreleaser"] = struct{}{}

			err := handler.checkReleases("test", "repo")
			if err != nil {
				t.Fatalf("checkReleases() error = %v", err)
			}

			for _, login := range tt.expectedActive {
				if !handler.active[strings.ToLower(login)] {
					t.Errorf("Expected %s to be marked active", login)
				}
			}

			if len(handler.active) != len(tt.expectedActive) {
				t.Errorf("Expected %d active users, got %d", len(tt.expectedActive), len(handler.active))
			}
		})
	}
}

// TestCheckIssueActivity_WithMockedAPI tests issue creation/closing activity.
func TestCheckIssueActivity_WithMockedAPI(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cutoff := now.Add(-180 * 24 * time.Hour)
	recentDate := now.Add(-30 * 24 * time.Hour)
	oldDate := now.Add(-200 * 24 * time.Hour)

	tests := []struct {
		name           string
		issues         []*github.Issue
		expectedActive []string
	}{
		{
			name: "recent issue creation marks author active",
			issues: []*github.Issue{
				{
					Number:    github.Ptr(1),
					User:      &github.User{Login: github.Ptr("issueauthor")},
					CreatedAt: &github.Timestamp{Time: recentDate},
				},
			},
			expectedActive: []string{"issueauthor"},
		},
		{
			name: "recent issue closure marks closer active",
			issues: []*github.Issue{
				{
					Number:    github.Ptr(2),
					User:      &github.User{Login: github.Ptr("oldauthor")},
					CreatedAt: &github.Timestamp{Time: oldDate},
					ClosedAt:  &github.Timestamp{Time: recentDate},
					ClosedBy:  &github.User{Login: github.Ptr("closer")},
				},
			},
			expectedActive: []string{"closer"},
		},
		{
			name: "pull requests skipped",
			issues: []*github.Issue{
				{
					Number:           github.Ptr(3),
					User:             &github.User{Login: github.Ptr("prauthor")},
					CreatedAt:        &github.Timestamp{Time: recentDate},
					PullRequestLinks: &github.PullRequestLinks{},
				},
			},
			expectedActive: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server, client := setupMockGitHubServer(t, map[string]http.HandlerFunc{
				"/repos/test/repo/issues": func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					mustEncode(w, tt.issues)
				},
			})
			defer server.Close()

			handler := &maintainerHandler{
				ctx:       context.Background(),
				ghClient:  client,
				cutoff:    cutoff,
				elevated:  map[string]struct{}{},
				active:    make(map[string]bool),
				setupOnce: nil,
			}

			for _, login := range tt.expectedActive {
				handler.elevated[strings.ToLower(login)] = struct{}{}
			}
			handler.elevated["oldauthor"] = struct{}{}
			handler.elevated["prauthor"] = struct{}{}

			err := handler.checkIssueActivity("test", "repo")
			if err != nil {
				t.Fatalf("checkIssueActivity() error = %v", err)
			}

			for _, login := range tt.expectedActive {
				if !handler.active[strings.ToLower(login)] {
					t.Errorf("Expected %s to be marked active", login)
				}
			}

			if len(handler.active) != len(tt.expectedActive) {
				t.Errorf("Expected %d active users, got %d", len(tt.expectedActive), len(handler.active))
			}
		})
	}
}

// TestCheckPRActivity_WithMockedAPI tests PR creation and assignment activity.
func TestCheckPRActivity_WithMockedAPI(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cutoff := now.Add(-180 * 24 * time.Hour)
	recentDate := now.Add(-30 * 24 * time.Hour)
	oldDate := now.Add(-200 * 24 * time.Hour)

	tests := []struct {
		name           string
		prs            []*github.PullRequest
		expectedActive []string
	}{
		{
			name: "recent PR creation marks author active",
			prs: []*github.PullRequest{
				{
					Number:    github.Ptr(1),
					User:      &github.User{Login: github.Ptr("prauthor")},
					CreatedAt: &github.Timestamp{Time: recentDate},
					UpdatedAt: &github.Timestamp{Time: recentDate},
				},
			},
			expectedActive: []string{"prauthor"},
		},
		{
			name: "assignees marked active",
			prs: []*github.PullRequest{
				{
					Number:    github.Ptr(2),
					UpdatedAt: &github.Timestamp{Time: recentDate},
					Assignee:  &github.User{Login: github.Ptr("assignee1")},
					Assignees: []*github.User{
						{Login: github.Ptr("assignee2")},
						{Login: github.Ptr("assignee3")},
					},
				},
			},
			expectedActive: []string{"assignee1", "assignee2", "assignee3"},
		},
		{
			name: "requested reviewers marked active",
			prs: []*github.PullRequest{
				{
					Number:    github.Ptr(3),
					UpdatedAt: &github.Timestamp{Time: recentDate},
					RequestedReviewers: []*github.User{
						{Login: github.Ptr("reviewer1")},
						{Login: github.Ptr("reviewer2")},
					},
				},
			},
			expectedActive: []string{"reviewer1", "reviewer2"},
		},
		{
			name: "old PRs trigger early termination",
			prs: []*github.PullRequest{
				{
					Number:    github.Ptr(4),
					UpdatedAt: &github.Timestamp{Time: oldDate},
					User:      &github.User{Login: github.Ptr("olduser")},
				},
			},
			expectedActive: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server, client := setupMockGitHubServer(t, map[string]http.HandlerFunc{
				"/repos/test/repo/pulls": func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					mustEncode(w, tt.prs)
				},
			})
			defer server.Close()

			handler := &maintainerHandler{
				ctx:       context.Background(),
				ghClient:  client,
				cutoff:    cutoff,
				elevated:  map[string]struct{}{},
				active:    make(map[string]bool),
				setupOnce: nil,
			}

			for _, login := range tt.expectedActive {
				handler.elevated[strings.ToLower(login)] = struct{}{}
			}
			handler.elevated["olduser"] = struct{}{}

			err := handler.checkPRActivity("test", "repo")
			if err != nil {
				t.Fatalf("checkPRActivity() error = %v", err)
			}

			for _, login := range tt.expectedActive {
				if !handler.active[strings.ToLower(login)] {
					t.Errorf("Expected %s to be marked active", login)
				}
			}
		})
	}
}

// TestCheckWorkflowActivity_WithMockedAPI tests GitHub Actions workflow activity.
func TestCheckWorkflowActivity_WithMockedAPI(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cutoff := now.Add(-180 * 24 * time.Hour)
	recentDate := now.Add(-30 * 24 * time.Hour)
	oldDate := now.Add(-200 * 24 * time.Hour)

	tests := []struct {
		name           string
		runs           []*github.WorkflowRun
		expectedActive []string
	}{
		{
			name: "recent workflow runs mark actor active",
			runs: []*github.WorkflowRun{
				{
					ID:        github.Ptr(int64(1)),
					CreatedAt: &github.Timestamp{Time: recentDate},
					Actor:     &github.User{Login: github.Ptr("workflowuser")},
				},
			},
			expectedActive: []string{"workflowuser"},
		},
		{
			name: "old workflow runs skipped",
			runs: []*github.WorkflowRun{
				{
					ID:        github.Ptr(int64(2)),
					CreatedAt: &github.Timestamp{Time: oldDate},
					Actor:     &github.User{Login: github.Ptr("oldactor")},
				},
			},
			expectedActive: []string{},
		},
		{
			name: "nil actor handled",
			runs: []*github.WorkflowRun{
				{
					ID:        github.Ptr(int64(3)),
					CreatedAt: &github.Timestamp{Time: recentDate},
				},
			},
			expectedActive: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server, client := setupMockGitHubServer(t, map[string]http.HandlerFunc{
				"/repos/test/repo/actions/runs": func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					response := struct {
						WorkflowRuns []*github.WorkflowRun `json:"workflow_runs"`
					}{
						WorkflowRuns: tt.runs,
					}
					mustEncode(w, response)
				},
			})
			defer server.Close()

			handler := &maintainerHandler{
				ctx:       context.Background(),
				ghClient:  client,
				cutoff:    cutoff,
				elevated:  map[string]struct{}{},
				active:    make(map[string]bool),
				setupOnce: nil,
			}

			for _, login := range tt.expectedActive {
				handler.elevated[strings.ToLower(login)] = struct{}{}
			}
			handler.elevated["oldactor"] = struct{}{}

			err := handler.checkWorkflowActivity("test", "repo")
			if err != nil {
				t.Fatalf("checkWorkflowActivity() error = %v", err)
			}

			for _, login := range tt.expectedActive {
				if !handler.active[strings.ToLower(login)] {
					t.Errorf("Expected %s to be marked active", login)
				}
			}
		})
	}
}

// TestCheckCommitComments_WithMockedAPI tests commit comment activity.
// This method calls ListCommits, then ListCommitComments for each commit.
func TestCheckCommitComments_WithMockedAPI(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cutoff := now.Add(-180 * 24 * time.Hour)
	recentDate := now.Add(-30 * 24 * time.Hour)

	server, client := setupMockGitHubServer(t, map[string]http.HandlerFunc{
		"/repos/test/repo/commits": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			commits := []*github.RepositoryCommit{
				{SHA: github.Ptr("abc123")},
			}
			mustEncode(w, commits)
		},
		"/repos/test/repo/commits/abc123/comments": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			comments := []*github.RepositoryComment{
				{
					User:      &github.User{Login: github.Ptr("commenter1")},
					CreatedAt: &github.Timestamp{Time: recentDate},
				},
			}
			mustEncode(w, comments)
		},
	})
	defer server.Close()

	handler := &maintainerHandler{
		ctx:       context.Background(),
		ghClient:  client,
		cutoff:    cutoff,
		elevated:  map[string]struct{}{"commenter1": {}},
		active:    make(map[string]bool),
		setupOnce: nil,
	}

	err := handler.checkCommitComments("test", "repo")
	if err != nil {
		t.Fatalf("checkCommitComments() error = %v", err)
	}

	if !handler.active["commenter1"] {
		t.Errorf("Expected commenter1 to be marked active")
	}
}

// TestCheckMilestoneActivity_WithMockedAPI tests milestone creation activity.
// This method calls ListMilestones, then ListByRepo for issues, then ListIssueTimeline.
func TestCheckMilestoneActivity_WithMockedAPI(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cutoff := now.Add(-180 * 24 * time.Hour)
	recentDate := now.Add(-30 * 24 * time.Hour)

	server, client := setupMockGitHubServer(t, map[string]http.HandlerFunc{
		"/repos/test/repo/milestones": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			milestones := []*github.Milestone{
				{
					Number:    github.Ptr(1),
					Creator:   &github.User{Login: github.Ptr("creator1")},
					CreatedAt: &github.Timestamp{Time: recentDate},
					UpdatedAt: &github.Timestamp{Time: recentDate},
				},
			}
			mustEncode(w, milestones)
		},
		"/repos/test/repo/issues": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			// Return empty list to avoid needing timeline endpoints
			mustEncode(w, []*github.Issue{})
		},
	})
	defer server.Close()

	handler := &maintainerHandler{
		ctx:       context.Background(),
		ghClient:  client,
		cutoff:    cutoff,
		elevated:  map[string]struct{}{"creator1": {}},
		active:    make(map[string]bool),
		setupOnce: nil,
	}

	err := handler.checkMilestoneActivity("test", "repo")
	if err != nil {
		t.Fatalf("checkMilestoneActivity() error = %v", err)
	}

	if !handler.active["creator1"] {
		t.Errorf("Expected creator1 to be marked active")
	}
}

// TestCheckIssueReactions_WithMockedAPI tests issue reaction tracking.
// Note: Reactions lack timestamps, so we can only track who reacted.
func TestCheckIssueReactions_WithMockedAPI(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cutoff := now.Add(-180 * 24 * time.Hour)
	recentDate := now.Add(-30 * 24 * time.Hour)

	server, client := setupMockGitHubServer(t, map[string]http.HandlerFunc{
		"/repos/test/repo/issues": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			issues := []*github.Issue{
				{
					Number:    github.Ptr(1),
					UpdatedAt: &github.Timestamp{Time: recentDate},
				},
			}
			mustEncode(w, issues)
		},
		"/repos/test/repo/issues/1/reactions": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			reactions := []*github.Reaction{
				{User: &github.User{Login: github.Ptr("reactor1")}},
			}
			mustEncode(w, reactions)
		},
	})
	defer server.Close()

	handler := &maintainerHandler{
		ctx:       context.Background(),
		ghClient:  client,
		cutoff:    cutoff,
		elevated:  map[string]struct{}{"reactor1": {}},
		active:    make(map[string]bool),
		setupOnce: nil,
	}

	err := handler.checkIssueReactions("test", "repo")
	if err != nil {
		t.Fatalf("checkIssueReactions() error = %v", err)
	}

	if !handler.active["reactor1"] {
		t.Errorf("Expected reactor1 to be marked active")
	}
}

// TestCheckRepoEvents_WithMockedAPI tests repository event tracking.
func TestCheckRepoEvents_WithMockedAPI(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cutoff := now.Add(-180 * 24 * time.Hour)
	recentDate := now.Add(-30 * 24 * time.Hour)

	server, client := setupMockGitHubServer(t, map[string]http.HandlerFunc{
		"/repos/test/repo/events": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			events := []*github.Event{
				{
					Type:      github.Ptr("PushEvent"),
					CreatedAt: &github.Timestamp{Time: recentDate},
					Actor:     &github.User{Login: github.Ptr("pusher1")},
				},
			}
			mustEncode(w, events)
		},
	})
	defer server.Close()

	handler := &maintainerHandler{
		ctx:       context.Background(),
		ghClient:  client,
		cutoff:    cutoff,
		elevated:  map[string]struct{}{"pusher1": {}},
		active:    make(map[string]bool),
		setupOnce: nil,
	}

	err := handler.checkRepoEvents("test", "repo")
	if err != nil {
		t.Fatalf("checkRepoEvents() error = %v", err)
	}

	if !handler.active["pusher1"] {
		t.Errorf("Expected pusher1 to be marked active")
	}
}

// TestCheckDiscussions_WithMockedAPI tests GitHub Discussions tracking.
func TestCheckDiscussions_WithMockedAPI(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cutoff := now.Add(-180 * 24 * time.Hour)
	recentDate := now.Add(-30 * 24 * time.Hour)

	server, client := setupMockGitHubServer(t, map[string]http.HandlerFunc{
		"/repos/test/repo/events": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			events := []*github.Event{
				{
					Type:      github.Ptr("DiscussionEvent"),
					CreatedAt: &github.Timestamp{Time: recentDate},
					Actor:     &github.User{Login: github.Ptr("discusser1")},
				},
			}
			mustEncode(w, events)
		},
	})
	defer server.Close()

	handler := &maintainerHandler{
		ctx:       context.Background(),
		ghClient:  client,
		cutoff:    cutoff,
		elevated:  map[string]struct{}{"discusser1": {}},
		active:    make(map[string]bool),
		setupOnce: nil,
	}

	err := handler.checkDiscussions("test", "repo")
	if err != nil {
		t.Fatalf("checkDiscussions() error = %v", err)
	}

	if !handler.active["discusser1"] {
		t.Errorf("Expected discusser1 to be marked active")
	}
}

// TestCheckLabelActivity_WithMockedAPI tests label management tracking.
func TestCheckLabelActivity_WithMockedAPI(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cutoff := now.Add(-180 * 24 * time.Hour)
	recentDate := now.Add(-30 * 24 * time.Hour)

	server, client := setupMockGitHubServer(t, map[string]http.HandlerFunc{
		"/repos/test/repo/issues": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			issues := []*github.Issue{
				{
					Number:    github.Ptr(1),
					UpdatedAt: &github.Timestamp{Time: recentDate},
				},
			}
			mustEncode(w, issues)
		},
		"/repos/test/repo/issues/1/timeline": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			timeline := []*github.Timeline{
				{
					Event:     github.Ptr("labeled"),
					CreatedAt: &github.Timestamp{Time: recentDate},
					Actor:     &github.User{Login: github.Ptr("labeler1")},
				},
			}
			mustEncode(w, timeline)
		},
	})
	defer server.Close()

	handler := &maintainerHandler{
		ctx:       context.Background(),
		ghClient:  client,
		cutoff:    cutoff,
		elevated:  map[string]struct{}{"labeler1": {}},
		active:    make(map[string]bool),
		setupOnce: nil,
	}

	err := handler.checkLabelActivity("test", "repo")
	if err != nil {
		t.Fatalf("checkLabelActivity() error = %v", err)
	}

	if !handler.active["labeler1"] {
		t.Errorf("Expected labeler1 to be marked active")
	}
}

// TestCheckPRReactions_WithMockedAPI tests PR reaction tracking.
func TestCheckPRReactions_WithMockedAPI(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cutoff := now.Add(-180 * 24 * time.Hour)
	recentDate := now.Add(-30 * 24 * time.Hour)

	server, client := setupMockGitHubServer(t, map[string]http.HandlerFunc{
		"/repos/test/repo/pulls": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			prs := []*github.PullRequest{
				{
					Number:    github.Ptr(1),
					UpdatedAt: &github.Timestamp{Time: recentDate},
				},
			}
			mustEncode(w, prs)
		},
		"/repos/test/repo/issues/1/reactions": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			reactions := []*github.Reaction{
				{User: &github.User{Login: github.Ptr("prreactor1")}},
			}
			mustEncode(w, reactions)
		},
	})
	defer server.Close()

	handler := &maintainerHandler{
		ctx:       context.Background(),
		ghClient:  client,
		cutoff:    cutoff,
		elevated:  map[string]struct{}{"prreactor1": {}},
		active:    make(map[string]bool),
		setupOnce: nil,
	}

	err := handler.checkPRReactions("test", "repo")
	if err != nil {
		t.Fatalf("checkPRReactions() error = %v", err)
	}

	if !handler.active["prreactor1"] {
		t.Errorf("Expected prreactor1 to be marked active")
	}
}

// TestCheckCommitCommentReactions_WithMockedAPI tests commit comment reaction tracking.
func TestCheckCommitCommentReactions_WithMockedAPI(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cutoff := now.Add(-180 * 24 * time.Hour)
	recentDate := now.Add(-30 * 24 * time.Hour)

	server, client := setupMockGitHubServer(t, map[string]http.HandlerFunc{
		"/repos/test/repo/commits": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			commits := []*github.RepositoryCommit{
				{SHA: github.Ptr("abc123")},
			}
			mustEncode(w, commits)
		},
		"/repos/test/repo/comments": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			comments := []*github.RepositoryComment{
				{
					ID:        github.Ptr(int64(1)),
					CreatedAt: &github.Timestamp{Time: recentDate},
					CommitID:  github.Ptr("abc123"),
				},
			}
			mustEncode(w, comments)
		},
		"/repos/test/repo/comments/1/reactions": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			reactions := []*github.Reaction{
				{User: &github.User{Login: github.Ptr("commitreactor1")}},
			}
			mustEncode(w, reactions)
		},
	})
	defer server.Close()

	handler := &maintainerHandler{
		ctx:       context.Background(),
		ghClient:  client,
		cutoff:    cutoff,
		elevated:  map[string]struct{}{"commitreactor1": {}},
		active:    make(map[string]bool),
		setupOnce: nil,
	}

	err := handler.checkCommitCommentReactions("test", "repo")
	if err != nil {
		t.Fatalf("checkCommitCommentReactions() error = %v", err)
	}

	if !handler.active["commitreactor1"] {
		t.Errorf("Expected commitreactor1 to be marked active")
	}
}

// TestCheckProjectActivity_WithMockedAPI tests GitHub Projects activity tracking.
// Note: This API is deprecated and always returns nil, so test just ensures no error.
func TestCheckProjectActivity_WithMockedAPI(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cutoff := now.Add(-180 * 24 * time.Hour)

	server, client := setupMockGitHubServer(t, map[string]http.HandlerFunc{
		"/repos/test/repo/projects": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			// Return empty array - API deprecated
			mustEncode(w, []interface{}{})
		},
	})
	defer server.Close()

	handler := &maintainerHandler{
		ctx:       context.Background(),
		ghClient:  client,
		cutoff:    cutoff,
		elevated:  map[string]struct{}{},
		active:    make(map[string]bool),
		setupOnce: nil,
	}

	err := handler.checkProjectActivity("test", "repo")
	if err != nil {
		t.Fatalf("checkProjectActivity() error = %v", err)
	}

	// This API is deprecated and returns early, so no users should be marked active
}

// TestCheckSecurityActivity_WithMockedAPI tests Dependabot alert activity tracking.
func TestCheckSecurityActivity_WithMockedAPI(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cutoff := now.Add(-180 * 24 * time.Hour)
	recentDate := now.Add(-30 * 24 * time.Hour)

	server, client := setupMockGitHubServer(t, map[string]http.HandlerFunc{
		"/repos/test/repo/dependabot/alerts": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			alerts := []*github.DependabotAlert{
				{
					DismissedAt: &github.Timestamp{Time: recentDate},
					DismissedBy: &github.User{Login: github.Ptr("securityuser1")},
				},
			}
			mustEncode(w, alerts)
		},
	})
	defer server.Close()

	handler := &maintainerHandler{
		ctx:       context.Background(),
		ghClient:  client,
		cutoff:    cutoff,
		elevated:  map[string]struct{}{"securityuser1": {}},
		active:    make(map[string]bool),
		setupOnce: nil,
	}

	err := handler.checkSecurityActivity("test", "repo")
	if err != nil {
		t.Fatalf("checkSecurityActivity() error = %v", err)
	}

	if !handler.active["securityuser1"] {
		t.Errorf("Expected securityuser1 to be marked active")
	}
}

// TestCheckCommentReactions_WithMockedAPI tests issue comment reaction tracking.
func TestCheckCommentReactions_WithMockedAPI(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cutoff := now.Add(-180 * 24 * time.Hour)
	recentDate := now.Add(-30 * 24 * time.Hour)

	server, client := setupMockGitHubServer(t, map[string]http.HandlerFunc{
		"/repos/test/repo/issues/comments": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			comments := []*github.IssueComment{
				{
					ID:        github.Ptr(int64(1)),
					UpdatedAt: &github.Timestamp{Time: recentDate},
				},
			}
			mustEncode(w, comments)
		},
		"/repos/test/repo/issues/comments/1/reactions": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			reactions := []*github.Reaction{
				{User: &github.User{Login: github.Ptr("commentreactor1")}},
			}
			mustEncode(w, reactions)
		},
	})
	defer server.Close()

	handler := &maintainerHandler{
		ctx:       context.Background(),
		ghClient:  client,
		cutoff:    cutoff,
		elevated:  map[string]struct{}{"commentreactor1": {}},
		active:    make(map[string]bool),
		setupOnce: nil,
	}

	err := handler.checkCommentReactions("test", "repo")
	if err != nil {
		t.Fatalf("checkCommentReactions() error = %v", err)
	}

	if !handler.active["commentreactor1"] {
		t.Errorf("Expected commentreactor1 to be marked active")
	}
}

// TestSecurityBoundary_NonElevatedUsersIgnored tests that users not in the elevated set
// are never marked as active, regardless of their activity.
//
//nolint:gocognit // Test function with multiple scenarios
func TestSecurityBoundary_NonElevatedUsersIgnored(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cutoff := now.Add(-180 * 24 * time.Hour)
	recentDate := now.Add(-30 * 24 * time.Hour)

	tests := []struct {
		name          string
		setupHandler  func(*testing.T) (*httptest.Server, *github.Client)
		checkFunc     func(*maintainerHandler, string, string) error
		elevatedUsers []string
		activeUsers   []string // users with activity in the mock
	}{
		{
			name: "checkCommits ignores non-elevated commit authors",
			setupHandler: func(t *testing.T) (*httptest.Server, *github.Client) {
				t.Helper()
				return setupMockGitHubServer(t, map[string]http.HandlerFunc{
					"/repos/test/repo/commits": func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("Content-Type", "application/json")
						commits := []*github.RepositoryCommit{
							{
								Author: &github.User{Login: github.Ptr("elevated-user")},
								Commit: &github.Commit{
									Author: &github.CommitAuthor{Date: &github.Timestamp{Time: recentDate}},
								},
							},
							{
								Author: &github.User{Login: github.Ptr("non-elevated-user")},
								Commit: &github.Commit{
									Author: &github.CommitAuthor{Date: &github.Timestamp{Time: recentDate}},
								},
							},
						}
						mustEncode(w, commits)
					},
				})
			},
			checkFunc: func(h *maintainerHandler, owner, repo string) error {
				return h.checkCommits(owner, repo)
			},
			elevatedUsers: []string{"elevated-user"},
			activeUsers:   []string{"elevated-user", "non-elevated-user"},
		},
		{
			name: "checkMergedPRs ignores non-elevated PR authors",
			setupHandler: func(t *testing.T) (*httptest.Server, *github.Client) {
				t.Helper()
				return setupMockGitHubServer(t, map[string]http.HandlerFunc{
					"/repos/test/repo/pulls": func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("Content-Type", "application/json")
						prs := []*github.PullRequest{
							{
								Number:   github.Ptr(1),
								MergedAt: &github.Timestamp{Time: recentDate},
								User:     &github.User{Login: github.Ptr("elevated-maintainer")},
								MergedBy: &github.User{Login: github.Ptr("external-contributor")},
							},
						}
						mustEncode(w, prs)
					},
				})
			},
			checkFunc: func(h *maintainerHandler, owner, repo string) error {
				return h.checkMergedPRs(owner, repo)
			},
			elevatedUsers: []string{"elevated-maintainer"},
			activeUsers:   []string{"elevated-maintainer", "external-contributor"},
		},
		{
			name: "checkIssueComments ignores non-elevated commenters",
			setupHandler: func(t *testing.T) (*httptest.Server, *github.Client) {
				t.Helper()
				return setupMockGitHubServer(t, map[string]http.HandlerFunc{
					"/repos/test/repo/issues/comments": func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("Content-Type", "application/json")
						comments := []*github.IssueComment{
							{
								UpdatedAt: &github.Timestamp{Time: recentDate},
								User:      &github.User{Login: github.Ptr("maintainer")},
							},
							{
								UpdatedAt: &github.Timestamp{Time: recentDate},
								User:      &github.User{Login: github.Ptr("random-user")},
							},
						}
						mustEncode(w, comments)
					},
				})
			},
			checkFunc: func(h *maintainerHandler, owner, repo string) error {
				return h.checkIssueComments(owner, repo)
			},
			elevatedUsers: []string{"maintainer"},
			activeUsers:   []string{"maintainer", "random-user"},
		},
		{
			name: "checkReleases ignores non-elevated release authors",
			setupHandler: func(t *testing.T) (*httptest.Server, *github.Client) {
				t.Helper()
				return setupMockGitHubServer(t, map[string]http.HandlerFunc{
					"/repos/test/repo/releases": func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("Content-Type", "application/json")
						releases := []*github.RepositoryRelease{
							{
								CreatedAt: &github.Timestamp{Time: recentDate},
								Author:    &github.User{Login: github.Ptr("release-manager")},
							},
							{
								CreatedAt: &github.Timestamp{Time: recentDate},
								Author:    &github.User{Login: github.Ptr("bot-account")},
							},
						}
						mustEncode(w, releases)
					},
				})
			},
			checkFunc: func(h *maintainerHandler, owner, repo string) error {
				return h.checkReleases(owner, repo)
			},
			elevatedUsers: []string{"release-manager"},
			activeUsers:   []string{"release-manager", "bot-account"},
		},
		{
			name: "checkReviews ignores non-elevated reviewers",
			setupHandler: func(t *testing.T) (*httptest.Server, *github.Client) {
				t.Helper()
				return setupMockGitHubServer(t, map[string]http.HandlerFunc{
					"/repos/test/repo/pulls": func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("Content-Type", "application/json")
						prs := []*github.PullRequest{
							{Number: github.Ptr(1), UpdatedAt: &github.Timestamp{Time: recentDate}},
						}
						mustEncode(w, prs)
					},
					"/repos/test/repo/pulls/1/reviews": func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("Content-Type", "application/json")
						reviews := []*github.PullRequestReview{
							{
								SubmittedAt: &github.Timestamp{Time: recentDate},
								User:        &github.User{Login: github.Ptr("core-reviewer")},
							},
							{
								SubmittedAt: &github.Timestamp{Time: recentDate},
								User:        &github.User{Login: github.Ptr("external-reviewer")},
							},
						}
						mustEncode(w, reviews)
					},
				})
			},
			checkFunc: func(h *maintainerHandler, owner, repo string) error {
				return h.checkReviews(owner, repo)
			},
			elevatedUsers: []string{"core-reviewer"},
			activeUsers:   []string{"core-reviewer", "external-reviewer"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server, client := tt.setupHandler(t)
			defer server.Close()

			handler := &maintainerHandler{
				ctx:       context.Background(),
				ghClient:  client,
				cutoff:    cutoff,
				elevated:  make(map[string]struct{}),
				active:    make(map[string]bool),
				setupOnce: nil,
			}

			// Only add elevated users to the elevated set
			for _, user := range tt.elevatedUsers {
				handler.elevated[strings.ToLower(user)] = struct{}{}
			}

			err := tt.checkFunc(handler, "test", "repo")
			if err != nil {
				t.Fatalf("%s error = %v", tt.name, err)
			}

			// Verify only elevated users are marked active
			for _, user := range tt.elevatedUsers {
				if !handler.active[strings.ToLower(user)] {
					t.Errorf("Expected elevated user %s to be marked active", user)
				}
			}

			// Verify non-elevated users are NOT marked active
			for _, user := range tt.activeUsers {
				userLower := strings.ToLower(user)
				isElevated := false
				for _, elevated := range tt.elevatedUsers {
					if strings.ToLower(elevated) == userLower {
						isElevated = true
						break
					}
				}
				if !isElevated && handler.active[userLower] {
					t.Errorf("Security violation: non-elevated user %s was marked active", user)
				}
			}
		})
	}
}

// TestSecurityBoundary_EmptyLoginIgnored tests that empty login strings are safely ignored.
func TestSecurityBoundary_EmptyLoginIgnored(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cutoff := now.Add(-180 * 24 * time.Hour)
	recentDate := now.Add(-30 * 24 * time.Hour)

	server, client := setupMockGitHubServer(t, map[string]http.HandlerFunc{
		"/repos/test/repo/commits": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			commits := []*github.RepositoryCommit{
				{
					Author: &github.User{Login: github.Ptr("")}, // Empty login
					Commit: &github.Commit{
						Author: &github.CommitAuthor{Date: &github.Timestamp{Time: recentDate}},
					},
				},
				{
					Author: &github.User{Login: github.Ptr("valid-user")},
					Commit: &github.Commit{
						Author: &github.CommitAuthor{Date: &github.Timestamp{Time: recentDate}},
					},
				},
			}
			mustEncode(w, commits)
		},
	})
	defer server.Close()

	handler := &maintainerHandler{
		ctx:       context.Background(),
		ghClient:  client,
		cutoff:    cutoff,
		elevated:  map[string]struct{}{"valid-user": {}, "": {}}, // Empty string in elevated set
		active:    make(map[string]bool),
		setupOnce: nil,
	}

	err := handler.checkCommits("test", "repo")
	if err != nil {
		t.Fatalf("checkCommits() error = %v", err)
	}

	// Empty login should not be marked active (markActive returns false for empty strings)
	if handler.active[""] {
		t.Errorf("Empty login should not be marked active")
	}

	// Valid user should be marked active
	if !handler.active["valid-user"] {
		t.Errorf("Expected valid-user to be marked active")
	}
}

// TestSecurityBoundary_CaseSensitivity tests that username matching is case-insensitive.
func TestSecurityBoundary_CaseSensitivity(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cutoff := now.Add(-180 * 24 * time.Hour)
	recentDate := now.Add(-30 * 24 * time.Hour)

	server, client := setupMockGitHubServer(t, map[string]http.HandlerFunc{
		"/repos/test/repo/commits": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			commits := []*github.RepositoryCommit{
				{
					Author: &github.User{Login: github.Ptr("TestUser")}, // Mixed case
					Commit: &github.Commit{
						Author: &github.CommitAuthor{Date: &github.Timestamp{Time: recentDate}},
					},
				},
				{
					Author: &github.User{Login: github.Ptr("UPPERCASE")}, // All uppercase
					Commit: &github.Commit{
						Author: &github.CommitAuthor{Date: &github.Timestamp{Time: recentDate}},
					},
				},
			}
			mustEncode(w, commits)
		},
	})
	defer server.Close()

	handler := &maintainerHandler{
		ctx:      context.Background(),
		ghClient: client,
		cutoff:   cutoff,
		elevated: map[string]struct{}{
			"testuser":  {}, // lowercase version
			"uppercase": {}, // lowercase version
		},
		active:    make(map[string]bool),
		setupOnce: nil,
	}

	err := handler.checkCommits("test", "repo")
	if err != nil {
		t.Fatalf("checkCommits() error = %v", err)
	}

	// Both users should be marked active (case-insensitive matching)
	if !handler.active["testuser"] {
		t.Errorf("Expected testuser (lowercase) to be marked active for TestUser activity")
	}
	if !handler.active["uppercase"] {
		t.Errorf("Expected uppercase (lowercase) to be marked active for UPPERCASE activity")
	}
}

// TestSecurityBoundary_OnlyElevatedInActiveMap tests that the active map only contains elevated users.
func TestSecurityBoundary_OnlyElevatedInActiveMap(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cutoff := now.Add(-180 * 24 * time.Hour)
	recentDate := now.Add(-30 * 24 * time.Hour)

	server, client := setupMockGitHubServer(t, map[string]http.HandlerFunc{
		"/repos/test/repo/commits": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			commits := []*github.RepositoryCommit{
				{
					Author: &github.User{Login: github.Ptr("maintainer1")},
					Commit: &github.Commit{
						Author: &github.CommitAuthor{Date: &github.Timestamp{Time: recentDate}},
					},
				},
				{
					Author: &github.User{Login: github.Ptr("maintainer2")},
					Commit: &github.Commit{
						Author: &github.CommitAuthor{Date: &github.Timestamp{Time: recentDate}},
					},
				},
				{
					Author: &github.User{Login: github.Ptr("contributor1")},
					Commit: &github.Commit{
						Author: &github.CommitAuthor{Date: &github.Timestamp{Time: recentDate}},
					},
				},
				{
					Author: &github.User{Login: github.Ptr("contributor2")},
					Commit: &github.Commit{
						Author: &github.CommitAuthor{Date: &github.Timestamp{Time: recentDate}},
					},
				},
			}
			mustEncode(w, commits)
		},
	})
	defer server.Close()

	handler := &maintainerHandler{
		ctx:      context.Background(),
		ghClient: client,
		cutoff:   cutoff,
		elevated: map[string]struct{}{
			"maintainer1": {},
			"maintainer2": {},
		},
		active:    make(map[string]bool),
		setupOnce: nil,
	}

	err := handler.checkCommits("test", "repo")
	if err != nil {
		t.Fatalf("checkCommits() error = %v", err)
	}

	// Verify active map only contains elevated users
	for user := range handler.active {
		if _, isElevated := handler.elevated[user]; !isElevated {
			t.Errorf("Security violation: non-elevated user %s found in active map", user)
		}
	}

	// Verify maintainers are marked active
	if !handler.active["maintainer1"] {
		t.Errorf("Expected maintainer1 to be marked active")
	}
	if !handler.active["maintainer2"] {
		t.Errorf("Expected maintainer2 to be marked active")
	}

	// Verify contributors are NOT in active map
	if handler.active["contributor1"] {
		t.Errorf("Non-elevated contributor1 should not be in active map")
	}
	if handler.active["contributor2"] {
		t.Errorf("Non-elevated contributor2 should not be in active map")
	}
}
