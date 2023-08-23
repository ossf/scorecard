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

package git

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	gitV5 "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v4/clients"
)

func createTestRepo(t *testing.T) (path string) {
	t.Helper()
	dir, err := os.MkdirTemp("", "testrepo")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir) // nolint:errcheck
	})
	r, err := gitV5.PlainInit(dir, false)
	if err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	w, err := r.Worktree()
	if err != nil {
		t.Fatalf("Failed to get worktree: %v", err)
	}

	// Create a new file
	filePath := filepath.Join(dir, "file")
	err = os.WriteFile(filePath, []byte("Hello, World!"), 0o644) //nolint:gosec
	if err != nil {
		t.Fatalf("Failed to write a file: %v", err)
	}

	// Add it to the staging area
	_, err = w.Add("file")
	if err != nil {
		t.Fatalf("Failed to add a file to staging area: %v", err)
	}

	// Commit
	_, err = w.Commit("Initial commit", &gitV5.CommitOptions{
		Author: &object.Signature{
			Name:  "Test Author",
			Email: "author@example.com",
			When:  time.Now(),
		},
	})

	if err != nil {
		t.Fatalf("Failed to make initial commit: %v", err)
	}

	return dir
}

//nolint:paralleltest
func TestInitRepo(t *testing.T) {
	tests := []struct { //nolint:govet
		name        string
		uri         string
		commitSHA   string
		commitDepth int
		expectedErr string
	}{
		{
			name:        "Success",
			uri:         "file://%s",
			commitSHA:   "HEAD",
			commitDepth: 1,
		},
		{
			name:        "InvalidUri",
			uri:         ":",
			commitSHA:   "",
			commitDepth: 1,
			expectedErr: "repository does not exist",
		},
		{
			name:        "NegativeCommitDepth",
			uri:         "file://%s",
			commitSHA:   "HEAD",
			commitDepth: -1,
		},
	}

	repoPath := createTestRepo(t)

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			uri := fmt.Sprintf(test.uri, repoPath)

			client := &Client{}
			err := client.InitRepo(uri, test.commitSHA, test.commitDepth)
			if (test.expectedErr != "") != (err != nil) {
				t.Errorf("Unexpected error during InitRepo: %v", err)
			}
		})
	}
}

func TestListCommits(t *testing.T) {
	repoPath := createTestRepo(t)

	client := &Client{}
	commitDepth := 1
	expectedLen := 1
	commitSHA := "HEAD"
	uri := fmt.Sprintf("file://%s", repoPath)
	if err := client.InitRepo(uri, commitSHA, commitDepth); err != nil {
		t.Fatalf("InitRepo(%s) failed: %v", uri, err)
	}

	// Act
	commits, err := client.ListCommits()
	if err != nil {
		t.Fatalf("ListCommits() failed: %v", err)
	}

	// Assert
	if len(commits) != expectedLen {
		t.Errorf("ListCommits() returned %d commits, want %d", len(commits), expectedLen)
	}
}

//nolint:paralleltest
func TestSearch(t *testing.T) {
	testCases := []struct {
		name     string
		request  clients.SearchRequest
		expected clients.SearchResponse
	}{
		{
			name: "Search with valid query",
			request: clients.SearchRequest{
				Query: "Hello",
			},
			expected: clients.SearchResponse{
				Results: []clients.SearchResult{
					{
						Path: "file",
					},
					{
						Path: "test.txt",
					},
				},
				Hits: 2,
			},
		},
		{
			name: "Search with zero results",
			request: clients.SearchRequest{
				Query: "Invalid",
			},
			expected: clients.SearchResponse{
				Hits: 0,
			},
		},
	}

	// Use the same test repo for all test cases.
	repoPath := createTestRepo(t)
	filePath := filepath.Join(repoPath, "test.txt")
	err := os.WriteFile(filePath, []byte("Hello, World!"), 0o644) //nolint:gosec
	if err != nil {
		t.Fatalf("WriteFile() failed: %v", err)
	}

	// Make a commit that adds the file.
	r, err := gitV5.PlainOpen(repoPath)
	if err != nil {
		t.Fatalf("PlainOpen() failed: %v", err)
	}
	w, err := r.Worktree()
	if err != nil {
		t.Fatalf("Worktree() failed: %v", err)
	}
	_, err = w.Add("test.txt")
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}
	_, err = w.Commit("Add test.txt", &gitV5.CommitOptions{
		Author: &object.Signature{
			Name:  "Test Author",
			Email: "author@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		t.Fatalf("Commit() failed: %v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &Client{}
			uri := fmt.Sprintf("file://%s", repoPath)
			if err := client.InitRepo(uri, "HEAD", 1); err != nil {
				t.Fatalf("InitRepo(%s) failed: %v", uri, err)
			}

			response, err := client.Search(tc.request)
			if err != nil {
				t.Fatalf("Search() failed: %v", err)
			}

			if diff := cmp.Diff(tc.expected, response, cmpopts.IgnoreUnexported(clients.SearchResult{})); diff != "" {
				t.Errorf("Search() returned diff (-want +got):\n%s", diff)
			}
		})
	}
}
