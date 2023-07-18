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
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
)

func TestNew(t *testing.T) {
	r := New()
	if len(r.enabledChecks) == 0 {
		t.Errorf("runner has no checks to run: %v", r.enabledChecks)
	}
}

func TestRunner_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := mockrepo.NewMockRepoClient(ctrl)
	commit := []clients.Commit{{SHA: "foo"}}
	mockRepo.EXPECT().ListCommits().Return(commit, nil)
	mockRepo.EXPECT().InitRepo(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mockRepo.EXPECT().GetDefaultBranchName().Return("main", nil)
	mockRepo.EXPECT().Close().Return(nil)
	r := Runner{
		enabledChecks: checker.CheckNameToFnMap{},
		repoClient:    mockRepo,
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
