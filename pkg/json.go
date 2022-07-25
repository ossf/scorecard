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

	docs "github.com/ossf/scorecard/v4/docs/checks"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/log"
)

//nolint
type jsonCheckResult struct {
	Name       string
	Details    []string
	Confidence int
	Pass       bool
}

type jsonScorecardResult struct {
	Repo     string
	Date     string
	Checks   []jsonCheckResult
	Metadata []string
}

type jsonCheckDocumentationV2 struct {
	URL   string `json:"url"`
	Short string `json:"short"`
	// Can be extended if needed.
}

//nolint
type jsonCheckResultV2 struct {
	Details []string                 `json:"details"`
	Score   int                      `json:"score"`
	Reason  string                   `json:"reason"`
	Name    string                   `json:"name"`
	Doc     jsonCheckDocumentationV2 `json:"documentation"`
}

type jsonRepoV2 struct {
	Name   string `json:"name"`
	Commit string `json:"commit"`
}

type jsonScorecardV2 struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

type jsonFloatScore float64

func (s jsonFloatScore) MarshalJSON() ([]byte, error) {
	// Note: for integers, this will show as X.0.
	return []byte(fmt.Sprintf("%.1f", s)), nil
}

//nolint:govet
// JSONScorecardResultV2 exports results as JSON for new detail format.
type JSONScorecardResultV2 struct {
	Date           string              `json:"date"`
	Repo           jsonRepoV2          `json:"repo"`
	Scorecard      jsonScorecardV2     `json:"scorecard"`
	AggregateScore jsonFloatScore      `json:"score"`
	Checks         []jsonCheckResultV2 `json:"checks"`
	Metadata       []string            `json:"metadata"`
}

// JSONDependencydiffResult exports dependency-diff check results as JSON for new detail format.
type JSONDependencydiffResult struct {
	ChangeType       *ChangeType `json:"changeType"`
	PackageURL       *string     `json:"packageUrl"`
	SourceRepository *string     `json:"sourceRepository"`
	ManifestPath     *string     `json:"manifestPath"`
	Ecosystem        *string     `json:"ecosystem"`
	Version          *string     `json:"packageVersion"`
	// TODO (issue#2078): map the current naming convention (GitHub) to the OSV naming convention.
	Name                string                 `json:"packageName"`
	JSONScorecardResult *JSONScorecardResultV2 `json:"scorecardResult"`
}

// AsJSON exports results as JSON for new detail format.
func (r *ScorecardResult) AsJSON(showDetails bool, logLevel log.Level, writer io.Writer) error {
	encoder := json.NewEncoder(writer)

	out := jsonScorecardResult{
		Repo:     r.Repo.Name,
		Date:     r.Date.Format("2006-01-02"),
		Metadata: r.Metadata,
	}

	for _, checkResult := range r.Checks {
		tmpResult := jsonCheckResult{
			Name: checkResult.Name,
		}
		if showDetails {
			for i := range checkResult.Details {
				d := checkResult.Details[i]
				m := DetailToString(&d, logLevel)
				if m == "" {
					continue
				}
				tmpResult.Details = append(tmpResult.Details, m)
			}
		}
		out.Checks = append(out.Checks, tmpResult)
	}
	if err := encoder.Encode(out); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("encoder.Encode: %v", err))
	}
	return nil
}

// AsJSON2 exports results as JSON for new detail format.
func (r *ScorecardResult) AsJSON2(showDetails bool,
	logLevel log.Level, checkDocs docs.Doc, writer io.Writer,
) error {
	score, err := r.GetAggregateScore(checkDocs)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(writer)
	out := JSONScorecardResultV2{
		Repo: jsonRepoV2{
			Name:   r.Repo.Name,
			Commit: r.Repo.CommitSHA,
		},
		Scorecard: jsonScorecardV2{
			Version: r.Scorecard.Version,
			Commit:  r.Scorecard.CommitSHA,
		},
		Date:           r.Date.Format("2006-01-02"),
		Metadata:       r.Metadata,
		AggregateScore: jsonFloatScore(score),
	}

	for _, checkResult := range r.Checks {
		doc, e := checkDocs.GetCheck(checkResult.Name)
		if e != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("GetCheck: %s: %v", checkResult.Name, e))
		}

		tmpResult := jsonCheckResultV2{
			Name: checkResult.Name,
			Doc: jsonCheckDocumentationV2{
				URL:   doc.GetDocumentationURL(r.Scorecard.CommitSHA),
				Short: doc.GetShort(),
			},
			Reason: checkResult.Reason,
			Score:  checkResult.Score,
		}
		if showDetails {
			for i := range checkResult.Details {
				d := checkResult.Details[i]
				m := DetailToString(&d, logLevel)
				if m == "" {
					continue
				}
				tmpResult.Details = append(tmpResult.Details, m)
			}
		}
		out.Checks = append(out.Checks, tmpResult)
	}
	if err := encoder.Encode(out); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("encoder.Encode: %v", err))
	}
	return nil
}

// DependencydiffResultsAsJSON exports dependencydiff results as JSON. This cannot be defined as the OOP-like
// ScorecardResult.AsJSON since we return a slice of DependencyCheckResult.
func DependencydiffResultsAsJSON(depdiffResults []DependencyCheckResult,
	logLevel log.Level, doc docs.Doc, writer io.Writer,
) error {
	out := []JSONDependencydiffResult{}
	for _, dr := range depdiffResults {
		// Copy every DependencydiffResult struct to a JSONDependencydiffResult for exporting as JSON.
		jsonDepdiff := JSONDependencydiffResult{
			ChangeType:       dr.ChangeType,
			PackageURL:       dr.PackageURL,
			SourceRepository: dr.SourceRepository,
			ManifestPath:     dr.ManifestPath,
			Ecosystem:        dr.Ecosystem,
			Version:          dr.Version,
			Name:             dr.Name,
		}
		scResult := dr.ScorecardResultWithError.ScorecardResult
		if scResult != nil {
			score, err := scResult.GetAggregateScore(doc)
			if err != nil {
				return err
			}
			jsonResult := JSONScorecardResultV2{
				Repo: jsonRepoV2{
					Name:   scResult.Repo.Name,
					Commit: scResult.Repo.CommitSHA,
				},
				Scorecard: jsonScorecardV2{
					Version: scResult.Scorecard.Version,
					Commit:  scResult.Scorecard.CommitSHA,
				},
				Date:           scResult.Date.Format("2006-01-02"),
				Metadata:       scResult.Metadata,
				AggregateScore: jsonFloatScore(score),
			}
			for _, c := range scResult.Checks {
				doc, e := doc.GetCheck(c.Name)
				if e != nil {
					return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("GetCheck: %s: %v", c.Name, e))
				}
				tmpResult := jsonCheckResultV2{
					Name: c.Name,
					Doc: jsonCheckDocumentationV2{
						URL:   doc.GetDocumentationURL(scResult.Scorecard.CommitSHA),
						Short: doc.GetShort(),
					},
					Reason: c.Reason,
					Score:  c.Score,
				}
				for i := range c.Details {
					d := c.Details[i]
					m := DetailToString(&d, logLevel)
					if m == "" {
						continue
					}
					tmpResult.Details = append(tmpResult.Details, m)
				}
				jsonResult.Checks = append(jsonResult.Checks, tmpResult)
				jsonDepdiff.JSONScorecardResult = &jsonResult
			}
		}
		out = append(out, jsonDepdiff)
	}
	encoder := json.NewEncoder(writer)
	if err := encoder.Encode(out); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("encoder.Encode: %v", err))
	}
	return nil
}
