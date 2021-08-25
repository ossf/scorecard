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

	"go.uber.org/zap/zapcore"

	sce "github.com/ossf/scorecard/v2/errors"
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

//nolint
type jsonCheckResultV2 struct {
	Details []string
	Score   int
	Reason  string
	Name    string
}

type jsonScorecardResultV2 struct {
	Repo     string
	Date     string
	Commit   string
	Checks   []jsonCheckResultV2
	Metadata []string
}

// AsJSON exports results as JSON for new detail format.
func (r *ScorecardResult) AsJSON(showDetails bool, logLevel zapcore.Level, writer io.Writer) error {
	encoder := json.NewEncoder(writer)

	out := jsonScorecardResult{
		Repo:     r.Repo,
		Date:     r.Date.Format("2006-01-02"),
		Metadata: r.Metadata,
	}

	//nolint
	for _, checkResult := range r.Checks {
		tmpResult := jsonCheckResult{
			Name:       checkResult.Name,
			Pass:       checkResult.Pass,
			Confidence: checkResult.Confidence,
		}
		if showDetails {
			for i := range checkResult.Details2 {
				d := checkResult.Details2[i]
				m := detailToString(&d, logLevel)
				if m == "" {
					continue
				}
				tmpResult.Details = append(tmpResult.Details)
			}
		}
		out.Checks = append(out.Checks, tmpResult)
	}
	if err := encoder.Encode(out); err != nil {
		//nolint:wrapcheck
		return sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("encoder.Encode: %v", err))
	}
	return nil
}

// AsJSON2 exports results as JSON for new detail format.
func (r *ScorecardResult) AsJSON2(showDetails bool, logLevel zapcore.Level, writer io.Writer) error {
	encoder := json.NewEncoder(writer)

	out := jsonScorecardResultV2{
		Repo:     r.Repo,
		Date:     r.Date.Format("2006-01-02"),
		Commit:   r.CommitSHA,
		Metadata: r.Metadata,
	}

	//nolint
	for _, checkResult := range r.Checks {
		tmpResult := jsonCheckResultV2{
			Name:   checkResult.Name,
			Reason: checkResult.Reason,
			Score:  checkResult.Score,
		}
		if showDetails {
			for i := range checkResult.Details2 {
				d := checkResult.Details2[i]
				m := detailToString(&d, logLevel)
				if m == "" {
					continue
				}
				tmpResult.Details = append(tmpResult.Details, detailToString(d, logLevel))
			}
		}
		out.Checks = append(out.Checks, tmpResult)
	}
	if err := encoder.Encode(out); err != nil {
		//nolint:wrapcheck
		return sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("encoder.Encode: %v", err))
	}
	return nil
}
