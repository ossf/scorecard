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

	docs "github.com/ossf/scorecard/v3/docs/checks"
	sce "github.com/ossf/scorecard/v3/errors"
)

//nolint
type jsonRawCheckResult struct {
	Name       string
	RawResults interface{}
}

//nolint
type jsonRawCheckResultV6 struct {
	Name       string                   `json:"name"`
	Doc        jsonCheckDocumentationV2 `json:"documentation"`
	RawResults interface{}              `json:"results"`
}

type jsonScorecardRawResultV6 struct {
	Date      string                 `json:"date"`
	Repo      jsonRepoV2             `json:"repo"`
	Scorecard jsonScorecardV2        `json:"scorecard"`
	Checks    []jsonRawCheckResultV6 `json:"checks"`
	Metadata  []string               `json:"metadata"`
}

// AsJSON exports results as JSON for new detail format.
func (r *ScorecardRawResult) AsJSON(checkDocs docs.Doc, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	out := jsonScorecardRawResultV6{
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

	//nolint
	for _, checkResult := range r.Checks {
		doc, e := checkDocs.GetCheck(checkResult.Name)
		if e != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("GetCheck: %s: %v", checkResult.Name, e))
		}

		tmpResult := jsonRawCheckResultV6{
			Name: checkResult.Name,
			Doc: jsonCheckDocumentationV2{
				URL:   doc.GetDocumentationURL(r.Scorecard.CommitSHA),
				Short: doc.GetShort(),
			},
			// TODO: create a level of indirection for raw results.
			RawResults: checkResult.RawResults,
		}

		out.Checks = append(out.Checks, tmpResult)
	}
	if err := encoder.Encode(out); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("encoder.Encode: %v", err))
	}

	return nil
}
