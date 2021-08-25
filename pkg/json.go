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

	"github.com/ossf/scorecard/v2/checker"
	sce "github.com/ossf/scorecard/v2/errors"
)

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

// AsJSON outputs the result in JSON format with a newline at the end.
// If called on []ScorecardResult will create NDJson formatted output.
// UPGRADEv2: will be removed.
func (r *ScorecardResult) AsJSON(showDetails bool, logLevel zapcore.Level, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	if showDetails {
		if err := encoder.Encode(r); err != nil {
			//nolint:wrapcheck
			return sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("encoder.Encode: %v", err))
		}
		return nil
	}
	out := ScorecardResult{
		Repo:     r.Repo,
		Date:     r.Date,
		Metadata: r.Metadata,
	}
	// UPGRADEv2: remove nolint after ugrade.
	//nolint
	for _, checkResult := range r.Checks {
		tmpResult := checker.CheckResult{
			Name:       checkResult.Name,
			Pass:       checkResult.Pass,
			Confidence: checkResult.Confidence,
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
		Date:     r.Date,
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
			for _, d := range checkResult.Details2 {
				tmpResult.Details = append(tmpResult.Details, d.Msg.Text)
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
