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

package format

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/pkg"
)

// Flat JSON structure to hold raw results.
type jsonScorecardRawResult struct {
	Date      string          `json:"date"`
	Repo      jsonRepoV2      `json:"repo"`
	Scorecard jsonScorecardV2 `json:"scorecard"`
	Metadata  []string        `json:"metadata"`
	Results   jsonRawResults  `json:"results"`
}

// TODO: separate each check extraction into its own file.
type jsonFile struct {
	Path   string `json:"path"`
	Offset int    `json:"offset,omitempty"`
}

type jsonTool struct {
	URL   *string    `json:"url"`
	Desc  *string    `json:"desc"`
	Name  string     `json:"name"`
	Files []jsonFile `json:"file"`
	// TODO: Runs, Issues, Merge requests.
}

type jsonBranchProtectionSettings struct {
	RequiredApprovingReviewCount        *int32   `json:"requiredReviewerCount"`
	AllowsDeletions                     *bool    `json:"allowsDeletions"`
	AllowsForcePushes                   *bool    `json:"allowsForcePushes"`
	RequiresCodeOwnerReviews            *bool    `json:"requiresCodeOwnerReview"`
	RequiresLinearHistory               *bool    `json:"requiredLinearHistory"`
	DismissesStaleReviews               *bool    `json:"dismissesStaleReviews"`
	EnforcesAdmins                      *bool    `json:"enforcesAdmin"`
	RequiresStatusChecks                *bool    `json:"requiresStatusChecks"`
	RequiresUpToDateBranchBeforeMerging *bool    `json:"requiresUpdatedBranchesToMerge"`
	StatusCheckContexts                 []string `json:"statusChecksContexts"`
}

type jsonBranchProtection struct {
	Protection *jsonBranchProtectionSettings `json:"protection"`
	Name       string                        `json:"name"`
}

type jsonReview struct {
	Reviewer jsonUser `json:"reviewer"`
	State    string   `json:"state"`
}

type jsonUser struct {
	Login string `json:"login"`
}

type jsonDefaultBranchChangeset struct {
	// ApprovedReviews *jsonApprovedReviews `json:"approved-reviews"`
	RevisionID     string       `json:"number"`
	ReviewPlatform string       `json:"platform"`
	Reviews        []jsonReview `json:"reviews"`
	Authors        []jsonUser   `json:"authors"`
	Commits        []jsonCommit `json:"commits"`
}

type jsonCommit struct {
	Committer jsonUser `json:"committer"`
	Message   string   `json:"message"`
	SHA       string   `json:"sha"`

	// TODO: check runs, etc.
}

type jsonDatabaseVulnerability struct {
	// For OSV: OSV-2020-484
	// For CVE: CVE-2022-23945
	ID string `json:"id"`
	// TODO: additional information
}

type jsonRawResults struct {
	DatabaseVulnerabilities []jsonDatabaseVulnerability `json:"databaseVulnerabilities"`
	// List of binaries found in the repo.
	Binaries []jsonFile `json:"binaries"`
	// List of security policy files found in the repo.
	// Note: we return one at most.
	SecurityPolicies []jsonFile `json:"securityPolicies"`
	// List of update tools.
	// Note: we return one at most.
	DependencyUpdateTools []jsonTool `json:"dependencyUpdateTools"`
	// Branch protection settings for development and release branches.
	BranchProtections []jsonBranchProtection `json:"branchProtections"`
	// Changesets
	DefaultBranchChangesets []jsonDefaultBranchChangeset `json:"defaultBranchChangesets"`
}

//nolint:unparam
func addCodeReviewRawResults(r *jsonScorecardRawResult, cr *checker.CodeReviewData) error {
	r.Results.DefaultBranchChangesets = []jsonDefaultBranchChangeset{}

	for i := range cr.DefaultBranchChangesets {
		cs := cr.DefaultBranchChangesets[i]

		// commits field
		commits := []jsonCommit{}
		for i := range cs.Commits {
			commits = append(commits, jsonCommit{
				Committer: jsonUser{
					Login: commits[i].Committer.Login,
				},
				Message: commits[i].Message,
				SHA:     commits[i].SHA,
			})
		}

		// reviews field
		reviews := []jsonReview{}
		for _, r := range cs.Reviews {
			reviews = append(reviews, jsonReview{
				State: r.State,
				Reviewer: jsonUser{
					Login: r.Author.Login,
				},
			})
		}

		// authors field
		authors := []jsonUser{}
		for _, a := range cs.Authors {
			authors = append(authors, jsonUser{
				Login: a.Login,
			})
		}

		r.Results.DefaultBranchChangesets = append(r.Results.DefaultBranchChangesets,
			jsonDefaultBranchChangeset{
				RevisionID: cs.RevisionID,
				Commits:    commits,
				Reviews:    reviews,
				Authors:    authors,
			},
		)
	}
	return nil
}

//nolint:unparam
func addVulnerbilitiesRawResults(r *jsonScorecardRawResult, vd *checker.VulnerabilitiesData) error {
	r.Results.DatabaseVulnerabilities = []jsonDatabaseVulnerability{}
	for _, v := range vd.Vulnerabilities {
		r.Results.DatabaseVulnerabilities = append(r.Results.DatabaseVulnerabilities,
			jsonDatabaseVulnerability{
				ID: v.ID,
			})
	}
	return nil
}

//nolint:unparam
func addBinaryArtifactRawResults(r *jsonScorecardRawResult, ba *checker.BinaryArtifactData) error {
	r.Results.Binaries = []jsonFile{}
	for _, v := range ba.Files {
		r.Results.Binaries = append(r.Results.Binaries, jsonFile{
			Path: v.Path,
		})
	}
	return nil
}

//nolint:unparam
func addSecurityPolicyRawResults(r *jsonScorecardRawResult, sp *checker.SecurityPolicyData) error {
	r.Results.SecurityPolicies = []jsonFile{}
	for _, v := range sp.Files {
		r.Results.SecurityPolicies = append(r.Results.SecurityPolicies, jsonFile{
			Path: v.Path,
		})
	}
	return nil
}

//nolint:unparam
func addDependencyUpdateToolRawResults(r *jsonScorecardRawResult,
	dut *checker.DependencyUpdateToolData,
) error {
	r.Results.DependencyUpdateTools = []jsonTool{}
	for i := range dut.Tools {
		t := dut.Tools[i]
		jt := jsonTool{
			Name: t.Name,
			URL:  t.URL,
			Desc: t.Desc,
		}
		if t.Files != nil && len(t.Files) > 0 {
			for _, f := range t.Files {
				jt.Files = append(jt.Files, jsonFile{
					Path: f.Path,
				})
			}
		}
		r.Results.DependencyUpdateTools = append(r.Results.DependencyUpdateTools, jt)
	}
	return nil
}

//nolint:unparam
func addBranchProtectionRawResults(r *jsonScorecardRawResult, bp *checker.BranchProtectionsData) error {
	r.Results.BranchProtections = []jsonBranchProtection{}
	for _, v := range bp.Branches {
		var bp *jsonBranchProtectionSettings
		if v.Protected != nil && *v.Protected {
			bp = &jsonBranchProtectionSettings{
				AllowsDeletions:                     v.BranchProtectionRule.AllowDeletions,
				AllowsForcePushes:                   v.BranchProtectionRule.AllowForcePushes,
				RequiresCodeOwnerReviews:            v.BranchProtectionRule.RequiredPullRequestReviews.RequireCodeOwnerReviews,
				RequiresLinearHistory:               v.BranchProtectionRule.RequireLinearHistory,
				DismissesStaleReviews:               v.BranchProtectionRule.RequiredPullRequestReviews.DismissStaleReviews,
				EnforcesAdmins:                      v.BranchProtectionRule.EnforceAdmins,
				RequiresStatusChecks:                v.BranchProtectionRule.CheckRules.RequiresStatusChecks,
				RequiresUpToDateBranchBeforeMerging: v.BranchProtectionRule.CheckRules.UpToDateBeforeMerge,
				RequiredApprovingReviewCount:        v.BranchProtectionRule.RequiredPullRequestReviews.RequiredApprovingReviewCount,
				StatusCheckContexts:                 v.BranchProtectionRule.CheckRules.Contexts,
			}
		}
		r.Results.BranchProtections = append(r.Results.BranchProtections, jsonBranchProtection{
			Name:       *v.Name,
			Protection: bp,
		})
	}
	return nil
}

func fillJSONRawResults(r *jsonScorecardRawResult, raw *checker.RawResults) error {
	// Vulnerabiliries.
	if err := addVulnerbilitiesRawResults(r, &raw.VulnerabilitiesResults); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	}

	// Binary-Artifacts.
	if err := addBinaryArtifactRawResults(r, &raw.BinaryArtifactResults); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	}

	// Security-Policy.
	if err := addSecurityPolicyRawResults(r, &raw.SecurityPolicyResults); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	}

	// Dependency-Update-Tool.
	if err := addDependencyUpdateToolRawResults(r, &raw.DependencyUpdateToolResults); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	}

	// Branch-Protection.
	if err := addBranchProtectionRawResults(r, &raw.BranchProtectionResults); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	}

	// Code-Review.
	if err := addCodeReviewRawResults(r, &raw.CodeReviewResults); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	}

	return nil
}

// AsRawJSON exports results as JSON for raw results.
func AsRawJSON(r *pkg.ScorecardResult, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	out := jsonScorecardRawResult{
		Repo: jsonRepoV2{
			Name:   r.Repo.Name,
			Commit: r.Repo.CommitSHA,
		},
		Scorecard: jsonScorecardV2{
			Version: r.Scorecard.Version,
			Commit:  r.Scorecard.CommitSHA,
		},
		Date:     r.Date.Format("2006-01-02"),
		Metadata: r.Metadata,
	}

	// if err := out.fillJSONRawResults(r.Checks[0].RawResults); err != nil {
	if err := fillJSONRawResults(&out, &r.RawResults); err != nil {
		return err
	}

	if err := encoder.Encode(out); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("encoder.Encode: %v", err))
	}

	return nil
}
