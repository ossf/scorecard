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

package data

import (
	"errors"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	sce "github.com/ossf/scorecard/v4/errors"
)

type outcome struct {
	expectedErr error
	repo        RepoFormat
	hasError    bool
}

//nolint:gocognit
func TestCsvIterator(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name     string
		filename string
		outcomes []outcome
	}{
		{
			name:     "Basic",
			filename: "testdata/basic.csv",
			outcomes: []outcome{
				{
					hasError: false,
					repo: RepoFormat{
						Repo: "github.com/owner1/repo1",
					},
				},
				{
					hasError: false,
					repo: RepoFormat{
						Repo: "github.com/owner2/repo2",
					},
				},
				{
					hasError: false,
					repo: RepoFormat{
						Repo:     "github.com/owner3/repo3",
						Metadata: []string{"meta"},
					},
				},
			},
		},
		{
			name:     "Comment",
			filename: "testdata/comment.csv",
			outcomes: []outcome{
				{
					hasError: false,
					repo: RepoFormat{
						Repo: "github.com/owner1/repo1",
					},
				},
				{
					hasError: false,
					repo: RepoFormat{
						Repo: "github.com/owner2/repo2",
					},
				},
				{
					hasError: false,
					repo: RepoFormat{
						Repo:     "github.com/owner3/repo3",
						Metadata: []string{"meta"},
					},
				},
			},
		},
		{
			name:     "FailingURLs",
			filename: "testdata/failing_urls.csv",
			outcomes: []outcome{
				{
					hasError:    true,
					expectedErr: sce.ErrorUnsupportedHost,
				},
				{
					hasError:    true,
					expectedErr: sce.ErrorInvalidURL,
				},
				{
					hasError:    true,
					expectedErr: sce.ErrorInvalidURL,
				},
			},
		},
		{
			name:     "EmptyRows",
			filename: "testdata/empty_row.csv",
			outcomes: []outcome{
				{
					hasError: false,
					repo: RepoFormat{
						Repo: "github.com/owner1/repo1",
					},
				},
				{
					hasError: false,
					repo: RepoFormat{
						Repo: "github.com/owner2/repo2",
					},
				},
				{
					hasError: false,
					repo: RepoFormat{
						Repo:     "github.com/owner3/repo3",
						Metadata: []string{"meta"},
					},
				},
			},
		},
		{
			name:     "ExtraColumns",
			filename: "testdata/extra_column.csv",
			outcomes: []outcome{
				{
					hasError: false,
					repo: RepoFormat{
						Repo: "github.com/owner1/repo1",
					},
				},
				{
					hasError: false,
					repo: RepoFormat{
						Repo: "github.com/owner2/repo2",
					},
				},
				{
					hasError: true,
				},
			},
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()
			testFile, err := os.OpenFile(testcase.filename, os.O_RDONLY, 0o644)
			if err != nil {
				t.Errorf("failed to open %s: %v", testcase.filename, err)
			}
			defer testFile.Close()

			testReader, err := MakeIteratorFrom(testFile)
			if err != nil {
				t.Errorf("failed to create reader: %v", err)
			}
			for _, outcome := range testcase.outcomes {
				if !testReader.HasNext() {
					t.Error("expected outcome, got EOF")
				}
				repoURL, err := testReader.Next()
				if (err != nil) != outcome.hasError {
					t.Errorf("expected hasError: %t, got: %v", outcome.hasError, err)
				}

				if !outcome.hasError {
					if !cmp.Equal(outcome.repo, repoURL) {
						t.Errorf("expected equal, got diff: %s", cmp.Diff(outcome.repo, repoURL))
					}
				}
				if outcome.hasError && outcome.expectedErr != nil && !errors.Is(err, outcome.expectedErr) {
					t.Errorf("expected error: %v, got %v", outcome.expectedErr, err)
				}
			}
			if testReader.HasNext() {
				t.Error("actual reader has more repos than expected")
			}
		})
	}
}

//nolint:gocognit
func TestNestedIterator(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name      string
		filenames []string
		outcomes  []outcome
	}{
		{
			name:      "Multiple files",
			filenames: []string{"testdata/basic.csv", "testdata/split_file.csv"},
			outcomes: []outcome{
				{
					hasError: false,
					repo: RepoFormat{
						Repo: "github.com/owner1/repo1",
					},
				},
				{
					hasError: false,
					repo: RepoFormat{
						Repo: "github.com/owner2/repo2",
					},
				},
				{
					hasError: false,
					repo: RepoFormat{
						Repo:     "github.com/owner3/repo3",
						Metadata: []string{"meta"},
					},
				},
				{
					hasError: false,
					repo: RepoFormat{
						Repo: "github.com/owner4/repo4",
					},
				},
				{
					hasError: false,
					repo: RepoFormat{
						Repo: "github.com/owner5/repo5",
					},
				},
			},
		},
		{
			name:      "Empty first file",
			filenames: []string{"testdata/split_file_empty.csv", "testdata/basic.csv"},
			outcomes: []outcome{
				{
					hasError: false,
					repo: RepoFormat{
						Repo: "github.com/owner1/repo1",
					},
				},
				{
					hasError: false,
					repo: RepoFormat{
						Repo: "github.com/owner2/repo2",
					},
				},
				{
					hasError: false,
					repo: RepoFormat{
						Repo:     "github.com/owner3/repo3",
						Metadata: []string{"meta"},
					},
				},
			},
		},
		{
			name:      "one file",
			filenames: []string{"testdata/basic.csv"},
			outcomes: []outcome{
				{
					hasError: false,
					repo: RepoFormat{
						Repo: "github.com/owner1/repo1",
					},
				},
				{
					hasError: false,
					repo: RepoFormat{
						Repo: "github.com/owner2/repo2",
					},
				},
				{
					hasError: false,
					repo: RepoFormat{
						Repo:     "github.com/owner3/repo3",
						Metadata: []string{"meta"},
					},
				},
			},
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()
			var iters []Iterator
			for _, file := range testcase.filenames {
				testFile, err := os.Open(file)
				if err != nil {
					t.Errorf("failed to open %s: %v", file, err)
				}
				defer testFile.Close()
				iter, err := MakeIteratorFrom(testFile)
				if err != nil {
					t.Errorf("failed to create reader: %v", err)
				}
				iters = append(iters, iter)
			}

			testReader, err := MakeNestedIterator(iters)
			if err != nil {
				t.Errorf("failed to create reader: %v", err)
			}
			for _, outcome := range testcase.outcomes {
				if !testReader.HasNext() {
					t.Error("expected outcome, got EOF")
				}
				repoURL, err := testReader.Next()
				if (err != nil) != outcome.hasError {
					t.Errorf("expected hasError: %t, got: %v", outcome.hasError, err)
				}

				if !outcome.hasError {
					if !cmp.Equal(outcome.repo, repoURL) {
						t.Errorf("expected equal, got diff: %s", cmp.Diff(outcome.repo, repoURL))
					}
				}
				if outcome.hasError && outcome.expectedErr != nil && !errors.Is(err, outcome.expectedErr) {
					t.Errorf("expected error: %v, got %v", outcome.expectedErr, err)
				}
			}
			if testReader.HasNext() {
				t.Error("actual reader has more repos than expected")
			}
		})
	}
}
