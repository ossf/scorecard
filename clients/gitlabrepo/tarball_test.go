// Copyright 2021 Security Scorecard Authors
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
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type listfileTest struct {
	predicate func(string) (bool, error)
	err       error
	outcome   []string
}

type getcontentTest struct {
	err      error
	filename string
	output   []byte
}

func isSortedString(x, y string) bool {
	return x < y
}

func setup(inputFile string) (tarballHandler, error) {
	tempDir, err := os.MkdirTemp("", repoDir)
	if err != nil {
		return tarballHandler{}, fmt.Errorf("test failed to create TempDir: %w", err)
	}
	tempFile, err := os.CreateTemp(tempDir, repoFilename)
	if err != nil {
		return tarballHandler{}, fmt.Errorf("test failed to create TempFile: %w", err)
	}
	testFile, err := os.OpenFile(inputFile, os.O_RDONLY, 0o644)
	if err != nil {
		return tarballHandler{}, fmt.Errorf("unable to open testfile: %w", err)
	}
	if _, err := io.Copy(tempFile, testFile); err != nil {
		return tarballHandler{}, fmt.Errorf("unable to do io.Copy: %w", err)
	}
	tarballHandler := tarballHandler{
		tempDir:     tempDir,
		tempTarFile: tempFile.Name(),
		once:        new(sync.Once),
	}
	tarballHandler.once.Do(func() {
		// We don't want to run the code in tarballHandler.setup(), so if we execute tarballHandler.once.Do() right
		// here, it won't get executed later when setup() is called.
	})
	return tarballHandler, nil
}

//nolint:gocognit
func TestExtractTarball(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name            string
		inputFile       string
		listfileTests   []listfileTest
		getcontentTests []getcontentTest
	}{
		{
			name:      "Basic",
			inputFile: "testdata/basic.tar.gz",
			listfileTests: []listfileTest{
				{
					// Returns all files in the tarball.
					predicate: func(string) (bool, error) { return true, nil },
					outcome:   []string{"file0", "dir1/file1", "dir1/dir2/file2"},
				},
				{
					// Skips all files inside `dir1/dir2` directory.
					predicate: func(fn string) (bool, error) { return !strings.HasPrefix(fn, "dir1/dir2"), nil },
					outcome:   []string{"file0", "dir1/file1"},
				},
				{
					// Skips all files.
					predicate: func(fn string) (bool, error) { return false, nil },
					outcome:   []string{},
				},
			},
			getcontentTests: []getcontentTest{
				{
					filename: "file0",
					output:   []byte("content0\n"),
				},
				{
					filename: "dir1/file1",
					output:   []byte("content1\n"),
				},
				{
					filename: "dir1/dir2/file2",
					output:   []byte("content2\n"),
				},
				{
					filename: "does/not/exist",
					err:      os.ErrNotExist,
				},
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			handler, err := setup(testcase.inputFile)
			if err != nil {
				t.Fatalf("test setup failed: %v", err)
			}

			// Extract tarball.
			if err := handler.extractTarball(); err != nil {
				t.Fatalf("test failed: %v", err)
			}

			// Test ListFiles API.
			for _, listfiletest := range testcase.listfileTests {
				matchedFiles, err := handler.listFiles(listfiletest.predicate)
				if !errors.Is(err, listfiletest.err) {
					t.Errorf("test failed: expected - %v, got - %v", listfiletest.err, err)
					continue
				}
				if !cmp.Equal(listfiletest.outcome,
					matchedFiles,
					cmpopts.SortSlices(isSortedString)) {
					t.Errorf("test failed: expected - %q, got - %q", listfiletest.outcome, matchedFiles)
				}
			}

			// Test GetFileContent API.
			for _, getcontenttest := range testcase.getcontentTests {
				f, err := handler.getFile(getcontenttest.filename)
				if getcontenttest.err != nil && !errors.Is(err, getcontenttest.err) {
					t.Errorf("test failed: expected - %v, got - %v", getcontenttest.err, err)
				}
				if getcontenttest.err == nil {
					content, err := io.ReadAll(f)
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
					if !cmp.Equal(getcontenttest.output, content) {
						t.Errorf("test failed: expected - %s, got - %s", string(getcontenttest.output), string(content))
					}
				}
			}

			// Test that files get deleted.
			if err := handler.cleanup(); err != nil {
				t.Errorf("test failed: %v", err)
			}
			if _, err := os.Stat(handler.tempDir); !os.IsNotExist(err) {
				t.Errorf("%v", err)
			}
			if len(handler.files) != 0 {
				t.Error("client.files not cleaned up!")
			}
		})
	}
}

// writeTarball writes a gzipped tarball holding one regular file to dir and
// returns its path.
func writeTarball(t *testing.T, dir, name, content string) string {
	t.Helper()
	f, err := os.CreateTemp(dir, repoFilename)
	if err != nil {
		t.Fatalf("creating tarball: %v", err)
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	hdr := &tar.Header{Typeflag: tar.TypeReg, Name: name, Mode: 0o644, Size: int64(len(content))}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("writing tar header: %v", err)
	}
	if _, err := tw.Write([]byte(content)); err != nil {
		t.Fatalf("writing tar entry: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("closing tar writer: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("closing gzip writer: %v", err)
	}
	return f.Name()
}

func TestExtractTarball_outsideTempDir(t *testing.T) {
	t.Parallel()
	tests := []struct {
		wantErr error
		name    string
		entry   string
		symlink bool
	}{
		{
			name:    "entry with parent directory",
			entry:   "repo/../outside/file",
			wantErr: errZipSlip,
		},
		{
			name:    "entry under a symlink to outside",
			entry:   "repo/link/file",
			symlink: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parent := t.TempDir()
			tempDir := filepath.Join(parent, "temp")
			outside := filepath.Join(parent, "outside")
			for _, dir := range []string{tempDir, outside} {
				if err := os.Mkdir(dir, 0o755); err != nil {
					t.Fatalf("creating dir: %v", err)
				}
			}
			if tt.symlink {
				if err := os.Symlink(filepath.Join("..", "outside"), filepath.Join(tempDir, "link")); err != nil {
					t.Skipf("creating symlink: %v", err)
				}
			}
			handler := tarballHandler{
				tempDir:     tempDir,
				tempTarFile: writeTarball(t, parent, tt.entry, "content"),
			}

			err := handler.extractTarball()
			if err == nil {
				t.Error("extractTarball() succeeded, want an error")
			} else if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("extractTarball() error = %v, want %v", err, tt.wantErr)
			}
			if _, err := os.Stat(filepath.Join(outside, "file")); !errors.Is(err, os.ErrNotExist) {
				t.Errorf("file written outside the temp dir: %v", err)
			}
		})
	}
}

func TestGetFile_outsideTempDir(t *testing.T) {
	t.Parallel()
	parent := t.TempDir()
	tempDir := filepath.Join(parent, "temp")
	if err := os.Mkdir(tempDir, 0o755); err != nil {
		t.Fatalf("creating dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(parent, "outside"), []byte("outside"), 0o600); err != nil {
		t.Fatalf("writing file: %v", err)
	}
	names := []string{"../outside"}
	if err := os.Symlink(filepath.Join("..", "outside"), filepath.Join(tempDir, "link")); err != nil {
		t.Logf("skipping symlink case: %v", err)
	} else {
		names = append(names, "link")
	}
	handler := tarballHandler{tempDir: tempDir, once: new(sync.Once)}
	handler.once.Do(func() {}) // skip setup()

	for _, name := range names {
		if f, err := handler.getFile(name); err == nil {
			f.Close()
			t.Errorf("getFile(%q) opened a file outside the temp dir", name)
		}
	}

	// With no tarball downloaded there is no temp dir, and names must not be
	// opened relative to the working directory, which holds tarball.go.
	handler.tempDir = ""
	if f, err := handler.getFile("tarball.go"); err == nil {
		f.Close()
		t.Error(`getFile("tarball.go") opened a file without a temp dir`)
	}
}
