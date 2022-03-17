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
	Name        string     `json:"name"`
	URL         string     `json:"url"`
	Desc        string     `json:"desc"`
	ConfigFiles []jsonFile `json:"files"`
	// TODO: Runs, Issues, Merge requests.
}

type jsonBranchProtectionSettings struct {
	RequiredApprovingReviewCount        *int     `json:"required-reviewer-count"`
	AllowsDeletions                     *bool    `json:"allows-deletions"`
	AllowsForcePushes                   *bool    `json:"allows-force-pushes"`
	RequiresCodeOwnerReviews            *bool    `json:"requires-code-owner-review"`
	RequiresLinearHistory               *bool    `json:"required-linear-history"`
	DismissesStaleReviews               *bool    `json:"dismisses-stale-reviews"`
	EnforcesAdmins                      *bool    `json:"enforces-admin"`
	RequiresStatusChecks                *bool    `json:"requires-status-checks"`
	RequiresUpToDateBranchBeforeMerging *bool    `json:"requires-updated-branches-to-merge"`
	StatusCheckContexts                 []string `json:"status-checks-contexts"`
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

//nolint:govet
type jsonMergeRequest struct {
	Number  int          `json:"number"`
	Labels  []string     `json:"labels"`
	Reviews []jsonReview `json:"reviews"`
	Author  jsonUser     `json:"author"`
}

type jsonDefaultBranchCommit struct {
	// ApprovedReviews *jsonApprovedReviews `json:"approved-reviews"`
	Committer     jsonUser          `json:"committer"`
	MergeRequest  *jsonMergeRequest `json:"merge-request"`
	CommitMessage string            `json:"commit-message"`
	SHA           string            `json:"sha"`

	// TODO: check runs, etc.
}

type jsonRawResults struct {
	DatabaseVulnerabilities []jsonDatabaseVulnerability `json:"database-vulnerabilities"`
	// List of binaries found in the repo.
	Binaries []jsonFile `json:"binaries"`
	// List of security policy files found in the repo.
	// Note: we return one at most.
	SecurityPolicies []jsonFile `json:"security-policies"`
	// List of update tools.
	// Note: we return one at most.
	DependencyUpdateTools []jsonTool `json:"dependency-update-tools"`
	// Branch protection settings for development and release branches.
	BranchProtections []jsonBranchProtection `json:"branch-protections"`
	// Commits.
	DefaultBranchCommits []jsonDefaultBranchCommit `json:"default-branch-commits"`
}

//nolint:unparam
func addCodeReviewRawResults(r *jsonScorecardRawResult, cr *checker.CodeReviewData) error {
	r.Results.DefaultBranchCommits = []jsonDefaultBranchCommit{}
	for _, commit := range cr.DefaultBranchCommits {
		com := jsonDefaultBranchCommit{
			Committer: jsonUser{
				Login: commit.Committer.Login,
			},
			CommitMessage: commit.CommitMessage,
			SHA:           commit.SHA,
		}

		// Merge request field.
		if commit.MergeRequest != nil {
			mr := jsonMergeRequest{
				Number: commit.MergeRequest.Number,
				Author: jsonUser{
					Login: commit.MergeRequest.Author.Login,
				},
			}

			if len(commit.MergeRequest.Labels) > 0 {
				mr.Labels = commit.MergeRequest.Labels
			}

			for _, r := range commit.MergeRequest.Reviews {
				mr.Reviews = append(mr.Reviews, jsonReview{
					State: r.State,
					Reviewer: jsonUser{
						Login: r.Reviewer.Login,
					},
				})
			}

			com.MergeRequest = &mr
		}

		com.CommitMessage = commit.CommitMessage

		r.Results.DefaultBranchCommits = append(r.Results.DefaultBranchCommits, com)
	}
	return nil
}

type jsonDatabaseVulnerability struct {
	// For OSV: OSV-2020-484
	// For CVE: CVE-2022-23945
	ID string
	// TODO: additional information
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
		offset := len(r.Results.DependencyUpdateTools)
		r.Results.DependencyUpdateTools = append(r.Results.DependencyUpdateTools, jsonTool{
			Name: t.Name,
			URL:  t.URL,
			Desc: t.Desc,
		})
		for _, f := range t.ConfigFiles {
			r.Results.DependencyUpdateTools[offset].ConfigFiles = append(
				r.Results.DependencyUpdateTools[offset].ConfigFiles,
				jsonFile{
					Path: f.Path,
				},
			)
		}
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
				AllowsDeletions:                     v.AllowsDeletions,
				AllowsForcePushes:                   v.AllowsForcePushes,
				RequiresCodeOwnerReviews:            v.RequiresCodeOwnerReviews,
				RequiresLinearHistory:               v.RequiresLinearHistory,
				DismissesStaleReviews:               v.DismissesStaleReviews,
				EnforcesAdmins:                      v.EnforcesAdmins,
				RequiresStatusChecks:                v.RequiresStatusChecks,
				RequiresUpToDateBranchBeforeMerging: v.RequiresUpToDateBranchBeforeMerging,
				RequiredApprovingReviewCount:        v.RequiredApprovingReviewCount,
				StatusCheckContexts:                 v.StatusCheckContexts,
			}
		}
		r.Results.BranchProtections = append(r.Results.BranchProtections, jsonBranchProtection{
			Name:       v.Name,
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
