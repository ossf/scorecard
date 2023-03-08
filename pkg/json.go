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

package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/ossf/scorecard/v4/checker"
	docs "github.com/ossf/scorecard/v4/docs/checks"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/log"
	rules "github.com/ossf/scorecard/v4/rule"
)

// nolint: govet
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

// nolint: govet
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

// JSONScorecardResultV2 exports results as JSON for new detail format.
//
//nolint:govet
type JSONScorecardResultV2 struct {
	Date           string              `json:"date"`
	Repo           jsonRepoV2          `json:"repo"`
	Scorecard      jsonScorecardV2     `json:"scorecard"`
	AggregateScore jsonFloatScore      `json:"score"`
	Checks         []jsonCheckResultV2 `json:"checks"`
	Metadata       []string            `json:"metadata"`
}

// nolint: govet
type jsonCheckResultV3 struct {
	Risk     rules.Risk        `json:"risk"`
	Outcome  finding.Outcome   `json:"outcome"`
	Findings []finding.Finding `json:"findings"`
	Score    int               `json:"score"`
	Reason   string            `json:"reason"`
	Name     string            `json:"name"`
	// TODO(X): list of rules run.
	// TODO(X): simplify the documentation for the overall check
	// and add the rules that are used in the description.
	Doc jsonCheckDocumentationV2 `json:"documentation"`
}

// JSONScorecardResultV3 exports results as JSON for structured detail format.
//
//nolint:govet
type JSONScorecardResultV3 struct {
	Date           string              `json:"date"`
	Repo           jsonRepoV2          `json:"repo"`
	Scorecard      jsonScorecardV2     `json:"scorecard"`
	AggregateScore jsonFloatScore      `json:"score"`
	Checks         []jsonCheckResultV3 `json:"checks"`
	Metadata       []string            `json:"metadata"`
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
		Date:           r.Date.Format(time.RFC3339),
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

func (r *ScorecardResult) AsSJSON(showDetails bool,
	logLevel log.Level, checkDocs docs.Doc, writer io.Writer,
) error {
	score, err := r.GetAggregateScore(checkDocs)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(writer)
	out := JSONScorecardResultV3{
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

		tmpResult := jsonCheckResultV3{
			Name: checkResult.Name,
			Doc: jsonCheckDocumentationV2{
				URL:   doc.GetDocumentationURL(r.Scorecard.CommitSHA),
				Short: doc.GetShort(),
			},
			Reason:  checkResult.Reason,
			Score:   checkResult.Score,
			Risk:    rules.RiskNone,
			Outcome: finding.OutcomePositive,
		}
		if showDetails {
			for i := range checkResult.Details {
				if checkResult.Details[i].Type == checker.DetailDebug && logLevel != log.DebugLevel {
					continue
				}

				f := checkResult.Details[i].Msg.Finding
				if f == nil {
					continue
				}

				if f.Risk.GreaterThan(tmpResult.Risk) {
					tmpResult.Risk = f.Risk
				}

				if f.Outcome.WorseThan(tmpResult.Outcome) {
					tmpResult.Outcome = f.Outcome
				}

				tmpResult.Findings = append(tmpResult.Findings, *f)
			}
		}
		out.Checks = append(out.Checks, tmpResult)
	}
	if err := encoder.Encode(out); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("encoder.Encode: %v", err))
	}

	return nil
}
