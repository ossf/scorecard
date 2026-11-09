// Copyright 2026 OpenSSF Scorecard Authors
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

package conformance

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type jsonReport struct {
	Date         string              `json:"date"`
	Repo         jsonRepo            `json:"repo"`
	Scorecard    jsonScorecard       `json:"scorecard"`
	Catalog      jsonCatalog         `json:"catalog"`
	Profile      string              `json:"profile"`
	Results      []RequirementResult `json:"results"`
	NotEvaluated []string            `json:"notEvaluated"`
	Experimental bool                `json:"experimental"`
}

type jsonRepo struct {
	Name   string `json:"name"`
	Commit string `json:"commit"`
}

type jsonScorecard struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

type jsonCatalog struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

// AsJSON writes the report in the conformance JSON format.
func (r *Report) AsJSON(w io.Writer) error {
	out := jsonReport{
		Experimental: true,
		Date:         r.Date.UTC().Format(time.RFC3339),
		Repo:         jsonRepo{Name: r.Repo, Commit: r.Commit},
		Scorecard:    jsonScorecard{Version: r.ScorecardVersion, Commit: r.ScorecardCommit},
		Catalog:      jsonCatalog{ID: r.Release.Catalog, Version: r.Release.Version},
		Profile:      r.Profile,
		Results:      r.Results,
		NotEvaluated: r.NotEvaluated,
	}
	if out.Results == nil {
		out.Results = []RequirementResult{}
	}
	if out.NotEvaluated == nil {
		out.NotEvaluated = []string{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		return fmt.Errorf("encoding conformance JSON: %w", err)
	}
	return nil
}
