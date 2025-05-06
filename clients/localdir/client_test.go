// Copyright 2020 OpenSSF Scorecard Authors
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

package localdir

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/log"
)

func TestClient_CreationAndCaching(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		inputFolder string
		err         error
		outputFiles []string
	}{
		{
			name:        "invalid fullpath",
			outputFiles: []string{},
			inputFolder: "invalid/fullpath",
			err:         os.ErrNotExist,
		},
		{
			name:        "invalid relative path",
			outputFiles: []string{},
			inputFolder: "invalid/relative/path",
			err:         os.ErrNotExist,
		},
		{
			name: "repo 0",
			outputFiles: []string{
				"file0", "dir1/file1", "dir1/dir2/file2",
			},
			inputFolder: "testdata/repo0",
			err:         nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			logger := log.NewLogger(log.DebugLevel)

			// Create repo.
			repo, err := MakeLocalDirRepo(tt.inputFolder)
			if !errors.Is(err, tt.err) {
				t.Errorf("MakeLocalDirRepo: %v, expected %v", err, tt.err)
			}

			if err != nil {
				return
			}

			client := CreateLocalDirClient(ctx, logger)
			if err := client.InitRepo(repo, clients.HeadSHA, 30); err != nil {
				t.Errorf("InitRepo: %v", err)
			}

			// List files.
			files, err := client.ListFiles(func(string) (bool, error) { return true, nil })
			if !errors.Is(err, tt.err) {
				t.Errorf("CreateLocalDirClient: %v, expected %v", err, tt.err)
			}

			if !cmp.Equal(tt.outputFiles, files, cmpopts.SortSlices(func(x, y string) bool { return x < y })) {
				t.Errorf("Got diff: %s", cmp.Diff(tt.outputFiles, files))
			}

			// List files a second time to test the caching.
			files2, err := client.ListFiles(func(string) (bool, error) { return true, nil })
			if !errors.Is(err, tt.err) {
				t.Errorf("CreateLocalDirClient: %v, expected %v", err, tt.err)
			}

			if !cmp.Equal(tt.outputFiles, files2, cmpopts.SortSlices(func(x, y string) bool { return x < y })) {
				t.Errorf("Got diff: %s", cmp.Diff(tt.outputFiles, files2))
			}
		})
	}
}

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

//nolint:gocognit
func TestClient_GetFileListAndContent(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name            string
		inputFolder     string
		listfileTests   []listfileTest
		getcontentTests []getcontentTest
	}{
		{
			name:        "Basic",
			inputFolder: "testdata/repo0",
			listfileTests: []listfileTest{
				{
					// Returns all files .
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
				},
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			// Test ListFiles API.
			for _, listfiletest := range testcase.listfileTests {
				files, e := listFiles(testcase.inputFolder, log.NewLogger(log.DebugLevel))
				matchedFiles, err := applyPredicate(files, e, listfiletest.predicate)
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
				f, err := getFile(testcase.inputFolder, getcontenttest.filename)
				if getcontenttest.err != nil && !errors.Is(err, getcontenttest.err) {
					t.Errorf("test failed: expected - %v, got - %v", getcontenttest.err, err)
				}
				if err == nil {
					content, err := io.ReadAll(f)
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
					if !cmp.Equal(getcontenttest.output, content) {
						t.Errorf("test failed: expected - %s, got - %s", string(getcontenttest.output), string(content))
					}
				}
			}
		})
	}
}

func TestListFile_symlink(t *testing.T) {
	t.Parallel()
	const (
		realFilename = "foo.txt"
		nonExistent  = "bar.txt"
		symlinkName  = "symlink.txt"
	)
	testcases := []struct {
		name      string
		symTarget string
		want      []string
	}{
		{
			name:      "dangling symlink ignored",
			symTarget: nonExistent,
			want:      []string{realFilename},
		},
		{
			name:      "symlink with real target included",
			symTarget: realFilename,
			want:      []string{realFilename, symlinkName},
		},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			realFileFullPath := filepath.Join(dir, realFilename)
			f, err := os.Create(realFileFullPath)
			if err != nil {
				t.Fatal("creating real file", err)
			}
			f.Close()
			target := filepath.Join(dir, tt.symTarget)
			linkname := filepath.Join(dir, symlinkName)
			if err = os.Symlink(target, linkname); err != nil {
				t.Fatal("making symlink", err)
			}
			files, err := listFiles(dir, log.NewLogger(log.DebugLevel))
			if err != nil {
				t.Fatal("expected listFiles to ignore the symlink error")
			}
			if !cmp.Equal(files, tt.want, cmpopts.SortSlices(isSortedString)) {
				t.Errorf("test failed: expected - %q, got - %q", tt.want, files)
			}
		})
	}
}

func TestGetFileReader_symlink(t *testing.T) {
	t.Parallel()
	const (
		realFilename = "foo.txt"
		linkEscapes  = "../bar.txt"
		symlinkName  = "symlink.txt"
	)
	testcases := []struct {
		name      string
		symTarget string
		wantErr   bool
	}{
		{
			name:      "symlink with real target",
			symTarget: realFilename,
			wantErr:   false,
		},
		{
			name:      "symlink which escapes should fail",
			symTarget: linkEscapes,
			wantErr:   true,
		},
	}
	contents := []byte("hello!")
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			realFileFullPath := filepath.Join(dir, realFilename)
			f, err := os.Create(realFileFullPath)
			if err != nil {
				t.Fatal("creating real file", err)
			}
			t.Log("created file", realFileFullPath)
			if _, err = f.Write(contents); err != nil {
				t.Fatal("writing file contents", err)
			}
			f.Close()
			target := tt.symTarget // must be a local symlink, do not join with temp dir
			linkname := filepath.Join(dir, symlinkName)
			if err = os.Symlink(target, linkname); err != nil {
				t.Fatal("making symlink", err)
			}
			t.Logf("created symlink %q -> %q", linkname, target)
			f, err = getFile(dir, symlinkName) // getFile handles path joining with temp dir
			if (err != nil) != tt.wantErr {
				t.Fatalf("wanted err? (%t) but err was %v", tt.wantErr, err)
			}
			defer f.Close()
			if tt.wantErr {
				return // dont compare contents if the read should have error'd
			}
			got, err := io.ReadAll(f)
			if err != nil {
				t.Fatal("reading symlink contents", err)
			}
			if !cmp.Equal(got, contents) {
				t.Errorf("expected file content: %q, got: %q", contents, got)
			}
		})
	}
}
