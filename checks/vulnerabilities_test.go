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

package checks

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
)

func TestVulnerabilities(t *testing.T) {
	t.Parallel()
	tests := []struct {
		err      error
		name     string
		expected clients.VulnerabilitiesResponse
		isError  bool
	}{
		{
			name:     "Valid response",
			isError:  false,
			expected: clients.VulnerabilitiesResponse{},
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockRepo := mockrepo.NewMockRepoClient(ctrl)

			mockRepo.EXPECT().ListCommits().DoAndReturn(func() ([]clients.Commit, error) {
				if tt.err != nil {
					return nil, tt.err
				}
				return []clients.Commit{{SHA: "test"}}, nil
			}).MinTimes(1)

			mockVulnClient := mockrepo.NewMockVulnerabilitiesClient(ctrl)
			mockVulnClient.EXPECT().HasUnfixedVulnerabilities(context.TODO(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, repo string) (clients.VulnerabilitiesResponse, error) {
					return tt.expected, tt.err
				}).MinTimes(1)

			req := checker.CheckRequest{
				RepoClient:            mockRepo,
				Ctx:                   context.TODO(),
				VulnerabilitiesClient: mockVulnClient,
			}
			res := Vulnerabilities(&req)
			if !tt.isError && res.Error != nil {
				t.Fail()
			} else if tt.isError && res.Error == nil {
				t.Fail()
			}
			ctrl.Finish()
		})
	}
}
