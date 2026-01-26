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
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	sce "github.com/ossf/scorecard/v5/errors"
	scut "github.com/ossf/scorecard/v5/utests"
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
			scut.ValidateTestReturn(t, tt.name, &tt.expected, &res, &dl)
			ctrl.Finish()
		})
	}
}

func TestCIIBestPractices_CustomURL_AllBadges(t *testing.T) {
	tests := []struct {
		name      string
		badgeJSON string
		expected  clients.BadgeLevel
	}{
		{
			"NotFoundBadge",
			`[]`,
			clients.NotFound,
		},
		{
			"InProgressBadge", `[{"badge_level":"in_progress"}]`,
			clients.InProgress,
		},
		{
			"PassingBadge", `[{"badge_level":"passing"}]`,
			clients.Passing,
		},
		{
			"SilverBadge", `[{"badge_level":"silver"}]`,
			clients.Silver,
		},
		{
			"GoldBadge", `[{"badge_level":"gold"}]`,
			clients.Gold,
		},
		{
			"UnknownBadge", `[{"badge_level":"foo"}]`,
			clients.Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/projects.json" {
					t.Errorf("Expected request path '/projects.json', got: %s", r.URL.Path)
				}
				fmt.Fprint(w, tt.badgeJSON)
			}))
			defer server.Close()

			t.Setenv("CII_BEST_PRACTICES_URL", server.URL)

			client := clients.DefaultCIIBestPracticesClient()
			badge, err := client.GetBadgeLevel(t.Context(), "github.com/owner/repo")
			if err != nil && tt.expected != clients.Unknown {
				t.Fatalf("GetBadgeLevel() returned unexpected error: %v", err)
			}

			if badge != tt.expected {
				t.Errorf("GetBadgeLevel() = %v, want %v", badge, tt.expected)
			}
		})
	}
}
