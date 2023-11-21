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

package raw

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
	sce "github.com/ossf/scorecard/v4/errors"
	scut "github.com/ossf/scorecard/v4/utests"
)

func TestWebhooks(t *testing.T) {
	t.Parallel()
	//nolint:govet
	tests := []struct {
		name                   string
		err                    error
		uri                    string
		wantErr                bool
		expectedUsesAuthSecret int
		expected               scut.TestReturn
		webhookResponse        []clients.Webhook
	}{
		{
			name:            "No Webhooks",
			wantErr:         false,
			webhookResponse: []clients.Webhook{},
		},
		{
			name:    "Error getting webhook",
			wantErr: true,
			err:     sce.ErrScorecardInternal,
		},
		{
			name:                   "Webhook with no secret",
			wantErr:                false,
			expectedUsesAuthSecret: 0,
			webhookResponse: []clients.Webhook{
				{
					UsesAuthSecret: false,
				},
			},
		},
		{
			name:                   "Webhook with secrets",
			wantErr:                false,
			expectedUsesAuthSecret: 2,
			webhookResponse: []clients.Webhook{
				{
					UsesAuthSecret: true,
				},
				{
					UsesAuthSecret: true,
				},
			},
		},
		{
			name:                   "Webhook with secrets and some without defined secrets",
			wantErr:                false,
			expectedUsesAuthSecret: 1,
			webhookResponse: []clients.Webhook{
				{
					UsesAuthSecret: true,
				},
				{
					UsesAuthSecret: false,
				},
				{
					UsesAuthSecret: false,
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockRepo := mockrepo.NewMockRepoClient(ctrl)

			mockRepo.EXPECT().URI().Return(tt.uri).AnyTimes()

			mockRepo.EXPECT().ListWebhooks().DoAndReturn(func() ([]clients.Webhook, error) {
				if tt.err != nil {
					return nil, tt.err
				}

				return tt.webhookResponse, nil
			}).AnyTimes()

			dl := scut.TestDetailLogger{}
			req := checker.CheckRequest{
				RepoClient: mockRepo,
				Ctx:        context.TODO(),
				Dlogger:    &dl,
			}
			got, err := WebHook(&req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Webhooks() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				gotHasSecret := 0
				for _, gotHook := range got.Webhooks {
					if gotHook.UsesAuthSecret {
						gotHasSecret++
					}
				}

				if gotHasSecret != tt.expectedUsesAuthSecret {
					t.Errorf("Webhooks() got = %v, want %v", gotHasSecret, tt.expectedUsesAuthSecret)
				}
			}

			if !scut.ValidateTestReturn(t, tt.name, &tt.expected, &checker.CheckResult{}, &dl) {
				t.Fatalf("Test %s failed", tt.name)
			}
		})
	}
}
