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
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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

// errEscapes is the text os.Root uses when a name resolves outside the root.
const errEscapes = "path escapes from parent"

func TestHandlerPathTraversal(t *testing.T) {
	t.Parallel()
	dir := setupGitRepo(t)

	var h Handler
	h.Init(t.Context(), dir, "HEAD")
	t.Cleanup(func() {
		if err := h.Cleanup(); err != nil {
			t.Errorf("Cleanup: %v", err)
		}
	})

	if _, err := h.ListFiles(allFiles); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err := h.GetFile("../example.txt")
	if err == nil || !strings.Contains(err.Error(), errEscapes) {
		t.Fatalf("GetFile(../example.txt): got %v, want error containing %q", err, errEscapes)
	}
}

func TestHandlerSymlinks(t *testing.T) {
	t.Parallel()

	secret := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(secret, []byte("host secret"), 0o600); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// A run of "../" longer than any temp path reaches the filesystem root.
	relSecret := strings.Repeat("../", 64) + strings.TrimPrefix(filepath.ToSlash(secret), "/")

	dir := setupGitRepoWithSymlinks(t, map[string]string{
		"in-repo.txt":         "example.txt",
		"escape-absolute.txt": secret,
		"escape-relative.txt": relSecret,
		"dangling.txt":        "missing.txt",
		"dir-link":            "sub",
	})

	var h Handler
	h.Init(t.Context(), dir, "HEAD")
	t.Cleanup(func() {
		if err := h.Cleanup(); err != nil {
			t.Errorf("Cleanup: %v", err)
		}
	})

	files, err := h.ListFiles(allFiles)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"example.txt", "in-repo.txt", "sub/nested.txt"}
	if d := cmp.Diff(want, files, cmpopts.SortSlices(func(a, b string) bool { return a < b })); d != "" {
		t.Errorf("ListFiles: -want,+got: %s", d)
	}

	r, err := h.GetFile("in-repo.txt")
	if err != nil {
		t.Fatalf("GetFile(in-repo.txt): unexpected error: %v", err)
	}
	t.Cleanup(func() { r.Close() })
	contents, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := string(contents); got != "hello world!" {
		t.Errorf("GetFile(in-repo.txt) = %q, want %q", got, "hello world!")
	}

	// Escaping links aren't listed, and reading one by name is still refused.
	for _, name := range []string{"escape-absolute.txt", "escape-relative.txt"} {
		_, err := h.GetFile(name)
		if err == nil || !strings.Contains(err.Error(), errEscapes) {
			t.Errorf("GetFile(%s): got %v, want error containing %q", name, err, errEscapes)
		}
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

// setupGitRepoWithSymlinks commits example.txt, sub/nested.txt, and the
// given symlinks (name to target).
func setupGitRepoWithSymlinks(t *testing.T, links map[string]string) string {
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

	if err = os.Mkdir(filepath.Join(dir, "sub"), 0o700); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	regular := map[string]string{
		"example.txt":    "hello world!",
		"sub/nested.txt": "nested",
	}
	for name, contents := range regular {
		if err = os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o600); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, err = w.Add(name); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	for name, target := range links {
		if err = os.Symlink(target, filepath.Join(dir, name)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, err = w.Add(name); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
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
