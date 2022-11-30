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

package raw

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v4/checker"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
	scut "github.com/ossf/scorecard/v4/utests"
)

func Test_Packaging(t *testing.T) {
	t.Parallel()

	//nolint
	tests := []struct {
		name          string
		filenames     []string
		err           error
		ecosystemName string
	}{
		{
			name:          "Go project root dir",
			filenames:     []string{"go.mod"},
			ecosystemName: string(checker.PackageEcosystemGo),
		},
		{
			name:      "Go project not root dir",
			filenames: []string{"path/go.mod"},
		},
		// TODO(2501): add unit tests for packaging.
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
			mockRepoClient.EXPECT().ListFiles(gomock.Any()).Return(tt.filenames, nil).AnyTimes()
			mockRepoClient.EXPECT().GetFileContent(gomock.Any()).Return(nil, nil).AnyTimes()

			mockRepo := mockrepo.NewMockRepo(ctrl)
			mockRepo.EXPECT().URI().Return("github.com/repo/name").AnyTimes()

			dl := scut.TestDetailLogger{}
			req := checker.CheckRequest{
				RepoClient: mockRepoClient,
				Ctx:        context.TODO(),
				Repo:       mockRepo,
				Dlogger:    &dl,
			}

			pkgs, err := Packaging(&req)

			if !errCmp(err, tt.err) {
				t.Errorf(cmp.Diff(err, tt.err, cmpopts.EquateErrors()))
			}
			if tt.err != nil {
				return
			}

			expectedLength := 1
			if len(pkgs.Packages) != expectedLength {
				t.Errorf(cmp.Diff(len(pkgs.Packages), expectedLength))
			}

			if pkgs.Packages[0].Ecosystem == nil && tt.ecosystemName != "" {
				t.Errorf("no ecosystem found")
			}

			if pkgs.Packages[0].Ecosystem != nil &&
				string(*pkgs.Packages[0].Ecosystem) != tt.ecosystemName {
				t.Errorf(cmp.Diff(*pkgs.Packages[0].Ecosystem, tt.ecosystemName))
			}
		})
	}
}
