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

package githubrepo

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type listfileTest struct {
	predicate func(string) bool
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

// nolint: gocognit
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
					predicate: func(string) bool { return true },
					outcome:   []string{"file0", "dir1/file1", "dir1/dir2/file2"},
				},
				{
					// Skips all files inside `dir1/dir2` directory.
					predicate: func(fn string) bool { return !strings.HasPrefix(fn, "dir1/dir2") },
					outcome:   []string{"file0", "dir1/file1"},
				},
				{
					// Skips all files.
					predicate: func(fn string) bool { return false },
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
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			tempDir, err := ioutil.TempDir("", repoDir)
			if err != nil {
				t.Fatalf("test failed to create TempDir: %v", err)
			}
			tempFile, err := ioutil.TempFile(tempDir, repoFilename)
			if err != nil {
				t.Fatalf("test failed to create TempFile: %v", err)
			}
			testFile, err := os.OpenFile(testcase.inputFile, os.O_RDONLY, 0o644)
			if err != nil {
				t.Fatalf("unable to open testfile: %v", err)
			}
			if _, err := io.Copy(tempFile, testFile); err != nil {
				t.Fatalf("unable to do io.Copy: %v", err)
			}

			// Extract tarball.
			client := Client{
				tempDir:     tempDir,
				tempTarFile: tempFile.Name(),
			}
			if err := client.extractTarball(); err != nil {
				t.Fatalf("test failed: %v", err)
			}

			// Test ListFiles API.
			for _, listfiletest := range testcase.listfileTests {
				if !cmp.Equal(listfiletest.outcome,
					client.ListFiles(listfiletest.predicate),
					cmpopts.SortSlices(isSortedString)) {
					t.Errorf("test failed: expected - %q, got - %q", listfiletest.outcome, client.ListFiles(listfiletest.predicate))
				}
			}

			// Test GetFileContent API.
			for _, getcontenttest := range testcase.getcontentTests {
				content, err := client.GetFileContent(getcontenttest.filename)
				if getcontenttest.err != nil && !errors.As(err, &getcontenttest.err) {
					t.Errorf("test failed: expected - %v, got - %v", getcontenttest.err, err)
				}
				if getcontenttest.err == nil && !cmp.Equal(getcontenttest.output, content) {
					t.Errorf("test failed: expected - %s, got - %s", string(getcontenttest.output), string(content))
				}
			}

			// Test that files get deleted.
			if err := client.cleanup(); err != nil {
				t.Errorf("test failed: %v", err)
			}
			if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
				t.Errorf("%v", err)
			}
		})
	}
}
