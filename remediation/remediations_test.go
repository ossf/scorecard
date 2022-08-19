// Copyright 2022 Security Scorecard Authors
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

package remediation

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/ossf/scorecard/v4/checker"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
)

func TestRepeatedSetup(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	for i := 0; i < 2; i++ {
		mockRepo := mockrepo.NewMockRepoClient(ctrl)
		mockRepo.EXPECT().GetDefaultBranchName().Return("main", nil).AnyTimes()
		uri := fmt.Sprintf("github.com/ossf/scorecard%d", i)
		mockRepo.EXPECT().URI().Return(uri).AnyTimes()

		c := checker.CheckRequest{
			RepoClient: mockRepo,
		}
		rmd, err := New(&c)
		if err != nil {
			t.Error(err)
		}

		want := fmt.Sprintf("ossf/scorecard%d", i)
		if rmd.repo != want {
			t.Errorf("failed. expected: %v, got: %v", want, rmd.repo)
		}
	}
}
