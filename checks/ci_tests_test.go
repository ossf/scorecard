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

package checks

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/ossf/scorecard/v4/checker"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
)

func TestCITestsRuntimeError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
	//nolint:goerr113
	mockRepoClient.EXPECT().ListCommits().Return(nil, fmt.Errorf("some runtime error")).AnyTimes()

	req := checker.CheckRequest{
		RepoClient: mockRepoClient,
	}
	res := CITests(&req)

	want := "CI-Tests"
	if res.Name != want {
		t.Errorf("got: %q, want: %q", res.Name, want)
	}
	ctrl.Finish()
}
