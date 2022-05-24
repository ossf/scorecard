// Copyright 2021 Security Scorecard Authors
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

package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
)

// TODO: add a "check" field to all results so that they can be linked to a check.
// TODO(#1874): Add a severity field in all results.

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
	Snippet *string `json:"snippet,omitempty"`
	Path    string  `json:"path"`
	Offset  int     `json:"offset,omitempty"`
}

type jsonTool struct {
	URL  *string          `json:"url"`
	Desc *string          `json:"desc"`
	Job  *jsonWorkflowJob `json:"job,omitempty"`
	File *jsonFile        `json:"file,omitempty"`
	Name string           `json:"name"`
	// TODO: Runs, Issues, Merge requests.
}

type jsonBranchProtectionSettings struct {
	RequiredApprovingReviewCount        *int32   `json:"required-reviewer-count"`
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
	State    string   `json:"state"`
	Reviewer jsonUser `json:"reviewer"`
}

type jsonUser struct {
	RepoAssociation *string `json:"repo-association,omitempty"`
	Login           string  `json:"login"`
	// Orgnization refers to a GitHub org.
	Organizations []jsonOrganization `json:"organization,omitempty"`
	// Companies refer to a claim by a user in their profile.
	Companies        []jsonCompany `json:"company,omitempty"`
	NumContributions uint          `json:"NumContributions"`
}

type jsonContributors struct {
	Users []jsonUser `json:"users"`
	// TODO: high-level statistics, etc
}

type jsonOrganization struct {
	Login string `json:"login"`
	// TODO: other info.
}

type jsonCompany struct {
	Name string `json:"name"`
	// TODO: other info.
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
	MergeRequest  *jsonMergeRequest `json:"merge-request"`
	CommitMessage string            `json:"commit-message"`
	SHA           string            `json:"sha"`
	Committer     jsonUser          `json:"committer"`
	// TODO: check runs, etc.
}

type jsonDatabaseVulnerability struct {
	// For OSV: OSV-2020-484
	// For CVE: CVE-2022-23945
	ID string `json:"id"`
	// TODO: additional information
}

type jsonArchivedStatus struct {
	Status bool `json:"status"`
	// TODO: add fields, e.g. date of archival, etc.
}

type jsonComment struct {
	CreatedAt *time.Time `json:"created-at"`
	Author    *jsonUser  `json:"author"`
	// TODO: add ields if needed, e.g., content.
}

type jsonIssue struct {
	CreatedAt *time.Time    `json:"created-at"`
	Author    *jsonUser     `json:"author"`
	URL       string        `json:"URL"`
	Comments  []jsonComment `json:"comments"`
	// TODO: add fields, e.g., state=[opened|closed]
}

type jsonRelease struct {
	Tag    string             `json:"tag"`
	URL    string             `json:"url"`
	Assets []jsonReleaseAsset `json:"assets"`
	// TODO: add needed fields, e.g. Path.
}

type jsonReleaseAsset struct {
	Path string `json:"path"`
	URL  string `json:"url"`
}

type jsonOssfBestPractices struct {
	Badge string `json:"badge"`
}

//nolint
type jsonLicense struct {
	File jsonFile `json:"file"`
	// TODO: add fields, like type of license, etc.
}

type jsonWorkflow struct {
	Job  *jsonWorkflowJob `json:"job"`
	File *jsonFile        `json:"file"`
	// Type is a string to allow different types for permissions, unpinned dependencies, etc.
	Type string `json:"type"`
}

type jsonWorkflowJob struct {
	Name *string `json:"name"`
	ID   *string `json:"id"`
}

//nolint
type jsonRawResults struct {
	// Workflow results.
	Workflows []jsonWorkflow `json:"workflows"`
	// License.
	Licenses []jsonLicense `json:"licenses"`
	// List of recent issues.
	RecentIssues []jsonIssue `json:"issues"`
	// OSSF best practices badge.
	OssfBestPractices jsonOssfBestPractices `json:"openssf-best-practices-badge"`
	// Vulnerabilities.
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
	// Contributors. Note: we could use the list of commits instead to store this data.
	// However, it's harder to get statistics using commit list, so we have a dedicated
	// structure for it.
	Contributors jsonContributors `json:"Contributors"`
	// Commits.
	DefaultBranchCommits []jsonDefaultBranchCommit `json:"default-branch-commits"`
	// Archived status of the repo.
	ArchivedStatus jsonArchivedStatus `json:"archived"`
	// Fuzzers.
	Fuzzers []jsonTool `json:"fuzzers"`
	// Releases.
	Releases []jsonRelease `json:"releases"`
}

//nolint:unparam
func (r *jsonScorecardRawResult) addDangerousWorkflowRawResults(df *checker.DangerousWorkflowData) error {
	r.Results.Workflows = []jsonWorkflow{}
	for _, e := range df.Workflows {
		v := jsonWorkflow{
			File: &jsonFile{
				Path:   e.File.Path,
				Offset: int(e.File.Offset),
			},
			Type: string(e.Type),
		}
		if e.File.Snippet != "" {
			v.File.Snippet = &e.File.Snippet
		}
		if e.Job != nil {
			v.Job = &jsonWorkflowJob{
				Name: e.Job.Name,
				ID:   e.Job.ID,
			}
		}

		r.Results.Workflows = append(r.Results.Workflows, v)
	}

	return nil
}

//nolint:unparam
func (r *jsonScorecardRawResult) addContributorsRawResults(cr *checker.ContributorsData) error {
	r.Results.Contributors = jsonContributors{}

	for _, user := range cr.Users {
		u := jsonUser{
			Login:            user.Login,
			NumContributions: user.NumContributions,
		}

		for _, org := range user.Organizations {
			u.Organizations = append(u.Organizations,
				jsonOrganization{
					Login: org.Login,
				},
			)
		}

		for _, comp := range user.Companies {
			u.Companies = append(u.Companies,
				jsonCompany{
					Name: comp.Name,
				},
			)
		}

		r.Results.Contributors.Users = append(r.Results.Contributors.Users, u)
	}

	return nil
}

//nolint:unparam
func (r *jsonScorecardRawResult) addSignedReleasesRawResults(sr *checker.SignedReleasesData) error {
	r.Results.Releases = []jsonRelease{}
	for i, release := range sr.Releases {
		r.Results.Releases = append(r.Results.Releases,
			jsonRelease{
				Tag: release.TagName,
				URL: release.URL,
			})
		for _, asset := range release.Assets {
			r.Results.Releases[i].Assets = append(r.Results.Releases[i].Assets,
				jsonReleaseAsset{
					Path: asset.Name,
					URL:  asset.URL,
				},
			)
		}
	}
	return nil
}

func (r *jsonScorecardRawResult) addMaintainedRawResults(mr *checker.MaintainedData) error {
	// Set archived status.
	r.Results.ArchivedStatus = jsonArchivedStatus{Status: mr.ArchivedStatus.Status}

	// Issues.
	for i := range mr.Issues {
		issue := jsonIssue{
			CreatedAt: mr.Issues[i].CreatedAt,
			URL:       *mr.Issues[i].URI,
		}

		if mr.Issues[i].Author != nil {
			issue.Author = &jsonUser{
				Login:           mr.Issues[i].Author.Login,
				RepoAssociation: getStrPtr(mr.Issues[i].AuthorAssociation.String()),
			}
		}

		for j := range mr.Issues[i].Comments {
			comment := jsonComment{
				CreatedAt: mr.Issues[i].Comments[j].CreatedAt,
			}

			if mr.Issues[i].Comments[j].Author != nil {
				comment.Author = &jsonUser{
					Login:           mr.Issues[i].Comments[j].Author.Login,
					RepoAssociation: getStrPtr(mr.Issues[i].Comments[j].AuthorAssociation.String()),
				}
			}

			issue.Comments = append(issue.Comments, comment)
		}

		r.Results.RecentIssues = append(r.Results.RecentIssues, issue)
	}

	return r.setDefaultCommitData(mr.DefaultBranchCommits)
}

func getStrPtr(s string) *string {
	ret := s
	return &ret
}

// Function shared between addMaintainedRawResults() and addCodeReviewRawResults().
func (r *jsonScorecardRawResult) setDefaultCommitData(commits []clients.Commit) error {
	if len(r.Results.DefaultBranchCommits) > 0 {
		return nil
	}

	r.Results.DefaultBranchCommits = []jsonDefaultBranchCommit{}
	for i := range commits {
		commit := commits[i]
		com := jsonDefaultBranchCommit{
			Committer: jsonUser{
				Login: commit.Committer.Login,
				// Note: repo association is not available. We could
				// try to use issue information to set it, but we're likely to miss
				// many anyway.
			},
			CommitMessage: commit.Message,
			SHA:           commit.SHA,
		}

		// Merge request field.
		mr := jsonMergeRequest{
			Number: commit.AssociatedMergeRequest.Number,
			Author: jsonUser{
				Login: commit.AssociatedMergeRequest.Author.Login,
			},
		}

		for _, l := range commit.AssociatedMergeRequest.Labels {
			mr.Labels = append(mr.Labels, l.Name)
		}

		for _, r := range commit.AssociatedMergeRequest.Reviews {
			mr.Reviews = append(mr.Reviews, jsonReview{
				State: r.State,
				Reviewer: jsonUser{
					Login: r.Author.Login,
				},
			})
		}

		com.MergeRequest = &mr

		com.CommitMessage = commit.Message

		r.Results.DefaultBranchCommits = append(r.Results.DefaultBranchCommits, com)
	}
	return nil
}

//nolint:unparam
func (r *jsonScorecardRawResult) addOssfBestPracticesRawResults(cbp *checker.CIIBestPracticesData) error {
	r.Results.OssfBestPractices.Badge = cbp.Badge.String()
	return nil
}

func (r *jsonScorecardRawResult) addCodeReviewRawResults(cr *checker.CodeReviewData) error {
	return r.setDefaultCommitData(cr.DefaultBranchCommits)
}

//nolint:unparam
func (r *jsonScorecardRawResult) addLicenseRawResults(ld *checker.LicenseData) error {
	r.Results.Licenses = []jsonLicense{}
	for _, file := range ld.Files {
		r.Results.Licenses = append(r.Results.Licenses,
			jsonLicense{
				File: jsonFile{
					Path: file.Path,
				},
			},
		)
	}
	return nil
}

//nolint:unparam
func (r *jsonScorecardRawResult) addVulnerbilitiesRawResults(vd *checker.VulnerabilitiesData) error {
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
func (r *jsonScorecardRawResult) addBinaryArtifactRawResults(ba *checker.BinaryArtifactData) error {
	r.Results.Binaries = []jsonFile{}
	for _, v := range ba.Files {
		r.Results.Binaries = append(r.Results.Binaries, jsonFile{
			Path: v.Path,
		})
	}
	return nil
}

//nolint:unparam
func (r *jsonScorecardRawResult) addSecurityPolicyRawResults(sp *checker.SecurityPolicyData) error {
	r.Results.SecurityPolicies = []jsonFile{}
	for _, v := range sp.Files {
		r.Results.SecurityPolicies = append(r.Results.SecurityPolicies, jsonFile{
			Path: v.Path,
		})
	}
	return nil
}

//nolint:unparam
func (r *jsonScorecardRawResult) addFuzzingRawResults(fd *checker.FuzzingData) error {
	r.Results.Fuzzers = []jsonTool{}
	for i := range fd.Fuzzers {
		f := fd.Fuzzers[i]
		jt := jsonTool{
			Name: f.Name,
			URL:  f.URL,
			Desc: f.Desc,
		}
		if f.File != nil {
			jt.File = &jsonFile{
				Path: f.File.Path,
			}
		}
		r.Results.Fuzzers = append(r.Results.Fuzzers, jt)
	}
	return nil
}

//nolint:unparam
func (r *jsonScorecardRawResult) addDependencyUpdateToolRawResults(dut *checker.DependencyUpdateToolData) error {
	r.Results.DependencyUpdateTools = []jsonTool{}
	for i := range dut.Tools {
		t := dut.Tools[i]
		jt := jsonTool{
			Name: t.Name,
			URL:  t.URL,
			Desc: t.Desc,
		}
		if t.File != nil {
			jt.File = &jsonFile{
				Path: t.File.Path,
			}
		}
		r.Results.DependencyUpdateTools = append(r.Results.DependencyUpdateTools, jt)
	}
	return nil
}

//nolint:unparam
func (r *jsonScorecardRawResult) addBranchProtectionRawResults(bp *checker.BranchProtectionsData) error {
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

func (r *jsonScorecardRawResult) fillJSONRawResults(raw *checker.RawResults) error {
	// Licenses.
	if err := r.addLicenseRawResults(&raw.LicenseResults); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	}

	// Vulnerabilities.
	if err := r.addVulnerbilitiesRawResults(&raw.VulnerabilitiesResults); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	}

	// Binary-Artifacts.
	if err := r.addBinaryArtifactRawResults(&raw.BinaryArtifactResults); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	}

	// Security-Policy.
	if err := r.addSecurityPolicyRawResults(&raw.SecurityPolicyResults); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	}

	// Dependency-Update-Tool.
	if err := r.addDependencyUpdateToolRawResults(&raw.DependencyUpdateToolResults); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	}

	// Branch-Protection.
	if err := r.addBranchProtectionRawResults(&raw.BranchProtectionResults); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	}

	// Code-Review.
	if err := r.addCodeReviewRawResults(&raw.CodeReviewResults); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	}

	// Maintained.
	if err := r.addMaintainedRawResults(&raw.MaintainedResults); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	}

	// Signed-Releases.
	if err := r.addSignedReleasesRawResults(&raw.SignedReleasesResults); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	}

	// Contributors.
	if err := r.addContributorsRawResults(&raw.ContributorsResults); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	}

	// CII-Best-Practices.
	if err := r.addOssfBestPracticesRawResults(&raw.CIIBestPracticesResults); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	}

	// Dangerous workflow.
	if err := r.addDangerousWorkflowRawResults(&raw.DangerousWorkflowResults); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	}

	// Fuzzers.
	if err := r.addFuzzingRawResults(&raw.FuzzingResults); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	}

	return nil
}

// AsRawJSON exports results as JSON for raw results.
func (r *ScorecardResult) AsRawJSON(writer io.Writer) error {
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
	if err := out.fillJSONRawResults(&r.RawResults); err != nil {
		return err
	}

	if err := encoder.Encode(out); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("encoder.Encode: %v", err))
	}

	return nil
}
