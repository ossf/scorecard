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

package checks

import (
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
)

func TestMaintainerResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		err       error
		issues    []clients.Issue
		wantScore int
		wantErr   bool
	}{
		{
			name:      "Error fetching issues",
			err:       errors.New("fetch error"),
			issues:    nil,
			wantScore: -1,
			wantErr:   true,
		},
		{
			name:      "No issues",
			err:       nil,
			issues:    []clients.Issue{},
			wantScore: checker.MaxResultScore,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mockrepo.NewMockRepoClient(ctrl)
			mockRepo.EXPECT().ListIssuesWithHistory().Return(tt.issues, tt.err).AnyTimes()

			req := &checker.CheckRequest{
				RepoClient: mockRepo,
			}

			result := MaintainerResponse(req)

			if (result.Error != nil) != tt.wantErr {
				t.Errorf("MaintainerResponse() error = %v, wantErr %v", result.Error, tt.wantErr)
			}
			if tt.wantScore >= 0 && result.Score != tt.wantScore {
				t.Errorf("MaintainerResponse() score = %v, want %v", result.Score, tt.wantScore)
			}
		})
	}
}
