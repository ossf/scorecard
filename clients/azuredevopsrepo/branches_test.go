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
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/policy"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/security"

	"github.com/ossf/scorecard/v5/clients"
)

func Test_getDefaultBranch(t *testing.T) {
	t.Parallel()
	tests := []struct {
		queryBranch   fnQueryBranch
		name          string
		expectedName  string
		expectedError bool
	}{
		{
			name: "successful branch retrieval",
			queryBranch: func(ctx context.Context, args git.GetBranchArgs) (*git.GitBranchStats, error) {
				return &git.GitBranchStats{Name: args.Name}, nil
			},
			expectedError: false,
			expectedName:  "main",
		},
		{
			name: "error during branch retrieval",
			queryBranch: func(ctx context.Context, args git.GetBranchArgs) (*git.GitBranchStats, error) {
				return nil, fmt.Errorf("error")
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			handler := &branchesHandler{
				queryBranch:         tt.queryBranch,
				hasPermissionsBatch: deniedBranchPermissions,
				once:                new(sync.Once),
				repourl: &Repo{
					id:            uuid.Nil.String(),
					projectID:     uuid.Nil.String(),
					defaultBranch: "main",
				},
				getPolicyConfigurations: func(context.Context, git.GetPolicyConfigurationsArgs) (*git.GitPolicyConfigurationResponse, error) {
					return &git.GitPolicyConfigurationResponse{PolicyConfigurations: &[]policy.PolicyConfiguration{}}, nil
				},
			}

			branch, err := handler.getDefaultBranch()
			if (err != nil) != tt.expectedError {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			}
			if branch != nil && *branch.Name != tt.expectedName {
				t.Errorf("expected branch name: %v, got: %v", tt.expectedName, *branch.Name)
			}
		})
	}
}

func TestAzureBranchProtectionData(t *testing.T) {
	t.Parallel()
	policies := []policy.PolicyConfiguration{
		{
			Type:      &policy.PolicyTypeRef{Id: toPtr(uuid.MustParse("fa4e907d-c16b-4a4c-9dfa-4906e5d171dd"))},
			IsEnabled: toPtr(true), IsBlocking: toPtr(true),
			Settings: map[string]interface{}{
				"minimumApproverCount": 2, "creatorVoteCounts": false,
				"resetOnSourcePush": true, "blockLastPusherVote": true,
			},
		},
		{
			Type:      &policy.PolicyTypeRef{Id: toPtr(uuid.MustParse("0609b952-1397-4640-95ec-e00a01b2c241"))},
			IsEnabled: toPtr(true), IsBlocking: toPtr(true),
			Settings: map[string]interface{}{
				"buildDefinitionId": 50, "queueOnSourceUpdateOnly": false,
				"validDuration": 0, "manualQueueOnly": true,
			},
		},
	}
	client, calls := newBranchTestClient(t, policies)
	want := &clients.BranchRef{
		Name: toPtr("main"), Protected: toPtr(true),
		BranchProtectionRule: clients.BranchProtectionRule{
			PullRequestRule: clients.PullRequestRule{
				Required: toPtr(true), RequiredApprovingReviewCount: toPtr(int32(2)),
				DismissStaleReviews: toPtr(true),
			},
			RequireLastPushApproval: toPtr(true),
			CheckRules: clients.StatusChecksRule{
				RequiresStatusChecks: toPtr(true), UpToDateBeforeMerge: toPtr(true),
				Contexts: []string{"azure-pipelines/50"},
			},
		},
	}
	got, err := client.GetDefaultBranch()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("default branch mismatch (-want +got):\n%s", diff)
	}
	named, err := client.GetBranch("refs/heads/main")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, named); diff != "" {
		t.Errorf("named branch mismatch (-want +got):\n%s", diff)
	}
	if calls.Load() != 1 {
		t.Errorf("policy reads = %d, want one shared snapshot", calls.Load())
	}
}

func newBranchTestClient(t *testing.T, policies []policy.PolicyConfiguration) (*Client, *atomic.Int32) {
	t.Helper()
	calls := new(atomic.Int32)
	repo := &Repo{id: uuid.Nil.String(), projectID: uuid.Nil.String(), project: "project", defaultBranch: "main"}
	handler := &branchesHandler{
		ctx: t.Context(), once: new(sync.Once),
		repourl:             repo,
		hasPermissionsBatch: deniedBranchPermissions,
		queryBranch: func(_ context.Context, args git.GetBranchArgs) (*git.GitBranchStats, error) {
			return &git.GitBranchStats{Name: toPtr(strings.TrimPrefix(*args.Name, "refs/heads/"))}, nil
		},
		getPolicyConfigurations: func(_ context.Context, args git.GetPolicyConfigurationsArgs) (*git.GitPolicyConfigurationResponse, error) {
			calls.Add(1)
			if args.RefName == nil || *args.RefName != "refs/heads/"+repo.defaultBranch ||
				args.Project == nil || *args.Project != "project" ||
				args.RepositoryId == nil || *args.RepositoryId != uuid.Nil {
				t.Errorf("unexpected policy scope: %+v", args)
			}
			return &git.GitPolicyConfigurationResponse{PolicyConfigurations: &policies}, nil
		},
	}
	return &Client{branches: handler}, calls
}

func Test_getBranch(t *testing.T) {
	t.Parallel()
	tests := []struct {
		getBranch               fnQueryBranch
		getPolicyConfigurations fnGetPolicyConfigurations
		expected                *clients.BranchRef
		name                    string
		branchName              string
		expectedError           bool
	}{
		{
			name:       "successful branch retrieval",
			branchName: "main",
			getBranch: func(ctx context.Context, args git.GetBranchArgs) (*git.GitBranchStats, error) {
				return &git.GitBranchStats{Name: args.Name}, nil
			},
			getPolicyConfigurations: func(ctx context.Context, args git.GetPolicyConfigurationsArgs) (*git.GitPolicyConfigurationResponse, error) {
				return &git.GitPolicyConfigurationResponse{
					PolicyConfigurations: &[]policy.PolicyConfiguration{},
				}, nil
			},
			expected: &clients.BranchRef{
				Name:      toPtr("main"),
				Protected: toPtr(false),
				BranchProtectionRule: clients.BranchProtectionRule{
					PullRequestRule: clients.PullRequestRule{
						Required: toPtr(false), RequiredApprovingReviewCount: toPtr(int32(0)),
						DismissStaleReviews: toPtr(false),
					},
					RequireLastPushApproval: toPtr(false),
					CheckRules: clients.StatusChecksRule{
						RequiresStatusChecks: toPtr(false), UpToDateBeforeMerge: toPtr(false),
					},
				},
			},
			expectedError: false,
		},
		{
			name:       "error during branch retrieval",
			branchName: "main",
			getBranch: func(ctx context.Context, args git.GetBranchArgs) (*git.GitBranchStats, error) {
				return nil, fmt.Errorf("error")
			},
			getPolicyConfigurations: func(ctx context.Context, args git.GetPolicyConfigurationsArgs) (*git.GitPolicyConfigurationResponse, error) {
				return &git.GitPolicyConfigurationResponse{}, nil
			},
			expected:      nil,
			expectedError: true,
		},
		{
			name:       "error during policy configuration retrieval",
			branchName: "main",
			getBranch: func(ctx context.Context, args git.GetBranchArgs) (*git.GitBranchStats, error) {
				return &git.GitBranchStats{Name: args.Name}, nil
			},
			getPolicyConfigurations: func(ctx context.Context, args git.GetPolicyConfigurationsArgs) (*git.GitPolicyConfigurationResponse, error) {
				return nil, fmt.Errorf("error")
			},
			expected:      nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			handler := &branchesHandler{
				queryBranch:             tt.getBranch,
				getPolicyConfigurations: tt.getPolicyConfigurations,
				hasPermissionsBatch:     deniedBranchPermissions,
				once:                    new(sync.Once),
				repourl: &Repo{
					id:            uuid.Nil.String(),
					projectID:     uuid.Nil.String(),
					defaultBranch: "main",
				},
			}

			branch, err := handler.getBranch(tt.branchName)
			if (err != nil) != tt.expectedError {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			}
			if diff := cmp.Diff(branch, tt.expected); diff != "" {
				t.Errorf("mismatch in branch ref (-want +got):\n%s", diff)
			}
		})
	}
}

func TestAzureBranchPolicyLimits(t *testing.T) {
	t.Parallel()
	review := branchPolicyFixture("fa4e907d-c16b-4a4c-9dfa-4906e5d171dd", map[string]interface{}{
		"minimumApproverCount": 2, "creatorVoteCounts": true,
		"resetOnSourcePush": false, "blockLastPusherVote": true,
	})
	optional, deleted, disabled := review, review, review
	optional.IsBlocking, deleted.IsDeleted, disabled.IsEnabled = toPtr(false), toPtr(true), toPtr(false)
	for _, tt := range []struct {
		reviews   *int32
		name      string
		policies  []policy.PolicyConfiguration
		protected bool
	}{
		{name: "self approval", policies: []policy.PolicyConfiguration{review}, protected: true, reviews: toPtr(int32(1))},
		{name: "optional", policies: []policy.PolicyConfiguration{optional}, reviews: toPtr(int32(0))},
		{name: "disabled", policies: []policy.PolicyConfiguration{disabled}, reviews: toPtr(int32(0))},
		{name: "deleted", policies: []policy.PolicyConfiguration{deleted}, reviews: toPtr(int32(0))},
		{name: "secrets restriction", policies: []policy.PolicyConfiguration{
			branchPolicyFixture("90f9629b-664b-4804-a560-dd79b0c628f8", nil),
		}, reviews: toPtr(int32(0))},
		{name: "conditional reviewer", policies: []policy.PolicyConfiguration{
			branchPolicyFixture("fa4e907d-c16b-4a4c-9dfa-4906e5d171dd", map[string]interface{}{
				"minimumApproverCount": 2, "creatorVoteCounts": false, "filenamePatterns": []string{"/src/*"},
			}),
		}, protected: true},
		{name: "unknown self vote", policies: []policy.PolicyConfiguration{
			branchPolicyFixture("fa4e907d-c16b-4a4c-9dfa-4906e5d171dd", map[string]interface{}{"minimumApproverCount": 2}),
		}, protected: true},
		{name: "required group is not code owners", policies: []policy.PolicyConfiguration{
			branchPolicyFixture("fd2167ab-b0be-447a-8ec8-39368250530e", nil),
		}, protected: true, reviews: toPtr(int32(0))},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client, _ := newBranchTestClient(t, tt.policies)
			branch, err := client.GetDefaultBranch()
			if err != nil {
				t.Fatal(err)
			}
			if *branch.Protected != tt.protected {
				t.Errorf("protected = %t, want %t", *branch.Protected, tt.protected)
			}
			if diff := cmp.Diff(tt.reviews, branch.BranchProtectionRule.PullRequestRule.RequiredApprovingReviewCount); diff != "" {
				t.Errorf("independent review floor mismatch (-want +got):\n%s", diff)
			}
			if branch.BranchProtectionRule.EnforceAdmins != nil ||
				branch.BranchProtectionRule.PullRequestRule.RequireCodeOwnerReviews != nil ||
				branch.BranchProtectionRule.AllowForcePushes != nil {
				t.Errorf("unsupported protection inferred: %+v", branch.BranchProtectionRule)
			}
		})
	}
}

func branchPolicyFixture(kind string, settings interface{}) policy.PolicyConfiguration {
	return policy.PolicyConfiguration{
		Id: toPtr(10), Type: &policy.PolicyTypeRef{Id: toPtr(uuid.MustParse(kind))},
		IsEnabled: toPtr(true), IsBlocking: toPtr(true), Settings: settings,
	}
}

func deniedBranchPermissions(_ context.Context, args security.HasPermissionsBatchArgs) (*security.PermissionEvaluationBatch, error) {
	return branchPermissionValues(args, nil), nil
}

func branchPermissionValues(args security.HasPermissionsBatchArgs, values map[int]bool) *security.PermissionEvaluationBatch {
	evaluations := slices.Clone(*args.EvalBatch.Evaluations)
	for i := range evaluations {
		evaluations[i].Value = toPtr(values[*evaluations[i].Permissions])
	}
	return &security.PermissionEvaluationBatch{Evaluations: &evaluations}
}

func assertBranchPermissionRequest(t *testing.T, args security.HasPermissionsBatchArgs, token string) {
	t.Helper()
	if args.EvalBatch.AlwaysAllowAdministrators == nil || *args.EvalBatch.AlwaysAllowAdministrators {
		t.Error("administrator shortcut was not disabled")
	}
	for _, evaluation := range *args.EvalBatch.Evaluations {
		if evaluation.Token == nil || *evaluation.Token != token || evaluation.SecurityNamespaceId == nil ||
			*evaluation.SecurityNamespaceId != uuid.MustParse("2e9eb7ed-3c0a-47d4-87c1-0ffdd275fd87") {
			t.Errorf("wrong branch permission scope: %+v", evaluation)
		}
	}
}

func TestAzureBranchPermissionEvidence(t *testing.T) {
	t.Parallel()
	accessDenied := azuredevops.WrappedError{Message: toPtr("permission denied"), StatusCode: toPtr(403)}
	serverErr := azuredevops.WrappedError{Message: toPtr("server error"), StatusCode: toPtr(500)}
	for _, tt := range []struct {
		err                                                      error
		name                                                     string
		branch                                                   string
		protected, force, bypass, wantForce, wantDelete, wantErr bool
	}{
		{name: "caller denied", protected: true},
		{name: "allowed unprotected branch", force: true, wantForce: true, wantDelete: true},
		{name: "required policy without bypass", protected: true, force: true},
		{name: "required policy bypass", protected: true, force: true, bypass: true, wantForce: true},
		{name: "access denied", protected: true, err: accessDenied},
		{name: "server failure", err: serverErr, wantErr: true},
		{name: "malformed evaluation", wantErr: true},
		{name: "case folder and unicode", branch: "Main/\u6f22", force: true, wantForce: true, wantDelete: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var policies []policy.PolicyConfiguration
			if tt.protected {
				policies = append(policies, branchPolicyFixture("fa4e907d-c16b-4a4c-9dfa-4906e5d171dd", map[string]interface{}{
					"minimumApproverCount": 2, "creatorVoteCounts": false,
				}))
			}
			client, _ := newBranchTestClient(t, policies)
			encoded := "6d00610069006e00/"
			if tt.branch != "" {
				client.branches.repourl.defaultBranch = tt.branch
				encoded = "4d00610069006e00/226f/"
			}
			token := "repoV2/" + uuid.Nil.String() + "/" + uuid.Nil.String() + "/refs/heads/" + encoded
			var calls int
			client.branches.hasPermissionsBatch = func(_ context.Context, args security.HasPermissionsBatchArgs) (*security.PermissionEvaluationBatch, error) {
				calls++
				assertBranchPermissionRequest(t, args, token)
				if tt.err != nil {
					return nil, tt.err
				}
				response := branchPermissionValues(args, map[int]bool{4: true, 8: tt.force, 128: tt.bypass})
				if tt.name == "malformed evaluation" {
					(*response.Evaluations)[0].Token = toPtr("another branch")
				}
				return response, nil
			}
			branch, err := client.GetDefaultBranch()
			assertBranchRead(t, branch, err, tt.wantErr, tt.err)
			if calls != 1 {
				t.Errorf("permission requests = %d, want 1", calls)
			}
			if err != nil {
				return
			}
			assertCallerPermissions(t, branch, tt.protected, tt.wantForce, tt.wantDelete)
		})
	}
}

func TestAzureBranchPolicyReads(t *testing.T) {
	t.Parallel()
	denied := errors.New("policy access denied")
	for _, tt := range []struct {
		cause                             error
		settings                          interface{}
		name                              string
		repeat, missing, unknown, wantErr bool
	}{
		{name: "later inherited policy", settings: map[string]interface{}{
			"buildDefinitionId": 50, "queueOnSourceUpdateOnly": true, "validDuration": 60,
			"scope": []map[string]interface{}{{"repositoryId": nil, "refName": "refs/heads/", "matchKind": "Prefix"}},
		}},
		{name: "repeated token", repeat: true, wantErr: true},
		{name: "denied policies", cause: denied, wantErr: true},
		{name: "missing response", missing: true, wantErr: true},
		{name: "malformed settings", settings: "not an object", wantErr: true},
		{name: "invalid count", settings: map[string]interface{}{"minimumApproverCount": -1}, wantErr: true},
		{name: "overflow count", settings: map[string]interface{}{"minimumApproverCount": int64(1) << 33}, wantErr: true},
		{name: "unrecognized enforcement", unknown: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client, _ := newBranchTestClient(t, nil)
			var calls int
			client.branches.getPolicyConfigurations = func(_ context.Context, args git.GetPolicyConfigurationsArgs) (*git.GitPolicyConfigurationResponse, error) {
				calls++
				switch {
				case calls > 2:
					t.Fatal("policy traversal did not terminate")
				case tt.cause != nil || tt.missing:
					return nil, tt.cause
				case args.ContinuationToken == nil || tt.repeat:
					return &git.GitPolicyConfigurationResponse{
						PolicyConfigurations: &[]policy.PolicyConfiguration{},
						ContinuationToken:    toPtr("next"),
					}, nil
				}
				kind := "0609b952-1397-4640-95ec-e00a01b2c241"
				switch {
				case strings.Contains(tt.name, "count"):
					kind = "fa4e907d-c16b-4a4c-9dfa-4906e5d171dd"
				case tt.unknown:
					kind = uuid.Nil.String()
				}
				return &git.GitPolicyConfigurationResponse{PolicyConfigurations: &[]policy.PolicyConfiguration{
					branchPolicyFixture(kind, tt.settings),
				}}, nil
			}
			branch, err := client.GetDefaultBranch()
			assertBranchRead(t, branch, err, tt.wantErr, tt.cause)
			if err != nil {
				return
			}
			if tt.unknown {
				if branch.Protected != nil || branch.BranchProtectionRule.PullRequestRule.Required != nil {
					t.Errorf("unknown policy interpreted as known protection: %+v", branch)
				}
			} else if len(branch.BranchProtectionRule.CheckRules.Contexts) != 1 ||
				branch.BranchProtectionRule.CheckRules.UpToDateBeforeMerge == nil ||
				*branch.BranchProtectionRule.CheckRules.UpToDateBeforeMerge {
				t.Errorf("later-page delayed build policy was not mapped: %+v", branch)
			}
		})
	}
}

func assertBranchRead(t *testing.T, branch *clients.BranchRef, err error, wantErr bool, cause error) {
	t.Helper()
	if (err != nil) != wantErr {
		t.Fatalf("branch = %+v, error = %v, want error = %t", branch, err, wantErr)
	}
	if wantErr && cause != nil && !errors.Is(err, cause) {
		t.Errorf("lost branch request cause: %v", err)
	}
}

func assertCallerPermissions(t *testing.T, branch *clients.BranchRef, protected, force, deletion bool) {
	t.Helper()
	rules := branch.BranchProtectionRule
	if (rules.AllowForcePushes != nil) != force || (rules.AllowDeletions != nil) != deletion {
		t.Fatalf("wrong permission evidence: %+v", rules)
	}
	if (force && !*rules.AllowForcePushes) || (deletion && !*rules.AllowDeletions) {
		t.Error("an allowed operation was reported as globally prohibited")
	}
	if rules.EnforceAdmins != nil || *branch.Protected != protected {
		t.Errorf("permissions altered policy or administrator evidence: %+v", branch)
	}
}
