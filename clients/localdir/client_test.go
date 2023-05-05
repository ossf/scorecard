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
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/log"
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
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
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
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			// Test ListFiles API.
			for _, listfiletest := range testcase.listfileTests {
				files, e := listFiles(testcase.inputFolder)
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
				content, err := getFileContent(testcase.inputFolder, getcontenttest.filename)
				if getcontenttest.err != nil && !errors.Is(err, getcontenttest.err) {
					t.Errorf("test failed: expected - %v, got - %v", getcontenttest.err, err)
				}
				if getcontenttest.err == nil && !cmp.Equal(getcontenttest.output, content) {
					t.Errorf("test failed: expected - %s, got - %s", string(getcontenttest.output), string(content))
				}
			}
		})
	}
}
