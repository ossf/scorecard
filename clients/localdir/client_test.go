// Copyright 2020 Security Scorecard Authors
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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.uber.org/zap/zapcore"

	"github.com/ossf/scorecard/v3/clients/githubrepo"
)

func TestClient_ListFiles(t *testing.T) {
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
			inputFolder: "/invalid/fullpath",
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
			logger, err := githubrepo.NewLogger(zapcore.DebugLevel)
			if err != nil {
				t.Errorf("githubrepo.NewLogger: %w", err)
			}
			// nolint
			defer logger.Sync() // Flushes buffer, if any.

			// Create repo.
			repo, err := MakeLocalDirRepo(tt.inputFolder)
			if !errors.Is(err, tt.err) {
				t.Errorf("MakeLocalDirRepo: %v, expected %v", err, tt.err)
			}

			if err != nil {
				return
			}

			client := CreateLocalDirClient(ctx, logger)
			if err := client.InitRepo(repo); err != nil {
				t.Errorf("InitRepo: %w", err)
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
