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

package webhooksUseSecrets

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/internal/utils/test"
)

func Test_Run(t *testing.T) {
	t.Parallel()
	//nolint:govet
	tests := []struct {
		name     string
		raw      *checker.RawResults
		outcomes []finding.Outcome
		err      error
	}{
		{
			name: "No Webhooks",
			raw: &checker.RawResults{
				WebhookResults: checker.WebhooksData{
					Webhooks: []clients.Webhook{},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeNotApplicable,
			},
		},
		{
			name: "Webhooks present with auth secret",
			raw: &checker.RawResults{
				WebhookResults: checker.WebhooksData{
					Webhooks: []clients.Webhook{
						{
							Path:           "https://github.com/owner/repo/settings/hooks/1234",
							ID:             1,
							UsesAuthSecret: true,
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue,
			},
		},
		{
			name: "Webhooks present without auth secret",
			raw: &checker.RawResults{
				WebhookResults: checker.WebhooksData{
					Webhooks: []clients.Webhook{
						{
							Path:           "https://github.com/owner/repo/settings/hooks/1234",
							ID:             1,
							UsesAuthSecret: false,
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeFalse,
			},
		},
		{
			name: "Multiple webhooks present, one without auth secret",
			raw: &checker.RawResults{
				WebhookResults: checker.WebhooksData{
					Webhooks: []clients.Webhook{
						{
							Path:           "https://github.com/owner/repo/settings/hooks/1234",
							ID:             1,
							UsesAuthSecret: false,
						},
						{
							Path:           "https://github.com/owner/repo/settings/hooks/12345",
							ID:             2,
							UsesAuthSecret: true,
						},
						{
							Path:           "https://github.com/owner/repo/settings/hooks/12346",
							ID:             3,
							UsesAuthSecret: true,
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeFalse, finding.OutcomeTrue, finding.OutcomeTrue,
			},
		},
		{
			name: "Multiple webhooks present, two without auth secret",
			raw: &checker.RawResults{
				WebhookResults: checker.WebhooksData{
					Webhooks: []clients.Webhook{
						{
							Path:           "https://github.com/owner/repo/settings/hooks/1234",
							ID:             1,
							UsesAuthSecret: false,
						},
						{
							Path:           "https://github.com/owner/repo/settings/hooks/12345",
							ID:             2,
							UsesAuthSecret: true,
						},
						{
							Path:           "https://github.com/owner/repo/settings/hooks/12346",
							ID:             3,
							UsesAuthSecret: true,
						},
						{
							Path:           "https://github.com/owner/repo/settings/hooks/12346",
							ID:             4,
							UsesAuthSecret: false,
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeFalse, finding.OutcomeTrue,
				finding.OutcomeTrue, finding.OutcomeFalse,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			findings, s, err := Run(tt.raw)
			if !cmp.Equal(tt.err, err, cmpopts.EquateErrors()) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(tt.err, err, cmpopts.EquateErrors()))
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(Probe, s); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
			test.AssertOutcomes(t, findings, tt.outcomes)
		})
	}
}
