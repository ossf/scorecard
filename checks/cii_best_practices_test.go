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

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
	sce "github.com/ossf/scorecard/v4/errors"
	scut "github.com/ossf/scorecard/v4/utests"
)

var errTest = errors.New("test error")

func TestCIIBestPractices(t *testing.T) {
	t.Parallel()
	tests := []struct {
		err        error
		name       string
		uri        string
		expected   scut.TestReturn
		badgeLevel clients.BadgeLevel
	}{
		{
			name:       "CheckURIUsed",
			uri:        "github.com/owner/repo",
			badgeLevel: clients.NotFound,
			expected: scut.TestReturn{
				Score: checker.MinResultScore,
			},
		},
		{
			name: "CheckErrorHandling",
			err:  errTest,
			expected: scut.TestReturn{
				Score: -1,
				Error: sce.ErrScorecardInternal,
			},
		},
		{
			name:       "NotFoundBadge",
			badgeLevel: clients.NotFound,
			expected: scut.TestReturn{
				Score: checker.MinResultScore,
			},
		},
		{
			name:       "InProgressBadge",
			badgeLevel: clients.InProgress,
			expected: scut.TestReturn{
				Score: 2,
			},
		},
		{
			name:       "PassingBadge",
			badgeLevel: clients.Passing,
			expected: scut.TestReturn{
				Score: 5,
			},
		},
		{
			name:       "SilverBadge",
			badgeLevel: clients.Silver,
			expected: scut.TestReturn{
				Score: 7,
			},
		},
		{
			name:       "GoldBadge",
			badgeLevel: clients.Gold,
			expected: scut.TestReturn{
				Score: checker.MaxResultScore,
			},
		},
		{
			name:       "UnknownBadge",
			badgeLevel: clients.Unknown,
			expected: scut.TestReturn{
				Score: -1,
				Error: sce.ErrScorecardInternal,
			},
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockRepo := mockrepo.NewMockRepo(ctrl)
			mockRepo.EXPECT().URI().Return(tt.uri).AnyTimes()

			mockCIIClient := mockrepo.NewMockCIIBestPracticesClient(ctrl)
			mockCIIClient.EXPECT().GetBadgeLevel(gomock.Any(), tt.uri).DoAndReturn(
				func(context.Context, string) (clients.BadgeLevel, error) {
					return tt.badgeLevel, tt.err
				}).MinTimes(1)

			req := checker.CheckRequest{
				Repo:      mockRepo,
				CIIClient: mockCIIClient,
			}
			res := CIIBestPractices(&req)
			dl := scut.TestDetailLogger{}
			if !scut.ValidateTestReturn(t, tt.name, &tt.expected, &res, &dl) {
				t.Fail()
			}
			ctrl.Finish()
		})
	}
}
