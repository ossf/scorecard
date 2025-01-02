// Copyright 2024 OpenSSF Scorecard Authors
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

package azuredevopsrepo

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/servicehooks"

	"github.com/ossf/scorecard/v5/clients"
)

func Test_listWebhooks(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name              string
		listSubscriptions fnListSubscriptions
		want              []clients.Webhook
		wantErr           bool
	}{
		{
			name: "empty response",
			listSubscriptions: func(ctx context.Context, args servicehooks.ListSubscriptionsArgs) (*[]servicehooks.Subscription, error) {
				return &[]servicehooks.Subscription{}, nil
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "single webhook",
			listSubscriptions: func(ctx context.Context, args servicehooks.ListSubscriptionsArgs) (*[]servicehooks.Subscription, error) {
				return &[]servicehooks.Subscription{{ConsumerInputs: &map[string]string{"url": "https://example.com"}}}, nil
			},
			want:    []clients.Webhook{{Path: "https://example.com"}},
			wantErr: false,
		},
		{
			name: "multiple webhooks",
			listSubscriptions: func(ctx context.Context, args servicehooks.ListSubscriptionsArgs) (*[]servicehooks.Subscription, error) {
				return &[]servicehooks.Subscription{
					{ConsumerInputs: &map[string]string{"url": "https://example.com"}},
					{ConsumerInputs: &map[string]string{"url": "https://example2.com"}},
				}, nil
			},
			want:    []clients.Webhook{{Path: "https://example.com"}, {Path: "https://example2.com"}},
			wantErr: false,
		},
		{
			name: "with secret",
			listSubscriptions: func(ctx context.Context, args servicehooks.ListSubscriptionsArgs) (*[]servicehooks.Subscription, error) {
				return &[]servicehooks.Subscription{{ConsumerInputs: &map[string]string{"url": "https://example.com", "basicAuthPassword": "hunter2"}}}, nil
			},
			want:    []clients.Webhook{{Path: "https://example.com", UsesAuthSecret: true}},
			wantErr: false,
		},
		{
			name: "error",
			listSubscriptions: func(ctx context.Context, args servicehooks.ListSubscriptionsArgs) (*[]servicehooks.Subscription, error) {
				return nil, fmt.Errorf("error")
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := servicehooksHandler{
				ctx:               context.Background(),
				once:              new(sync.Once),
				repourl:           &Repo{},
				listSubscriptions: tt.listSubscriptions,
			}

			got, err := s.listWebhooks()
			if (err != nil) != tt.wantErr {
				t.Errorf("listWebhooks() error = %v, wantErr %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("listWebhooks() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
