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

// Package main implements cron worker job.
package main

import (
	"encoding/json"
	"fmt"
	"io"

	// nolint:gosec
	_ "net/http/pprof"

	"go.uber.org/zap/zapcore"

	sce "github.com/ossf/scorecard/v2/errors"
	"github.com/ossf/scorecard/v2/pkg"
)

//nolint
type jsonCheckCronResult struct {
	Name       string
	Details    []string
	Confidence int
	Pass       bool
}

type jsonScorecardCronResult struct {
	Repo     string
	Date     string
	Checks   []jsonCheckCronResult
	Metadata []string
}

//nolint
type jsonCheckCronResultV2 struct {
	Details []string
	Score   int
	Reason  string
	Name    string
}

type jsonScorecardCronResultV2 struct {
	Repo     string
	Date     string
	Commit   string
	Checks   []jsonCheckCronResultV2
	Metadata []string
}

// AsJSON exports results as JSON for new detail format.
func AsJSON(r *pkg.ScorecardResult, showDetails bool, logLevel zapcore.Level, writer io.Writer) error {
	encoder := json.NewEncoder(writer)

	out := jsonScorecardCronResult{
		Repo:     r.Repo.Name,
		Date:     r.Date.Format("2006-01-02"),
		Metadata: r.Metadata,
	}

	//nolint
	for _, checkResult := range r.Checks {
		tmpResult := jsonCheckCronResult{
			Name:       checkResult.Name,
			Pass:       checkResult.Pass,
			Confidence: checkResult.Confidence,
		}
		if showDetails {
			for i := range checkResult.Details2 {
				d := checkResult.Details2[i]
				m := pkg.DetailToString(&d, logLevel)
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

// AsJSON2 exports results as JSON for the cron job and in the new detail format.
func AsJSON2(r *pkg.ScorecardResult, showDetails bool, logLevel zapcore.Level, writer io.Writer) error {
	encoder := json.NewEncoder(writer)

	out := jsonScorecardCronResultV2{
		Repo:     r.Repo.Name,
		Date:     r.Date.Format("2006-01-02"),
		Commit:   r.Repo.CommitSHA,
		Metadata: r.Metadata,
	}

	//nolint
	for _, checkResult := range r.Checks {
		tmpResult := jsonCheckCronResultV2{
			Name:   checkResult.Name,
			Reason: checkResult.Reason,
			Score:  checkResult.Score,
		}
		if showDetails {
			for i := range checkResult.Details2 {
				d := checkResult.Details2[i]
				m := pkg.DetailToString(&d, logLevel)
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
