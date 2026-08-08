// Copyright 2025 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gitfile

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/go-cmp/cmp"
)

func TestHandler(t *testing.T) {
	t.Parallel()

	var (
		want         = []string{"example.txt"}
		wantContents = []byte("hello world!")
	)

	dir := setupGitRepo(t)

	var h Handler
	h.Init(t.Context(), dir, "HEAD")

	files, err := h.ListFiles(allFiles)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if d := cmp.Diff(want, files); d != "" {
		t.Errorf("-got,+want: %s", d)
	}

	r, err := h.GetFile("example.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() { r.Close() })

	contents, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if d := cmp.Diff(wantContents, contents); d != "" {
		t.Errorf("-got,+want: %s", d)
	}

	err = h.Cleanup()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandlerPathTraversal(t *testing.T) {
	t.Parallel()
	dir := setupGitRepo(t)

	var h Handler
	h.Init(t.Context(), dir, "HEAD")

	if _, err := h.GetFile("../example.txt"); err == nil {
		t.Fatal("expected error reading a path outside the repo, got nil")
	}
}

func TestHandlerSymlinkEscape(t *testing.T) {
	t.Parallel()

	secret := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(secret, []byte("host secret"), 0o600); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dir := setupGitRepoWithSymlink(t, secret)

	var h Handler
	h.Init(t.Context(), dir, "HEAD")

	// The symlink is listed as a regular file, but reading it must not follow
	// the link to a location outside the checkout.
	if _, err := h.GetFile("innocent.txt"); err == nil {
		t.Fatal("expected error reading a symlink that escapes the repo, got nil")
	}
}

func setupGitRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	r, err := git.PlainInitWithOptions(dir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	w, err := r.Worktree()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	filename := filepath.Join(dir, "example.txt")

	if err = os.WriteFile(filename, []byte("hello world!"), 0o600); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err = w.Add("example.txt"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = w.Commit("commit message", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "John Doe",
			Email: "john@doe.org",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return dir
}

func setupGitRepoWithSymlink(t *testing.T, target string) string {
	t.Helper()

	dir := t.TempDir()
	r, err := git.PlainInitWithOptions(dir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	w, err := r.Worktree()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err = os.Symlink(target, filepath.Join(dir, "innocent.txt")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err = w.Add("innocent.txt"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = w.Commit("commit message", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "John Doe",
			Email: "john@doe.org",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return dir
}

func allFiles(path string) (bool, error) {
	return true, nil
}
