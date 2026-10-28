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
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	gitV5 "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/clients/localdir"
)

func createTestRepo(t *testing.T) (path string) {
	t.Helper()
	dir := t.TempDir()

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
	err = os.WriteFile(filePath, []byte("Hello, World!"), 0o600)
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

func TestInitRepo(t *testing.T) {
	t.Parallel()
	repoPath := createTestRepo(t)
	r, err := gitV5.PlainOpen(repoPath)
	if err != nil {
		t.Fatal(err)
	}
	firstCommit, err := r.Head()
	if err != nil {
		t.Fatal(err)
	}
	w, err := r.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	latestCommit, err := w.Commit("Second commit", &gitV5.CommitOptions{
		AllowEmptyCommits: true,
		Author: &object.Signature{
			Name: "Test Author", Email: "author@example.com", When: time.Now(),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	validSHA := firstCommit.Hash().String()

	tests := []struct {
		name        string
		commitSHA   string
		expectedErr error
		expectedSHA string
		commitDepth int
	}{
		{
			name:        "Success",
			commitSHA:   "HEAD",
			expectedSHA: latestCommit.String(),
			commitDepth: 1,
		},
		{
			name:        "NegativeCommitDepth",
			commitSHA:   "HEAD",
			commitDepth: -1,
		},
		{name: "ExplicitCommit", commitSHA: validSHA, expectedSHA: validSHA, commitDepth: 1},
		{name: "UppercaseCommit", commitSHA: strings.ToUpper(validSHA), expectedSHA: validSHA, commitDepth: 1},
		{name: "EmptyCommit", commitSHA: "", expectedErr: clients.ErrInvalidCommitSHA},
		{name: "NonHexCommit", commitSHA: strings.Repeat("z", 40), expectedErr: clients.ErrInvalidCommitSHA},
		{name: "ShortCommit", commitSHA: validSHA[:8], expectedErr: clients.ErrInvalidCommitSHA},
		{name: "OddLengthCommit", commitSHA: validSHA[:39], expectedErr: clients.ErrInvalidCommitSHA},
		{name: "LongCommit", commitSHA: validSHA + "00", expectedErr: clients.ErrInvalidCommitSHA},
		{name: "InvalidSuffix", commitSHA: validSHA + "zz", expectedErr: clients.ErrInvalidCommitSHA},
		{name: "ZeroCommit", commitSHA: strings.Repeat("0", 40), expectedErr: clients.ErrInvalidCommitSHA},
		{name: "UnknownCommit", commitSHA: strings.Repeat("1", 40), expectedErr: plumbing.ErrObjectNotFound},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			uri := repoPath

			client := &Client{}
			t.Cleanup(func() {
				if err := client.Close(); err != nil {
					t.Errorf("Close() failed: %v", err)
				}
			})
			repo, err := localdir.MakeLocalDirRepo(uri)
			if err != nil {
				t.Fatalf("MakeLocalDirRepo(%s) failed: %v", uri, err)
			}
			err = client.InitRepo(repo, test.commitSHA, test.commitDepth)
			if test.expectedErr != nil {
				if !errors.Is(err, test.expectedErr) {
					t.Fatalf("InitRepo() error = %v, want %v", err, test.expectedErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("InitRepo() unexpected error: %v", err)
			}
			if test.expectedSHA != "" {
				assertHead(t, client, test.expectedSHA)
			}
		})
	}
}

func assertHead(t *testing.T, client *Client, want string) {
	t.Helper()
	head, err := client.gitRepo.Head()
	if err != nil {
		t.Fatal(err)
	}
	if got := head.Hash().String(); got != want {
		t.Errorf("checked out commit = %q, want %q", got, want)
	}
}

func TestInitRepoInvalidSHAKeepsState(t *testing.T) {
	t.Parallel()
	repo, err := localdir.MakeLocalDirRepo(createTestRepo(t))
	if err != nil {
		t.Fatal(err)
	}
	client := &Client{}
	t.Cleanup(func() { client.Close() })
	if err := client.InitRepo(repo, clients.HeadSHA, 1); err != nil {
		t.Fatalf("InitRepo(HEAD) unexpected error: %v", err)
	}
	if err := client.InitRepo(repo, "bad", 1); !errors.Is(err, clients.ErrInvalidCommitSHA) {
		t.Fatalf("InitRepo(bad) error = %v, want %v", err, clients.ErrInvalidCommitSHA)
	}
	// The rejected call must not tear down the existing clone.
	if _, err := client.ListFiles(func(string) (bool, error) { return true, nil }); err != nil {
		t.Errorf("ListFiles() after rejected InitRepo: %v", err)
	}
}

func TestListCommits(t *testing.T) {
	t.Parallel()
	repoPath := createTestRepo(t)

	client := &Client{}
	commitDepth := 1
	expectedLen := 1
	commitSHA := "HEAD"
	uri := repoPath
	repo, err := localdir.MakeLocalDirRepo(uri)
	if err != nil {
		t.Fatalf("MakeLocalDirRepo(%s) failed: %v", uri, err)
	}
	if err := client.InitRepo(repo, commitSHA, commitDepth); err != nil {
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

func TestSearch(t *testing.T) {
	t.Parallel()
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
	err := os.WriteFile(filePath, []byte("Hello, World!"), 0o600)
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
			t.Parallel()
			client := &Client{}
			uri := repoPath
			repo, err := localdir.MakeLocalDirRepo(uri)
			if err != nil {
				t.Fatalf("MakeLocalDirRepo(%s) failed: %v", uri, err)
			}
			if err := client.InitRepo(repo, "HEAD", 1); err != nil {
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

func TestFileAccessSymlinks(t *testing.T) {
	t.Parallel()
	const errEscapes = "path escapes from parent"

	secret := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(secret, []byte("host secret"), 0o600); err != nil {
		t.Fatalf("Failed to write a file: %v", err)
	}
	// A run of "../" longer than any temp path reaches the filesystem root.
	relSecret := strings.Repeat("../", 64) + strings.TrimPrefix(filepath.ToSlash(secret), "/")

	dir := createTestRepo(t)
	r, err := gitV5.PlainOpen(dir)
	if err != nil {
		t.Fatalf("Failed to open git repo: %v", err)
	}
	w, err := r.Worktree()
	if err != nil {
		t.Fatalf("Failed to get worktree: %v", err)
	}
	links := map[string]string{
		"in-repo":         "file",
		"escape-absolute": secret,
		"escape-relative": relSecret,
		"dangling":        "missing",
	}
	for name, target := range links {
		if err := os.Symlink(target, filepath.Join(dir, name)); err != nil {
			t.Fatalf("Failed to create symlink: %v", err)
		}
		if _, err := w.Add(name); err != nil {
			t.Fatalf("Failed to add a file to staging area: %v", err)
		}
	}
	_, err = w.Commit("Add symlinks", &gitV5.CommitOptions{
		Author: &object.Signature{
			Name:  "Test Author",
			Email: "author@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	repo, err := localdir.MakeLocalDirRepo(dir)
	if err != nil {
		t.Fatalf("MakeLocalDirRepo(%s) failed: %v", dir, err)
	}
	client := &Client{}
	if err := client.InitRepo(repo, clients.HeadSHA, 0); err != nil {
		t.Fatalf("InitRepo: %v", err)
	}
	t.Cleanup(func() { client.Close() })

	files, err := client.ListFiles(func(string) (bool, error) { return true, nil })
	if err != nil {
		t.Fatalf("ListFiles: %v", err)
	}
	if d := cmp.Diff([]string{"file", "in-repo"}, files, cmpopts.SortSlices(func(a, b string) bool { return a < b })); d != "" {
		t.Errorf("ListFiles: -want,+got: %s", d)
	}

	rc, err := client.GetFileReader("in-repo")
	if err != nil {
		t.Fatalf("GetFileReader(in-repo): %v", err)
	}
	t.Cleanup(func() { rc.Close() })
	contents, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if got := string(contents); got != "Hello, World!" {
		t.Errorf("GetFileReader(in-repo) = %q, want %q", got, "Hello, World!")
	}

	// Escaping names aren't listed, and reading one is still refused.
	for _, name := range []string{"escape-absolute", "escape-relative", "../file"} {
		_, err := client.GetFileReader(name)
		if err == nil || !strings.Contains(err.Error(), errEscapes) {
			t.Errorf("GetFileReader(%s): got %v, want error containing %q", name, err, errEscapes)
		}
	}
}
