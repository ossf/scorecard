// Copyright 2025 OpenSSF Scorecard Authors
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
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
)

func TestSASTWithStatuses(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepoClient := mockrepo.NewMockRepoClient(ctrl)

	mergedAt := time.Now().Add(time.Hour * time.Duration(-1))

	// Prow job with govulncheck in status context name
	mockRepoClient.EXPECT().ListCommits().Return([]clients.Commit{
		{
			AssociatedMergeRequest: clients.PullRequest{
				Number:   1,
				HeadSHA:  "sha1",
				MergedAt: mergedAt,
			},
		},
	}, nil)

	// No CheckRuns
	mockRepoClient.EXPECT().ListCheckRunsForRef("sha1").Return([]clients.CheckRun{}, nil)

	// But has Statuses with govulncheck
	mockRepoClient.EXPECT().ListStatuses("sha1").Return([]clients.Status{
		{
			State:   "success",
			Context: "govulncheck_myproject",
			URL:     "https://prow.example.com/view/govulncheck",
		},
	}, nil)

	// No workflow files
	mockRepoClient.EXPECT().ListFiles(gomock.Any()).Return([]string{}, nil).AnyTimes()

	req := &checker.CheckRequest{
		RepoClient: mockRepoClient,
	}

	result, err := SAST(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Commits) != 1 {
		t.Errorf("expected 1 commit, got %d", len(result.Commits))
	}

	if !result.Commits[0].Compliant {
		t.Error("expected commit to be compliant (SAST tool detected in status)")
	}
}
