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
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/policy"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/security"

	"github.com/ossf/scorecard/v5/clients"
)

const (
	minimumReviewersPolicy  = "fa4e907d-c16b-4a4c-9dfa-4906e5d171dd"
	requiredReviewersPolicy = "fd2167ab-b0be-447a-8ec8-39368250530e"
	buildValidationPolicy   = "0609b952-1397-4640-95ec-e00a01b2c241"
	statusCheckPolicy       = "cbdc66da-9728-4af8-aada-9a5a32e4a226"
	workItemLinkPolicy      = "40e92b44-2fe1-4dd6-b3d8-74a9c21d0c6e"
	commentResolutionPolicy = "c6a1889d-b943-4856-b76f-9e46bb6b0df2"
	mergeStrategyPolicy     = "fa4e907d-c16b-4a4c-9dfa-4916e5d171ab"

	secretsScanningPolicy          = "90f9629b-664b-4804-a560-dd79b0c628f8" // #nosec G101 -- Public policy type UUID.
	alternateSecretsScanningPolicy = "ec003f37-8db0-4e10-992a-a2895045752c" // #nosec G101 -- Public policy type UUID.
	repositorySettingsPolicy       = "0517f88d-4ec5-4343-9d26-9930ebd53069"
	commitAuthorEmailPolicy        = "77ed4bd3-b063-4689-934a-175e4d0a78d7"
	fileNameRestrictionPolicy      = "51c78909-e838-41a2-9496-c647091e3c61"
	fileSizeRestrictionPolicy      = "2e26e725-8201-4edd-8bf5-978563c34a80"
	pathLengthRestrictionPolicy    = "001a79cf-fda1-4c4e-9e7c-bac40ee5ead8"
	reservedNamesRestrictionPolicy = "db2b9b4c-180d-4529-9701-01541d19f36b"
)

var errBranchProtectionData = errors.New("invalid Azure DevOps branch protection data")

type branchPolicySettings struct {
	MinimumApproverCount       *int32   `json:"minimumApproverCount"`
	CreatorVoteCounts          *bool    `json:"creatorVoteCounts"`
	ResetOnSourcePush          *bool    `json:"resetOnSourcePush"`
	BlockLastPusherVote        *bool    `json:"blockLastPusherVote"`
	RequireVoteOnLastIteration *bool    `json:"requireVoteOnLastIteration"`
	RequireVoteOnEachIteration *bool    `json:"requireVoteOnEachIteration"`
	BuildDefinitionID          *int32   `json:"buildDefinitionId"`
	QueueOnSourceUpdateOnly    *bool    `json:"queueOnSourceUpdateOnly"`
	ValidDuration              *int32   `json:"validDuration"`
	StatusName                 *string  `json:"statusName"`
	StatusGenre                *string  `json:"statusGenre"`
	FilenamePatterns           []string `json:"filenamePatterns"`
}

type policyRequirement struct {
	required bool
	unknown  bool
}

func (r *policyRequirement) add(value *bool) {
	if value == nil {
		r.unknown = true
	} else if *value {
		r.required = true
	}
}

func (r *policyRequirement) value() *bool {
	if !r.required && r.unknown {
		return nil
	}
	return &r.required
}

type branchesHandler struct {
	gitClient               git.Client
	securityClient          security.Client
	ctx                     context.Context
	once                    *sync.Once
	errSetup                error
	repourl                 *Repo
	defaultBranchRef        *clients.BranchRef
	queryBranch             fnQueryBranch
	getPolicyConfigurations fnGetPolicyConfigurations
	hasPermissionsBatch     fnHasPermissionsBatch
}

func (b *branchesHandler) init(ctx context.Context, repourl *Repo) {
	b.ctx = ctx
	b.repourl = repourl
	b.errSetup = nil
	b.once = new(sync.Once)
	b.defaultBranchRef = nil
	b.queryBranch = b.gitClient.GetBranch
	b.getPolicyConfigurations = b.gitClient.GetPolicyConfigurations
	b.hasPermissionsBatch = b.securityClient.HasPermissionsBatch
}

type (
	fnQueryBranch             func(ctx context.Context, args git.GetBranchArgs) (*git.GitBranchStats, error)
	fnGetPolicyConfigurations func(
		ctx context.Context,
		args git.GetPolicyConfigurationsArgs,
	) (*git.GitPolicyConfigurationResponse, error)
	fnHasPermissionsBatch func(
		context.Context,
		security.HasPermissionsBatchArgs,
	) (*security.PermissionEvaluationBatch, error)
)

func (b *branchesHandler) setup() error {
	b.once.Do(func() {
		b.defaultBranchRef, b.errSetup = b.loadBranch(b.repourl.defaultBranch)
	})
	return b.errSetup
}

func (b *branchesHandler) getDefaultBranch() (*clients.BranchRef, error) {
	if err := b.setup(); err != nil {
		return nil, fmt.Errorf("error during branchesHandler.setup: %w", err)
	}

	return b.defaultBranchRef, nil
}

func (b *branchesHandler) getBranch(branchName string) (*clients.BranchRef, error) {
	branchName = strings.TrimPrefix(branchName, "refs/heads/")
	if branchName == strings.TrimPrefix(b.repourl.defaultBranch, "refs/heads/") {
		return b.getDefaultBranch()
	}
	return b.loadBranch(branchName)
}

func (b *branchesHandler) loadBranch(branchName string) (*clients.BranchRef, error) {
	branchName = strings.TrimPrefix(branchName, "refs/heads/")
	branch, err := b.queryBranch(b.ctx, git.GetBranchArgs{
		RepositoryId: &b.repourl.id,
		Project:      &b.repourl.project,
		Name:         &branchName,
	})
	if err != nil {
		return nil, fmt.Errorf("request for branch %s failed with error %w", branchName, err)
	}
	if branch == nil || branch.Name == nil || strings.TrimPrefix(*branch.Name, "refs/heads/") != branchName ||
		branchName == "" {
		return nil, fmt.Errorf("%w: missing or mismatched branch %s", errBranchProtectionData, branchName)
	}

	refName := "refs/heads/" + branchName
	repositoryID, err := uuid.Parse(b.repourl.id)
	if err != nil {
		return nil, fmt.Errorf("error parsing repository ID %s: %w", b.repourl.id, err)
	}
	policies, err := b.branchPolicies(repositoryID, refName)
	if err != nil {
		return nil, fmt.Errorf("request for policies on branch %s failed: %w", branchName, err)
	}
	ref, err := branchProtectionFromPolicies(branchName, policies)
	if err != nil {
		return nil, fmt.Errorf("mapping branch %s: %w", branchName, err)
	}
	if err := b.applyBranchPermissions(ref); err != nil {
		return nil, err
	}
	return ref, nil
}

func (b *branchesHandler) branchPolicies(repositoryID uuid.UUID, refName string) ([]policy.PolicyConfiguration, error) {
	var policies []policy.PolicyConfiguration
	var continuation *string
	seen := make(map[string]bool)
	for {
		response, err := b.getPolicyConfigurations(b.ctx, git.GetPolicyConfigurationsArgs{
			Project: &b.repourl.project, RepositoryId: &repositoryID,
			RefName: &refName, ContinuationToken: continuation,
		})
		if err != nil {
			return nil, fmt.Errorf("getting applicable policies: %w", err)
		}
		if response == nil || response.PolicyConfigurations == nil {
			return nil, fmt.Errorf("%w: missing policy response", errBranchProtectionData)
		}
		policies = append(policies, (*response.PolicyConfigurations)...)
		continuation = response.ContinuationToken
		if continuation == nil || *continuation == "" {
			return policies, nil
		}
		if seen[*continuation] {
			return nil, fmt.Errorf("%w: repeated policy continuation token", errBranchProtectionData)
		}
		seen[*continuation] = true
	}
}

type branchPolicyData struct {
	contexts                           []string
	stale, lastPush, checks, upToDate  policyRequirement
	floor                              int32
	protected, unknown, unknownReviews bool
}

func branchProtectionFromPolicies(name string, policies []policy.PolicyConfiguration) (*clients.BranchRef, error) {
	var data branchPolicyData
	for i := range policies {
		p := &policies[i]
		if err := data.addPolicy(p); err != nil {
			id := 0
			if p.Id != nil {
				id = *p.Id
			}
			return nil, fmt.Errorf("policy %d: %w", id, err)
		}
	}
	protected, required := &data.protected, data.protected
	requiredRef := &required
	if !data.protected && data.unknown {
		protected, requiredRef = nil, nil
	}
	reviews := &data.floor
	if data.floor == 0 && data.unknownReviews {
		reviews = nil
	}
	slices.Sort(data.contexts)
	return &clients.BranchRef{
		Name: &name, Protected: protected,
		BranchProtectionRule: clients.BranchProtectionRule{
			PullRequestRule: clients.PullRequestRule{
				Required: requiredRef, RequiredApprovingReviewCount: reviews, DismissStaleReviews: data.stale.value(),
			},
			RequireLastPushApproval: data.lastPush.value(),
			CheckRules: clients.StatusChecksRule{
				RequiresStatusChecks: data.checks.value(), UpToDateBeforeMerge: data.upToDate.value(),
				Contexts: slices.Compact(data.contexts),
			},
		},
	}, nil
}

func (d *branchPolicyData) addPolicy(p *policy.PolicyConfiguration) error {
	if p.Type == nil || p.Type.Id == nil {
		return fmt.Errorf("%w: missing policy type", errBranchProtectionData)
	}
	kind := p.Type.Id.String()
	switch kind {
	case secretsScanningPolicy, alternateSecretsScanningPolicy, repositorySettingsPolicy,
		commitAuthorEmailPolicy, fileNameRestrictionPolicy, fileSizeRestrictionPolicy,
		pathLengthRestrictionPolicy, reservedNamesRestrictionPolicy:
		return nil // Repository push restrictions are not PR policies.
	}
	if p.IsEnabled == nil || p.IsBlocking == nil {
		return fmt.Errorf("%w: missing enforcement flags", errBranchProtectionData)
	}
	if !*p.IsEnabled || !*p.IsBlocking || (p.IsDeleted != nil && *p.IsDeleted) {
		return nil
	}
	switch kind {
	case minimumReviewersPolicy, buildValidationPolicy, statusCheckPolicy:
	case requiredReviewersPolicy, workItemLinkPolicy, commentResolutionPolicy, mergeStrategyPolicy:
		d.protected = true
		return nil
	default:
		d.unknown, d.unknownReviews = true, true
		d.stale.add(nil)
		d.lastPush.add(nil)
		d.checks.add(nil)
		d.upToDate.add(nil)
		return nil
	}
	d.protected = true
	encoded, err := json.Marshal(p.Settings)
	if err != nil {
		return fmt.Errorf("encoding settings: %w", err)
	}
	var settings branchPolicySettings
	if err := json.Unmarshal(encoded, &settings); err != nil {
		return fmt.Errorf("decoding settings: %w", err)
	}
	if kind == minimumReviewersPolicy {
		return d.addReviewPolicy(&settings)
	}
	return d.addCheckPolicy(kind, &settings)
}

func (d *branchPolicyData) addReviewPolicy(s *branchPolicySettings) error {
	if s.MinimumApproverCount == nil || *s.MinimumApproverCount < 0 {
		return fmt.Errorf("%w: invalid reviewer count", errBranchProtectionData)
	}
	if len(s.FilenamePatterns) > 0 {
		d.unknownReviews = true
		d.stale.add(nil)
		d.lastPush.add(nil)
		return nil
	}
	if s.CreatorVoteCounts == nil {
		d.unknownReviews = true
	} else {
		count := *s.MinimumApproverCount
		if *s.CreatorVoteCounts && count > 0 {
			count--
		}
		d.floor = max(d.floor, count)
	}
	d.stale.add(s.ResetOnSourcePush)
	var fresh policyRequirement
	fresh.add(s.ResetOnSourcePush)
	fresh.add(s.RequireVoteOnLastIteration)
	fresh.add(s.RequireVoteOnEachIteration)
	approval := s.BlockLastPusherVote
	if approval != nil && *approval {
		approval = fresh.value()
	}
	if *s.MinimumApproverCount == 0 {
		disabled := false
		approval = &disabled
	}
	d.lastPush.add(approval)
	return nil
}

func (d *branchPolicyData) addCheckPolicy(kind string, s *branchPolicySettings) error {
	var identifier string
	if kind == buildValidationPolicy {
		if s.BuildDefinitionID == nil || *s.BuildDefinitionID <= 0 ||
			(s.ValidDuration != nil && *s.ValidDuration < 0) {
			return fmt.Errorf("%w: invalid build settings", errBranchProtectionData)
		}
		identifier = fmt.Sprintf("azure-pipelines/%d", *s.BuildDefinitionID)
	} else {
		if s.StatusName == nil || strings.TrimSpace(*s.StatusName) == "" {
			return fmt.Errorf("%w: missing status name", errBranchProtectionData)
		}
		identifier = *s.StatusName
		if s.StatusGenre != nil && *s.StatusGenre != "" {
			identifier = *s.StatusGenre + "/" + identifier
		}
	}
	if len(s.FilenamePatterns) > 0 {
		d.checks.add(nil)
		d.upToDate.add(nil)
		return nil
	}
	d.checks.required = true
	d.contexts = append(d.contexts, identifier)
	if kind == buildValidationPolicy {
		var immediate *bool
		if s.QueueOnSourceUpdateOnly != nil {
			if *s.QueueOnSourceUpdateOnly {
				expiring := false
				immediate = &expiring
			} else if s.ValidDuration != nil {
				expiring := *s.ValidDuration == 0
				immediate = &expiring
			}
		}
		d.upToDate.add(immediate)
	}
	return nil
}
