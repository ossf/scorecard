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

// Package model defines the provider-agnostic data types that the MCP server
// exposes to clients.
package model

import "strings"

// InconclusiveScore is the score Scorecard reports when a check (or the
// aggregate) could not be evaluated. It is not a failing score.
const InconclusiveScore = -1

// Source identifies which provider produced a Result.
type Source string

const (
	// SourceCachedREST indicates results came from the cached public REST API.
	SourceCachedREST Source = "cached-rest"
	// SourceLiveLocal indicates results came from an in-process Scorecard run.
	SourceLiveLocal Source = "live-local"
)

// RepoRef identifies a repository on a supported platform.
type RepoRef struct {
	Platform string `json:"platform" jsonschema:"the hosting platform, github.com or gitlab.com"`
	Org      string `json:"org" jsonschema:"the owner or organization"`
	Name     string `json:"name" jsonschema:"the repository name"`
}

// String returns the platform/org/name form.
func (r RepoRef) String() string { return r.Platform + "/" + r.Org + "/" + r.Name }

// Documentation links a check to its published documentation.
type Documentation struct {
	Short string `json:"short,omitempty"`
	URL   string `json:"url,omitempty"`
}

// Check is a single Scorecard check result. A Score of -1
// (InconclusiveScore) means the check could not be evaluated.
type Check struct {
	Documentation *Documentation `json:"documentation,omitempty"`
	Name          string         `json:"name"`
	Reason        string         `json:"reason,omitempty"`
	Details       []string       `json:"details,omitempty"`
	Score         int            `json:"score" jsonschema:"score from 0-10, or -1 for inconclusive"`
}

// ScorecardInfo records the Scorecard tool version that produced a Result.
type ScorecardInfo struct {
	Version string `json:"version,omitempty"`
	Commit  string `json:"commit,omitempty"`
}

// Attribution records the license and source of the data.
type Attribution struct {
	DataLicense string `json:"data_license"`
	SourceURL   string `json:"source_url"`
}

// Result is a Scorecard result with provenance and caveats attached. A Score of
// -1 (InconclusiveScore) means the aggregate could not be evaluated.
type Result struct {
	Attribution *Attribution  `json:"attribution,omitempty"`
	Repo        RepoRef       `json:"repo"`
	Scorecard   ScorecardInfo `json:"scorecard"`
	Commit      string        `json:"commit,omitempty"`
	Date        string        `json:"date,omitempty"`
	Source      Source        `json:"source"`
	Checks      []Check       `json:"checks"`
	Caveats     []string      `json:"caveats,omitempty"`
	Annotations []string      `json:"annotations,omitempty"`
	// MissingChecks and Complete are computed by the server from the check
	// catalog, not by the provider.
	MissingChecks []string `json:"missing_checks,omitempty"`
	Score         float64  `json:"score"`
	Complete      bool     `json:"complete"`
}

// FindCheck returns the check with the given name (case-insensitive).
func (r *Result) FindCheck(name string) (Check, bool) {
	for _, c := range r.Checks {
		if strings.EqualFold(c.Name, name) {
			return c, true
		}
	}
	return Check{}, false
}
