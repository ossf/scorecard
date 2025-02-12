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
	"context"
	"errors"
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
	h.Init(context.Background(), dir, "HEAD")

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
	h.Init(context.Background(), dir, "HEAD")

	_, err := h.GetFile("../example.txt")
	if !errors.Is(err, errPathTraversal) {
		t.Fatalf("expected path traversal error: got %v", err)
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

func allFiles(path string) (bool, error) {
	return true, nil
}
