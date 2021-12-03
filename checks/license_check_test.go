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

package checks

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"go.uber.org/zap/zapcore"

	"github.com/ossf/scorecard/v3/checker"
	"github.com/ossf/scorecard/v3/clients/githubrepo"
	"github.com/ossf/scorecard/v3/clients/localdir"
	scut "github.com/ossf/scorecard/v3/utests"
)

func TestLicenseFileCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
	}{
		{
			name:     "LICENSE.md",
			filename: "LICENSE.md",
		},
		{
			name:     "LICENSE",
			filename: "LICENSE",
		},
		{
			name:     "COPYING",
			filename: "COPYING",
		},
		{
			name:     "COPYING.md",
			filename: "COPYING.md",
		},
		{
			name:     "LICENSE.textile",
			filename: "LICENSE.textile",
		},
		{
			name:     "COPYING.textile",
			filename: "COPYING.textile",
		},
		{
			name:     "LICENSE-MIT",
			filename: "LICENSE-MIT",
		},
		{
			name:     "COPYING-MIT",
			filename: "COPYING-MIT",
		},
		{
			name:     "MIT-LICENSE-MIT",
			filename: "MIT-LICENSE-MIT",
		},
		{
			name:     "MIT-COPYING",
			filename: "MIT-COPYING",
		},
		{
			name:     "OFL.md",
			filename: "OFL.md",
		},
		{
			name:     "OFL.textile",
			filename: "OFL.textile",
		},
		{
			name:     "OFL",
			filename: "OFL",
		},
		{
			name:     "PATENTS",
			filename: "PATENTS",
		},
		{
			name:     "PATENTS.txt",
			filename: "PATENTS.txt",
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := testLicenseCheck(tt.filename)
			if !s {
				t.Fail()
			}
		})
	}
}

func TestLicenseFileSubdirectory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		inputFolder string
		err         error
		expected    scut.TestReturn
	}{
		{
			name:        "With LICENSE",
			inputFolder: "file://licensedir/withlicense",
			expected: scut.TestReturn{
				Error:        nil,
				Score:        checker.MaxResultScore,
				NumberOfInfo: 1,
			},
			err: nil,
		},
		{
			name:        "Without LICENSE",
			inputFolder: "file://licensedir/withoutlicense",
			expected: scut.TestReturn{
				Error: nil,
				Score: checker.MinResultScore,
			},
			err: nil,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger, err := githubrepo.NewLogger(zapcore.DebugLevel)
			if err != nil {
				t.Errorf("githubrepo.NewLogger: %v", err)
			}

			// nolint
			defer logger.Sync()

			ctrl := gomock.NewController(t)
			repo, err := localdir.MakeLocalDirRepo(tt.inputFolder)

			if !errors.Is(err, tt.err) {
				t.Errorf("MakeLocalDirRepo: %v, expected %v", err, tt.err)
			}

			ctx := context.Background()

			client := localdir.CreateLocalDirClient(ctx, logger)
			if err := client.InitRepo(repo); err != nil {
				t.Errorf("InitRepo: %v", err)
			}

			dl := scut.TestDetailLogger{}

			req := checker.CheckRequest{
				Ctx:        ctx,
				RepoClient: client,
				Dlogger:    &dl,
			}

			res := LicenseCheck(&req)

			if !scut.ValidateTestReturn(t, tt.name, &tt.expected, &res, &dl) {
				t.Fail()
			}

			ctrl.Finish()
		})
	}
}
