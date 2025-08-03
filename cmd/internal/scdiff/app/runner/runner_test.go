// Copyright 2023 OpenSSF Scorecard Authors
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

package runner

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/clients"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	"github.com/ossf/scorecard/v5/internal/checknames"
)

func TestNew(t *testing.T) {
	t.Parallel()
	requestedChecks := []string{"Code-Review"}
	r := New(requestedChecks)
	if len(r.enabledChecks) != len(requestedChecks) {
		t.Errorf("requested %d checks but only got: %v", len(requestedChecks), r.enabledChecks)
	}
}

func TestRunner_Run(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	mockRepo := mockrepo.NewMockRepoClient(ctrl)
	commit := []clients.Commit{{SHA: "foo"}}
	mockRepo.EXPECT().ListCommits().Return(commit, nil)
	mockRepo.EXPECT().ListFiles(gomock.Any()).Return(nil, nil)
	mockRepo.EXPECT().InitRepo(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mockRepo.EXPECT().GetDefaultBranchName().Return("main", nil)
	mockRepo.EXPECT().Close().Return(nil)
	mockRepo.EXPECT().GetFileReader(gomock.Any()).Return(nil, errors.New("reading files unsupported for this test")).AnyTimes()
	mockRepo.EXPECT().LocalPath().Return(".", nil)
	r := Runner{
		ctx: context.Background(),
		// use a check which works locally, but we declare no files above so no-op
		enabledChecks: []string{checknames.BinaryArtifacts},
		githubClient:  mockRepo,
	}
	const repo = "github.com/foo/bar"
	result, err := r.Run(repo)
	if err != nil {
		t.Errorf("unexpected test error: %v", err)
	}
	if result.Repo.Name != repo {
		t.Errorf("got: %v, wanted: %v", result.Repo.Name, repo)
	}
}
