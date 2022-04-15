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

	"github.com/mcuadros/go-jsonschema-generator"
	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
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
	RepoAssociation *string `json:"repo-association"`
	Login           string  `json:"login"`
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

//nolint
type jsonLicense struct {
	File jsonFile `json:"file"`
	// TODO: add fields, like type of license, etc.
}

//nolint
type jsonRawResults struct {
	// License.
	Licenses []jsonLicense `json:"licenses"`
	// List of recent issues.
	RecentIssues []jsonIssue `json:"issues"`
	// List of vulnerabilities.
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
	// Archived status of the repo.
	ArchivedStatus jsonArchivedStatus `json:"archived"`
	// Releases.
	Releases []jsonRelease `json:"releases"`
}

//nolint:unparam
func (r *jsonScorecardRawResult) addSignedReleasesRawResults(sr *checker.SignedReleasesData) error {
	r.Results.Releases = []jsonRelease{}
	for i, release := range sr.Releases {
		r.Results.Releases = append(r.Results.Releases,
			jsonRelease{
				Tag: release.Tag,
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

func getRepoAssociation(author *checker.User) *string {
	if author == nil || author.RepoAssociation == nil {
		return nil
	}

	s := string(*author.RepoAssociation)
	return &s
}

func (r *jsonScorecardRawResult) addMaintainedRawResults(mr *checker.MaintainedData) error {
	// Set archived status.
	r.Results.ArchivedStatus = jsonArchivedStatus{Status: mr.ArchivedStatus.Status}

	// Issues.
	for i := range mr.Issues {
		issue := jsonIssue{
			CreatedAt: mr.Issues[i].CreatedAt,
			URL:       mr.Issues[i].URL,
		}

		if mr.Issues[i].Author != nil {
			issue.Author = &jsonUser{
				Login:           mr.Issues[i].Author.Login,
				RepoAssociation: getRepoAssociation(mr.Issues[i].Author),
			}
		}

		for j := range mr.Issues[i].Comments {
			comment := jsonComment{
				CreatedAt: mr.Issues[i].Comments[j].CreatedAt,
			}

			if mr.Issues[i].Comments[j].Author != nil {
				comment.Author = &jsonUser{
					Login:           mr.Issues[i].Comments[j].Author.Login,
					RepoAssociation: getRepoAssociation(mr.Issues[i].Comments[j].Author),
				}
			}

			issue.Comments = append(issue.Comments, comment)
		}

		r.Results.RecentIssues = append(r.Results.RecentIssues, issue)
	}

	return r.setDefaultCommitData(mr.DefaultBranchCommits)
}

// Function shared between addMaintainedRawResults() and addCodeReviewRawResults().
func (r *jsonScorecardRawResult) setDefaultCommitData(commits []checker.DefaultBranchCommit) error {
	if len(r.Results.DefaultBranchCommits) > 0 {
		return nil
	}

	r.Results.DefaultBranchCommits = []jsonDefaultBranchCommit{}
	for _, commit := range commits {
		com := jsonDefaultBranchCommit{
			Committer: jsonUser{
				Login: commit.Committer.Login,
				// Note: repo association is not available. We could
				// try to use issue information to set it, but we're likely to miss
				// many anyway.
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
func (r *jsonScorecardRawResult) addDependencyUpdateToolRawResults(dut *checker.DependencyUpdateToolData) error {
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
func (r *jsonScorecardRawResult) addBranchProtectionRawResults(bp *checker.BranchProtectionsData) error {
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

	s := &jsonschema.Document{}
	s.Read(&jsonScorecardRawResult{})
	fmt.Println(s)

	return nil
}
