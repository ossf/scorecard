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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/policy"

	"github.com/ossf/scorecard/v5/clients"
)

func Test_listCheckRunsForRef(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                 string
		ref                  string
		getPullRequestQuery  fnGetPullRequestQuery
		getPolicyEvaluations fnGetPolicyEvaluations
		want                 []clients.CheckRun
		wantErr              bool
	}{
		{
			name: "happy path",
			ref:  "4b825dc642cb6eb9a060e54bf8d69288fbee4904",
			getPullRequestQuery: func(ctx context.Context, args git.GetPullRequestQueryArgs) (*git.GitPullRequestQuery, error) {
				return &git.GitPullRequestQuery{
					Results: &[]map[string][]git.GitPullRequest{
						{
							"4b825dc642cb6eb9a060e54bf8d69288fbee4904": {
								{
									PullRequestId: toPtr(1),
								},
							},
						},
					},
				}, nil
			},
			getPolicyEvaluations: func(ctx context.Context, args policy.GetPolicyEvaluationsArgs) (*[]policy.PolicyEvaluationRecord, error) {
				return &[]policy.PolicyEvaluationRecord{
					{
						Status: toPtr(policy.PolicyEvaluationStatusValues.Approved),
					},
				}, nil
			},
			want: []clients.CheckRun{
				{
					Status:     "completed",
					Conclusion: "success",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := policyHandler{
				ctx: t.Context(),
				repourl: &Repo{
					id: "1",
				},
				getPullRequestQuery:  tt.getPullRequestQuery,
				getPolicyEvaluations: tt.getPolicyEvaluations,
			}
			got, err := p.listCheckRunsForRef(tt.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("listCheckRunsForRef() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("listCheckRunsForRef() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
