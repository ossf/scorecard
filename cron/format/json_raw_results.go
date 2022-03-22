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

	"cloud.google.com/go/bigquery"
	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/pkg"
)

type jsonScorecardRawResult struct {
	// TODO: update Date type to generate type of `DATE` rather then `STRING`.
	Date      string          `json:"date" bigquery:"date"`
	Repo      jsonRepoV2      `json:"repo" bigquery:"repo"`
	Scorecard jsonScorecardV2 `json:"scorecard" bigquery:"scorecard"`
	Metadata  []string        `json:"metadata" bigquery:"metadata"`
	Results   jsonRawResults  `json:"results" bigquery:"results"`
}

// TODO: separate each check extraction into its own file.
type jsonFile struct {
	Path   string `json:"path" bigquery:"path"`
	Offset int    `json:"offset" bigquery:"offset"`
}

type jsonTool struct {
	Name        string     `json:"name" bigquery:"name"`
	URL         string     `json:"url" bigquery:"url"`
	Desc        string     `json:"desc" bigquery:"desc"`
	ConfigFiles []jsonFile `json:"files" bigquery:"files"`
	// TODO: Runs, Issues, Merge requests.
}

type jsonBranchProtectionSettings struct {
	RequiredApprovingReviewCount int  `json:"requiredReviewerCount" bigquery:"requiredReviewerCount"`
	AllowsDeletions              bool `json:"allowsDeletions" bigquery:"allowsDeletions"`
	AllowsForcePushes            bool `json:"allowsForcePushes" bigquery:"allowsForcePushes"`
	RequiresCodeOwnerReviews     bool `json:"requiresCodeOwnerReview" bigquery:"requiresCodeOwnerReview"`
	RequiresLinearHistory        bool `json:"requiredLinearHistory" bigquery:"requiredLinearHistory"`
	DismissesStaleReviews        bool `json:"dismissesStaleReviews" bigquery:"dismissesStaleReviews"`
	EnforcesAdmins               bool `json:"enforcesAdmin" bigquery:"enforcesAdmin"`
	RequiresStatusChecks         bool `json:"requiresStatusChecks" bigquery:"requiresStatusChecks"`
	//nolint
	RequiresUpToDateBranchBeforeMerging bool     `json:"requiresUpdatedBranchesToMerge" bigquery:"requiresUpdatedBranchesToMerge"`
	StatusCheckContexts                 []string `json:"statusChecksContexts" bigquery:"statusChecksContexts"`
}

type jsonBranchProtection struct {
	Protection *jsonBranchProtectionSettings `json:"protection" bigquery:"protection"`
	Name       string                        `json:"name" bigquery:"name"`
}

type jsonReview struct {
	Reviewer jsonUser `json:"reviewer" bigquery:"reviewer"`
	State    string   `json:"state" bigquery:"state"`
}

type jsonUser struct {
	Login string `json:"login" bigquery:"login"`
}

//nolint:govet
type jsonMergeRequest struct {
	Number  int          `json:"number" bigquery:"number"`
	Labels  []string     `json:"labels" bigquery:"labels"`
	Reviews []jsonReview `json:"reviews" bigquery:"reviews"`
	Author  jsonUser     `json:"author" bigquery:"author"`
}

type jsonDefaultBranchCommit struct {
	// ApprovedReviews *jsonApprovedReviews `bigquery:"approved-reviews"`
	Committer     jsonUser          `json:"committer" bigquery:"committer"`
	MergeRequest  *jsonMergeRequest `json:"mergeRequest" bigquery:"mergeRequest"`
	CommitMessage string            `json:"commitMessage" bigquery:"commitMessage"`
	SHA           string            `json:"sha" bigquery:"sha"`

	// TODO: check runs, etc.
}

type jsonDatabaseVulnerability struct {
	// For OSV: OSV-2020-484
	// For CVE: CVE-2022-23945
	ID string `json:"id" bigquery:"id"`
	// TODO: additional information
}

type jsonRawResults struct {
	DatabaseVulnerabilities []jsonDatabaseVulnerability `json:"databaseVulnerabilities" bigquery:"databaseVulnerabilities"`
	// List of binaries found in the repo.
	Binaries []jsonFile `json:"binaries" bigquery:"binaries"`
	// List of security policy files found in the repo.
	// Note: we return one at most.
	SecurityPolicies []jsonFile `json:"securityPolicies" bigquery:"securityPolicies"`
	// List of update tools.
	// Note: we return one at most.
	DependencyUpdateTools []jsonTool `json:"dependencyUpdateTools" bigquery:"dependencyUpdateTools"`
	// Branch protection settings for development and release branches.
	BranchProtections []jsonBranchProtection `json:"branchProtections" bigquery:"branchProtections"`
	// Commits.
	DefaultBranchCommits []jsonDefaultBranchCommit `json:"defaultBranchCommits" bigquery:"defaultBranchCommits"`
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

//nolint
func addBranchProtectionRawResults(r *jsonScorecardRawResult, bp *checker.BranchProtectionsData) error {
	r.Results.BranchProtections = []jsonBranchProtection{}
	for _, v := range bp.Branches {
		var bp *jsonBranchProtectionSettings
		if v.Protected != nil && *v.Protected {
			bp = &jsonBranchProtectionSettings{
				StatusCheckContexts: v.StatusCheckContexts,
			}

			if v.AllowsDeletions != nil {
				bp.AllowsDeletions = *v.AllowsDeletions
			}
			if v.AllowsForcePushes != nil {
				bp.AllowsForcePushes = *v.AllowsForcePushes
			}
			if v.RequiresCodeOwnerReviews != nil {
				bp.RequiresCodeOwnerReviews = *v.RequiresCodeOwnerReviews
			}
			if v.RequiresLinearHistory != nil {
				bp.RequiresLinearHistory = *v.RequiresLinearHistory
			}
			if v.DismissesStaleReviews != nil {
				bp.DismissesStaleReviews = *v.DismissesStaleReviews
			}
			if v.EnforcesAdmins != nil {
				bp.EnforcesAdmins = *v.EnforcesAdmins
			}
			if v.RequiresStatusChecks != nil {
				bp.RequiresStatusChecks = *v.RequiresStatusChecks
			}
			if v.RequiresUpToDateBranchBeforeMerging != nil {
				bp.RequiresUpToDateBranchBeforeMerging = *v.RequiresUpToDateBranchBeforeMerging
			}
			if v.RequiresStatusChecks != nil {
				bp.RequiredApprovingReviewCount = *v.RequiredApprovingReviewCount
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

// https://github.com/googleapis/google-cloud-go/blob/bigquery/v1.30.0/bigquery/schema.go#L544
type bigQueryJSONField struct {
	Description string              `json:"description"`
	Fields      []bigQueryJSONField `json:"fields,omitempty"`
	Mode        string              `json:"mode"`
	Name        string              `json:"name"`
	Type        string              `json:"type"`
}

func generateSchema(schema bigquery.Schema) []bigQueryJSONField {
	var bqs []bigQueryJSONField
	for _, fs := range schema {
		bq := bigQueryJSONField{
			Description: fs.Description,
			Name:        fs.Name,
			Type:        string(fs.Type),
			Fields:      generateSchema(fs.Schema),
		}
		// https://github.com/googleapis/google-cloud-go/blob/bigquery/v1.30.0/bigquery/schema.go#L125

		switch {
		// Make all fields optional to give us flexibility:
		// discard `fs.Required`.
		case fs.Repeated:
			bq.Mode = "REPEATED"
		default:
			bq.Mode = "NULLABLE"
		}

		bqs = append(bqs, bq)
	}

	return bqs
}

// GenerateBqSchema generates the BQ schema in JSON format.
func GenerateBqSchema(r *pkg.ScorecardResult, writer io.Writer) error {
	schema, err := bigquery.InferSchema(jsonScorecardRawResult{})
	if err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	}
	jsonFields := generateSchema(schema)

	jsonData, err := json.Marshal(jsonFields)
	if err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("json.Marshal: %v", err.Error()))
	}

	_, err = writer.Write(jsonData)
	if err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf(" writer.Write: %v", err.Error()))
	}

	return nil
}
