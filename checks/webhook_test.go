// Copyright 2022 OpenSSF Scorecard Authors
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
	scut "github.com/ossf/scorecard/v4/utests"
)

//nolint:paralleltest // since t.Setenv is used
func TestWebhooks(t *testing.T) {
	tests := []struct {
		expected checker.CheckResult
		uri      string
		err      error
		name     string
		webhooks []clients.Webhook
	}{
		{
			name: "No Webhooks",
			uri:  "github.com/owner/repo",
			expected: checker.CheckResult{
				Score: 10,
			},
			err:      nil,
			webhooks: []clients.Webhook{},
		},
		{
			name: "With Webhooks and secret set",
			uri:  "github.com/owner/repo",
			expected: checker.CheckResult{
				Score: 10,
			},
			err: nil,
			webhooks: []clients.Webhook{
				{
					ID:             12345,
					UsesAuthSecret: true,
				},
			},
		},
		{
			name: "With Webhooks and no secret set",
			uri:  "github.com/owner/repo",
			expected: checker.CheckResult{
				Score: 0,
			},
			err: nil,
			webhooks: []clients.Webhook{
				{
					ID:             12345,
					UsesAuthSecret: false,
				},
			},
		},
		{
			name: "With 2 Webhooks with and whitout secrets configured",
			uri:  "github.com/owner/repo",
			expected: checker.CheckResult{
				Score: 5,
			},
			err: nil,
			webhooks: []clients.Webhook{
				{
					ID:             12345,
					UsesAuthSecret: false,
				},
				{
					ID:             54321,
					UsesAuthSecret: true,
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SCORECARD_EXPERIMENTAL", "true")
			ctrl := gomock.NewController(t)
			mockRepo := mockrepo.NewMockRepoClient(ctrl)

			mockRepo.EXPECT().ListWebhooks().DoAndReturn(func() ([]clients.Webhook, error) {
				if tt.err != nil {
					return nil, tt.err
				}
				return tt.webhooks, tt.err
			}).MaxTimes(1)

			mockRepo.EXPECT().URI().Return(tt.uri).AnyTimes()

			dl := scut.TestDetailLogger{}
			req := checker.CheckRequest{
				RepoClient: mockRepo,
				Ctx:        context.TODO(),
				Dlogger:    &dl,
			}
			res := WebHooks(&req)
			if tt.err != nil {
				if res.Error == nil {
					t.Errorf("Expected error %v, got nil", tt.err)
				}
				// return as we don't need to check the rest of the fields.
				return
			}

			if res.Score != tt.expected.Score {
				t.Errorf("Expected score %d, got %d for %v", tt.expected.Score, res.Score, tt.name)
			}
			ctrl.Finish()
		})
	}
}
