// Copyright 2021 OpenSSF Scorecard Authors
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

package githubrepo

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/google/go-github/v53/github"
	"github.com/shurcooL/githubv4"

	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/clients/githubrepo/internal/fnmatch"
	sce "github.com/ossf/scorecard/v4/errors"
)

const (
	refPrefix = "refs/heads/"
)

// See https://github.community/t/graphql-api-protected-branch/14380
/* Example of query:
query {
  repository(owner: "laurentsimon", name: "test3") {
    branchProtectionRules(first: 100) {
      edges {
        node {
          allowsDeletions
          allowsForcePushes
          dismissesStaleReviews
          isAdminEnforced
          pattern
          matchingRefs(first: 100) {
            nodes {
              name
            }
          }
        }
      }
    }
    refs(first: 100, refPrefix: "refs/heads/") {
      nodes {
        name
        refUpdateRule {
          requiredApprovingReviewCount
          allowsForcePushes
        }
      }
    }
    rulesets(first: 100) {
      edges {
        node {
          name
          enforcement
          target
          conditions {
            refName {
              exclude
              include
            }
          }
          bypassActors(first: 100) {
            nodes {
              actor {
                __typename
                ... on App {
                  name
                  databaseId
                }
              }
              bypassMode
              organizationAdmin
              repositoryRoleName
            }
          }
          rules(first: 100) {
            nodes {
              type
              parameters {
                ... on PullRequestParameters {
                  dismissStaleReviewsOnPush
                  requireCodeOwnerReview
                  requireLastPushApproval
                  requiredApprovingReviewCount
                  requiredReviewThreadResolution
                }
                ... on RequiredStatusChecksParameters {
                  requiredStatusChecks {
                    context
                    integrationId
                  }
                  strictRequiredStatusChecksPolicy
                }
              }
            }
          }
        }
      }
    }
  }
}
*/

// Used for non-admin settings.
type refUpdateRule struct {
	AllowsDeletions              *bool
	AllowsForcePushes            *bool
	RequiredApprovingReviewCount *int32
	RequiresCodeOwnerReviews     *bool
	RequiresLinearHistory        *bool
	RequiredStatusCheckContexts  []string
}

// Used for all settings, both admin and non-admin ones.
// This only works with an admin token.
type branchProtectionRule struct {
	DismissesStaleReviews        *bool
	IsAdminEnforced              *bool
	RequiresStrictStatusChecks   *bool
	RequiresStatusChecks         *bool
	AllowsDeletions              *bool
	AllowsForcePushes            *bool
	RequiredApprovingReviewCount *int32
	RequiresCodeOwnerReviews     *bool
	RequiresLinearHistory        *bool
	RequireLastPushApproval      *bool
	RequiredStatusCheckContexts  []string
	// TODO: verify there is no conflicts.
	// BranchProtectionRuleConflicts interface{}
}

type branch struct {
	Name                 *string
	RefUpdateRule        *refUpdateRule
	BranchProtectionRule *branchProtectionRule
}
type defaultBranchData struct {
	Repository struct {
		DefaultBranchRef *branch
	} `graphql:"repository(owner: $owner, name: $name)"`
	RateLimit struct {
		Cost *int
	}
}

type pullRequestRuleParameters struct {
	DismissStaleReviewsOnPush      *bool
	RequireCodeOwnerReview         *bool
	RequireLastPushApproval        *bool
	RequiredApprovingReviewCount   *int32
	RequiredReviewThreadResolution *bool
}
type requiredStatusCheckParameters struct {
	StrictRequiredStatusChecksPolicy *bool
	RequiredStatusChecks             []statusCheck
}
type statusCheck struct {
	Context       *string
	IntegrationID *int64
}
type repoRule struct {
	Type       string
	Parameters repoRulesParameters
}
type repoRulesParameters struct {
	PullRequestParameters pullRequestRuleParameters     `graphql:"... on PullRequestParameters"`
	StatusCheckParameters requiredStatusCheckParameters `graphql:"... on RequiredStatusChecksParameters"`
}
type ruleSetConditionRefs struct {
	Include []string
	Exclude []string
}
type ruleSetCondition struct {
	RefName ruleSetConditionRefs
}
type ruleSetBypass struct {
	BypassMode         *string
	OrganizationAdmin  *bool
	RepositoryRoleName *string
}
type repoRuleSet struct {
	Name         *string
	Enforcement  *string
	Conditions   ruleSetCondition
	BypassActors struct {
		Nodes []*ruleSetBypass
	} `graphql:"bypassActors(first: 100)"`
	Rules struct {
		Nodes []*repoRule
	} `graphql:"rules(first: 100)"`
}
type ruleSetData struct {
	Repository struct {
		DefaultBranchRef struct {
			Name *string
		}
		Rulesets struct {
			Nodes []*repoRuleSet
		} `graphql:"rulesets(first: 100)"`
	} `graphql:"repository(owner: $owner, name: $name)"`
}

type branchData struct {
	Repository struct {
		Ref *branch `graphql:"ref(qualifiedName: $branchRefName)"`
	} `graphql:"repository(owner: $owner, name: $name)"`
}

type branchesHandler struct {
	ghClient          *github.Client
	graphClient       *githubv4.Client
	data              *defaultBranchData
	once              *sync.Once
	ctx               context.Context
	errSetup          error
	repourl           *repoURL
	defaultBranchRef  *clients.BranchRef
	defaultBranchName string
	ruleSets          []*repoRuleSet
}

func (handler *branchesHandler) init(ctx context.Context, repourl *repoURL) {
	handler.ctx = ctx
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
	handler.defaultBranchRef = nil
	handler.defaultBranchName = ""
	handler.ruleSets = nil
	handler.data = nil
}

func (handler *branchesHandler) setup() error {
	handler.once.Do(func() {
		if !strings.EqualFold(handler.repourl.commitSHA, clients.HeadSHA) {
			handler.errSetup = fmt.Errorf("%w: branches only supported for HEAD queries", clients.ErrUnsupportedFeature)
			return
		}
		vars := map[string]interface{}{
			"owner": githubv4.String(handler.repourl.owner),
			"name":  githubv4.String(handler.repourl.repo),
		}

		// Fetch default branch name and any repository rulesets, which are available with basic read permission.
		rulesData := new(ruleSetData)
		if err := handler.graphClient.Query(handler.ctx, rulesData, vars); err != nil {
			handler.errSetup = sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("githubv4.Query: %v", err))
			return
		}
		handler.defaultBranchName = getDefaultBranchNameFrom(rulesData)
		handler.ruleSets = getActiveRuleSetsFrom(rulesData)

		// Attempt to fetch branch protection rules, which require admin permission.
		// Ignore permissions errors if we know the repository is using rulesets, so non-admins can still get a score.
		handler.data = new(defaultBranchData)
		if err := handler.graphClient.Query(handler.ctx, handler.data, vars); err != nil &&
			(!isPermissionsError(err) || len(handler.ruleSets) == 0) {
			handler.errSetup = sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("githubv4.Query: %v", err))
			return
		}

		rules, err := rulesMatchingBranch(handler.ruleSets, handler.defaultBranchName, true)
		if err != nil {
			handler.errSetup = sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("rulesMatchingBranch: %v", err))
			return
		}
		handler.defaultBranchRef = getBranchRefFrom(handler.data.Repository.DefaultBranchRef, rules)
	})
	return handler.errSetup
}

func (handler *branchesHandler) query(branchName string) (*clients.BranchRef, error) {
	if !strings.EqualFold(handler.repourl.commitSHA, clients.HeadSHA) {
		return nil, fmt.Errorf("%w: branches only supported for HEAD queries", clients.ErrUnsupportedFeature)
	}
	// Call setup(), so we know if branchName == handler.defaultBranchName.
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during branchesHandler.setup: %w", err)
	}
	vars := map[string]interface{}{
		"owner":         githubv4.String(handler.repourl.owner),
		"name":          githubv4.String(handler.repourl.repo),
		"branchRefName": githubv4.String(refPrefix + branchName),
	}
	queryData := new(branchData)
	if err := handler.graphClient.Query(handler.ctx, queryData, vars); err != nil {
		return nil, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("githubv4.Query: %v", err))
	}
	rules, err := rulesMatchingBranch(handler.ruleSets, branchName, branchName == handler.defaultBranchName)
	if err != nil {
		return nil, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("rulesMatchingBranch: %v", err))
	}
	return getBranchRefFrom(queryData.Repository.Ref, rules), nil
}

func (handler *branchesHandler) getDefaultBranch() (*clients.BranchRef, error) {
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during branchesHandler.setup: %w", err)
	}
	return handler.defaultBranchRef, nil
}

func (handler *branchesHandler) getBranch(branch string) (*clients.BranchRef, error) {
	branchRef, err := handler.query(branch)
	if err != nil {
		return nil, fmt.Errorf("error during branchesHandler.query: %w", err)
	}
	return branchRef, nil
}

func copyAdminSettings(src *branchProtectionRule, dst *clients.BranchProtectionRule) {
	copyBoolPtr(src.IsAdminEnforced, &dst.EnforceAdmins)
	copyBoolPtr(src.RequireLastPushApproval, &dst.RequireLastPushApproval)
	copyBoolPtr(src.DismissesStaleReviews, &dst.RequiredPullRequestReviews.DismissStaleReviews)
	if src.RequiresStatusChecks != nil {
		copyBoolPtr(src.RequiresStatusChecks, &dst.CheckRules.RequiresStatusChecks)
		// TODO(#3255): Update when GitHub GraphQL bug is fixed
		// Workaround for GitHub GraphQL bug https://github.com/orgs/community/discussions/59471
		// The setting RequiresStrictStatusChecks should tell if the branch is required
		// to be up to date before merge, but it only returns the correct value if
		// RequiresStatusChecks is true. If RequiresStatusChecks is false, RequiresStrictStatusChecks
		// is wrongly retrieved as true.
		if src.RequiresStrictStatusChecks != nil {
			upToDateBeforeMerge := *src.RequiresStatusChecks && *src.RequiresStrictStatusChecks
			copyBoolPtr(&upToDateBeforeMerge, &dst.CheckRules.UpToDateBeforeMerge)
		}
	}
}

func copyNonAdminSettings(src interface{}, dst *clients.BranchProtectionRule) {
	// TODO: requiresConversationResolution, requiresSignatures, viewerAllowedToDismissReviews, viewerCanPush
	switch v := src.(type) {
	case *branchProtectionRule:
		copyBoolPtr(v.AllowsDeletions, &dst.AllowDeletions)
		copyBoolPtr(v.AllowsForcePushes, &dst.AllowForcePushes)
		copyBoolPtr(v.RequiresLinearHistory, &dst.RequireLinearHistory)
		copyInt32Ptr(v.RequiredApprovingReviewCount, &dst.RequiredPullRequestReviews.RequiredApprovingReviewCount)
		copyBoolPtr(v.RequiresCodeOwnerReviews, &dst.RequiredPullRequestReviews.RequireCodeOwnerReviews)
		copyStringSlice(v.RequiredStatusCheckContexts, &dst.CheckRules.Contexts)

	case *refUpdateRule:
		copyBoolPtr(v.AllowsDeletions, &dst.AllowDeletions)
		copyBoolPtr(v.AllowsForcePushes, &dst.AllowForcePushes)
		copyBoolPtr(v.RequiresLinearHistory, &dst.RequireLinearHistory)
		copyInt32Ptr(v.RequiredApprovingReviewCount, &dst.RequiredPullRequestReviews.RequiredApprovingReviewCount)
		copyBoolPtr(v.RequiresCodeOwnerReviews, &dst.RequiredPullRequestReviews.RequireCodeOwnerReviews)
		copyStringSlice(v.RequiredStatusCheckContexts, &dst.CheckRules.Contexts)
	}
}

func getDefaultBranchNameFrom(data *ruleSetData) string {
	if data == nil || data.Repository.DefaultBranchRef.Name == nil {
		return ""
	}
	return *data.Repository.DefaultBranchRef.Name
}

func getActiveRuleSetsFrom(data *ruleSetData) []*repoRuleSet {
	ret := make([]*repoRuleSet, 0)
	for _, rule := range data.Repository.Rulesets.Nodes {
		if rule.Enforcement == nil || *rule.Enforcement != "ACTIVE" {
			continue
		}
		ret = append(ret, rule)
	}
	return ret
}

func getBranchRefFrom(data *branch, rules []*repoRuleSet) *clients.BranchRef {
	if data == nil {
		return nil
	}
	branchRef := new(clients.BranchRef)
	if data.Name != nil {
		branchRef.Name = data.Name
	}

	// Protected means we found some data,
	// i.e., there's a rule for the branch.
	// It says nothing about what protection is enabled at all.
	branchRef.Protected = new(bool)
	if data.RefUpdateRule == nil &&
		data.BranchProtectionRule == nil &&
		len(rules) == 0 {
		*branchRef.Protected = false
		return branchRef
	}

	*branchRef.Protected = true
	branchRule := &branchRef.BranchProtectionRule

	switch {
	// All settings are available. This typically means
	// scorecard is run with a token that has access
	// to admin settings.
	case data.BranchProtectionRule != nil:
		rule := data.BranchProtectionRule

		// Admin settings.
		copyAdminSettings(rule, branchRule)

		// Non-admin settings.
		copyNonAdminSettings(rule, branchRule)

	// Only non-admin settings are available.
	// https://docs.github.com/en/graphql/reference/objects#refupdaterule.
	case data.RefUpdateRule != nil:
		rule := data.RefUpdateRule
		copyNonAdminSettings(rule, branchRule)
	}

	applyRepoRules(branchRef, rules)

	return branchRef
}

func isPermissionsError(err error) bool {
	return strings.Contains(err.Error(), "Resource not accessible")
}

const (
	ruleConditionDefaultBranch = "~DEFAULT_BRANCH"
	ruleConditionAllBranches   = "~ALL"
	ruleDeletion               = "DELETION"
	ruleForcePush              = "NON_FAST_FORWARD"
	ruleLinear                 = "REQUIRED_LINEAR_HISTORY"
	rulePullRequest            = "PULL_REQUEST"
	ruleStatusCheck            = "REQUIRED_STATUS_CHECKS"
)

func rulesMatchingBranch(rules []*repoRuleSet, name string, defaultRef bool) ([]*repoRuleSet, error) {
	refName := refPrefix + name
	ret := make([]*repoRuleSet, 0)
nextRule:
	for _, rule := range rules {
		for _, cond := range rule.Conditions.RefName.Exclude {
			if match, err := fnmatch.Match(cond, refName); err != nil {
				return nil, fmt.Errorf("exclude match error: %w", err)
			} else if match {
				continue nextRule
			}
		}

		for _, cond := range rule.Conditions.RefName.Include {
			if cond == ruleConditionAllBranches {
				ret = append(ret, rule)
				break
			}
			if cond == ruleConditionDefaultBranch && defaultRef {
				ret = append(ret, rule)
				break
			}

			if match, err := fnmatch.Match(cond, refName); err != nil {
				return nil, fmt.Errorf("include match error: %w", err)
			} else if match {
				ret = append(ret, rule)
			}
		}
	}
	return ret, nil
}

func applyRepoRules(branchRef *clients.BranchRef, rules []*repoRuleSet) {
	falseVal := false
	trueVal := true
	for _, r := range rules {
		adminEnforced := len(r.BypassActors.Nodes) == 0
		translated := clients.BranchProtectionRule{
			EnforceAdmins: &adminEnforced,
		}

		for _, rule := range r.Rules.Nodes {
			switch rule.Type {
			case ruleDeletion:
				translated.AllowDeletions = &falseVal
			case ruleForcePush:
				translated.AllowForcePushes = &falseVal
			case ruleLinear:
				translated.RequireLinearHistory = &trueVal
			case rulePullRequest:
				translatePullRequestRepoRule(&translated, rule)
			case ruleStatusCheck:
				translateRequiredStatusRepoRule(&translated, rule)
			}
		}
		mergeBranchProtectionRules(&branchRef.BranchProtectionRule, &translated)
	}
}

func translatePullRequestRepoRule(base *clients.BranchProtectionRule, rule *repoRule) {
	if readBoolPtr(rule.Parameters.PullRequestParameters.DismissStaleReviewsOnPush) {
		base.RequiredPullRequestReviews.DismissStaleReviews = rule.Parameters.PullRequestParameters.DismissStaleReviewsOnPush
	}
	if readBoolPtr(rule.Parameters.PullRequestParameters.RequireCodeOwnerReview) {
		base.RequiredPullRequestReviews.RequireCodeOwnerReviews = rule.Parameters.PullRequestParameters.RequireCodeOwnerReview
	}
	if readBoolPtr(rule.Parameters.PullRequestParameters.RequireLastPushApproval) {
		base.RequireLastPushApproval = rule.Parameters.PullRequestParameters.RequireLastPushApproval
	}
	if reviewerCount := readIntPtr(rule.Parameters.PullRequestParameters.RequiredApprovingReviewCount); reviewerCount > 0 {
		base.RequiredPullRequestReviews.RequiredApprovingReviewCount = &reviewerCount
	}
}

func translateRequiredStatusRepoRule(base *clients.BranchProtectionRule, rule *repoRule) {
	statusParams := rule.Parameters.StatusCheckParameters
	if len(statusParams.RequiredStatusChecks) == 0 {
		return
	}
	enabled := true
	base.CheckRules.RequiresStatusChecks = &enabled
	base.CheckRules.UpToDateBeforeMerge = statusParams.StrictRequiredStatusChecksPolicy
	for _, chk := range statusParams.RequiredStatusChecks {
		if chk.Context == nil {
			continue
		}
		base.CheckRules.Contexts = append(base.CheckRules.Contexts, *chk.Context)
	}
}

func mergeBranchProtectionRules(base, translated *clients.BranchProtectionRule) {
	if base.AllowDeletions == nil || translated.AllowDeletions != nil && !*translated.AllowDeletions {
		base.AllowDeletions = translated.AllowDeletions
	}
	if base.AllowForcePushes == nil || translated.AllowForcePushes != nil && !*translated.AllowForcePushes {
		base.AllowForcePushes = translated.AllowForcePushes
	}
	if base.EnforceAdmins == nil || translated.EnforceAdmins != nil && !*translated.EnforceAdmins {
		// this is an over simplification to get preliminary support for repo rules merged.
		// A more complete approach would process all rules without bypass actors first,
		// then process those with bypass actors. If no settings improve (due to rule layering),
		// then we can ignore the bypass actors.
		// https://github.com/ossf/scorecard/issues/3480
		base.EnforceAdmins = translated.EnforceAdmins
	}
	if base.RequireLastPushApproval == nil || readBoolPtr(translated.RequireLastPushApproval) {
		base.RequireLastPushApproval = translated.RequireLastPushApproval
	}
	if base.RequireLinearHistory == nil || readBoolPtr(translated.RequireLinearHistory) {
		base.RequireLinearHistory = translated.RequireLinearHistory
	}
	mergePullRequestReviews(&base.RequiredPullRequestReviews, &translated.RequiredPullRequestReviews)
	mergeCheckRules(&base.CheckRules, &translated.CheckRules)
}

func mergeCheckRules(base, translated *clients.StatusChecksRule) {
	if base.UpToDateBeforeMerge == nil || readBoolPtr(translated.UpToDateBeforeMerge) {
		base.UpToDateBeforeMerge = translated.UpToDateBeforeMerge
	}
	if base.RequiresStatusChecks == nil || readBoolPtr(translated.RequiresStatusChecks) {
		base.RequiresStatusChecks = translated.RequiresStatusChecks
	}
	for _, context := range translated.Contexts {
		// this isn't optimal, but probably not a bottleneck.
		if !slices.Contains(base.Contexts, context) {
			base.Contexts = append(base.Contexts, context)
		}
	}
}

func mergePullRequestReviews(base, translated *clients.PullRequestReviewRule) {
	if readIntPtr(translated.RequiredApprovingReviewCount) > readIntPtr(base.RequiredApprovingReviewCount) {
		base.RequiredApprovingReviewCount = translated.RequiredApprovingReviewCount
	}
	if base.DismissStaleReviews == nil || readBoolPtr(translated.DismissStaleReviews) {
		base.DismissStaleReviews = translated.DismissStaleReviews
	}
	if base.RequireCodeOwnerReviews == nil || readBoolPtr(translated.RequireCodeOwnerReviews) {
		base.RequireCodeOwnerReviews = translated.RequireCodeOwnerReviews
	}
}

func readBoolPtr(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func readIntPtr(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}
