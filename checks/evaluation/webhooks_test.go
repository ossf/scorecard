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

package evaluation

import (
	"testing"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	scut "github.com/ossf/scorecard/v4/utests"
)

// TestWebhooks tests the webhooks check.
func TestWebhooks(t *testing.T) {
	t.Parallel()
	//nolint
	type args struct {
		name string
		dl   checker.DetailLogger
		r    *checker.WebhooksData
	}
	tests := []struct {
		name    string
		args    args
		want    checker.CheckResult
		wantErr bool
	}{
		{
			name: "r nil",
			args: args{
				name: "test_webhook_check_pass",
				dl:   &scut.TestDetailLogger{},
			},
			wantErr: true,
		},
		{
			name: "no webhooks",
			args: args{
				name: "no webhooks",
				dl:   &scut.TestDetailLogger{},
				r:    &checker.WebhooksData{},
			},
			want: checker.CheckResult{
				Score: checker.MaxResultScore,
			},
		},
		{
			name: "1 webhook with secret",
			args: args{
				name: "1 webhook with secret",
				dl:   &scut.TestDetailLogger{},
				r: &checker.WebhooksData{
					Webhooks: []clients.Webhook{
						{
							Path:           "https://github.com/owner/repo/settings/hooks/1234",
							ID:             1234,
							UsesAuthSecret: true,
						},
					},
				},
			},
			want: checker.CheckResult{
				Score: 10,
			},
		},
		{
			name: "1 webhook with no secret",
			args: args{
				name: "1 webhook with no secret",
				dl:   &scut.TestDetailLogger{},
				r: &checker.WebhooksData{
					Webhooks: []clients.Webhook{
						{
							Path:           "https://github.com/owner/repo/settings/hooks/1234",
							ID:             1234,
							UsesAuthSecret: false,
						},
					},
				},
			},
			want: checker.CheckResult{
				Score: 0,
			},
		},
		{
			name: "many webhooks with no secret and with secret",
			args: args{
				name: "many webhooks with no secret and with secret",
				dl:   &scut.TestDetailLogger{},
				r: &checker.WebhooksData{
					Webhooks: []clients.Webhook{
						{
							Path:           "https://github.com/owner/repo/settings/hooks/1234",
							ID:             1234,
							UsesAuthSecret: false,
						},
						{
							Path:           "https://github.com/owner/repo/settings/hooks/1111",
							ID:             1111,
							UsesAuthSecret: true,
						},
						{
							Path:           "https://github.com/owner/repo/settings/hooks/4444",
							ID:             4444,
							UsesAuthSecret: true,
						},
						{
							Path:           "https://github.com/owner/repo/settings/hooks/3333",
							ID:             3333,
							UsesAuthSecret: false,
						},
						{
							Path:           "https://github.com/owner/repo/settings/hooks/2222",
							ID:             2222,
							UsesAuthSecret: false,
						},
					},
				},
			},
			want: checker.CheckResult{
				Score: 6,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Webhooks(tt.args.name, tt.args.dl, tt.args.r)
			if tt.wantErr {
				if got.Error == nil {
					t.Errorf("Webhooks() error = %v, wantErr %v", got.Error, tt.wantErr)
				}
			} else {
				if got.Score != tt.want.Score {
					t.Errorf("Webhooks() = %v, want %v", got.Score, tt.want.Score)
				}
			}
		})
	}
}
