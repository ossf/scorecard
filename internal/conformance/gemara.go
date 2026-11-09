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
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ossf/scorecard/v5/finding"
)

// GemaraVersion is the Gemara schema release the Evaluation Log targets.
const GemaraVersion = "1.6.0"

// Gemara #Result values.
const (
	gemaraPassed        = "Passed"
	gemaraFailed        = "Failed"
	gemaraUnknown       = "Unknown"
	gemaraNeedsReview   = "Needs Review"
	gemaraNotApplicable = "Not Applicable"
)

const (
	evidenceType = "ScorecardProbeFinding"
	scorecardURI = "https://github.com/ossf/scorecard"
)

var (
	errNoResults          = errors.New("no requirements were evaluated, and a Gemara Evaluation Log needs at least one")
	errUnknownRequirement = errors.New("result for a requirement with no mapping in the release")
)

// The types below follow the Gemara v1.6.0 #EvaluationLog schema. They are
// Scorecard's own, so Scorecard does not depend on a Gemara Go module.

type gemaraLog struct {
	Metadata    gemaraMetadata      `json:"metadata"`
	Target      gemaraEntity        `json:"target"`
	Result      string              `json:"result"`
	Evaluations []gemaraControlEval `json:"evaluations"`
}

type gemaraMetadata struct {
	ID                string                   `json:"id"`
	Type              string                   `json:"type"`
	GemaraVersion     string                   `json:"gemara-version"`
	Date              string                   `json:"date"`
	Description       string                   `json:"description"`
	Author            gemaraEntity             `json:"author"`
	MappingReferences []gemaraMappingReference `json:"mapping-references"`
}

type gemaraEntity struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version,omitempty"`
	URI     string `json:"uri,omitempty"`
}

type gemaraMappingReference struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Version string `json:"version"`
	URL     string `json:"url,omitempty"`
}

type gemaraEntryMapping struct {
	ReferenceID string `json:"reference-id"`
	EntryID     string `json:"entry-id"`
}

type gemaraControlEval struct {
	Control        gemaraEntryMapping `json:"control"`
	Name           string             `json:"name"`
	Result         string             `json:"result"`
	Message        string             `json:"message"`
	AssessmentLogs []gemaraAssessment `json:"assessment-logs"`
}

type gemaraAssessment struct {
	Requirement   gemaraEntryMapping `json:"requirement"`
	Description   string             `json:"description"`
	Result        string             `json:"result"`
	Message       string             `json:"message"`
	Start         string             `json:"start,omitempty"`
	Applicability []string           `json:"applicability"`
	Steps         []string           `json:"steps"`
	Evidence      []gemaraEvidence   `json:"evidence,omitempty"`
}

type gemaraEvidence struct {
	Payload     finding.Finding `json:"payload"`
	ID          string          `json:"id"`
	Type        string          `json:"type"`
	CollectedAt string          `json:"collected-at"`
	Description string          `json:"description,omitempty"`
}

// gemaraResult maps a result to its Gemara #Result value.
func gemaraResult(r Result) string {
	switch r {
	case Pass:
		return gemaraPassed
	case Fail:
		return gemaraFailed
	case NotApplicable:
		return gemaraNotApplicable
	case NeedsReview:
		return gemaraNeedsReview
	case Unknown:
		return gemaraUnknown
	default:
		return gemaraUnknown
	}
}

// aggregateRank orders Gemara results for roll-up. A higher rank wins, as
// in the Gemara Go SDK's UpdateAggregateResult: Failed, then Unknown, then
// Needs Review, then Passed, then Not Applicable.
func aggregateRank(gemara string) int {
	switch gemara {
	case gemaraFailed:
		return 4
	case gemaraUnknown:
		return 3
	case gemaraNeedsReview:
		return 2
	case gemaraPassed:
		return 1
	default:
		return 0
	}
}

func aggregate(results []string) string {
	agg := gemaraNotApplicable
	for _, r := range results {
		if aggregateRank(r) > aggregateRank(agg) {
			agg = r
		}
	}
	return agg
}

// AsGemara writes the report as a Gemara Layer 5 Evaluation Log in JSON.
// Requirements that were not evaluated are left out of the log.
func (r *Report) AsGemara(w io.Writer) error {
	if len(r.Results) == 0 {
		return errNoResults
	}
	date := r.Date.UTC().Format(time.RFC3339)
	rel := r.Release

	var evals []gemaraControlEval
	index := map[string]int{}
	for i := range r.Results {
		res := &r.Results[i]
		req, reqOK := rel.Requirement(res.Requirement)
		m, mapOK := rel.Mapping(res.Requirement)
		if !reqOK || !mapOK {
			return fmt.Errorf("%w: %s in %s %s", errUnknownRequirement, res.Requirement, rel.Catalog, rel.Version)
		}
		a := gemaraAssessment{
			Requirement:   gemaraEntryMapping{ReferenceID: rel.Catalog, EntryID: req.ID},
			Description:   m.Description,
			Result:        gemaraResult(res.Result),
			Message:       res.Explanation,
			Applicability: []string{r.Profile},
			Steps:         m.Steps,
		}
		if res.Result != NotApplicable {
			a.Start = date
		}
		for j := range res.Evidence {
			f := &res.Evidence[j]
			id := f.Probe
			if b, ok := f.Values["branchName"]; ok {
				id += "/" + b
			}
			a.Evidence = append(a.Evidence, gemaraEvidence{
				ID:          id,
				Type:        evidenceType,
				CollectedAt: date,
				Payload:     *f,
				Description: f.Message,
			})
		}

		ci, ok := index[req.Control]
		if !ok {
			ci = len(evals)
			index[req.Control] = ci
			evals = append(evals, gemaraControlEval{
				Name:    m.ControlTitle,
				Control: gemaraEntryMapping{ReferenceID: rel.Catalog, EntryID: req.Control},
			})
		}
		evals[ci].AssessmentLogs = append(evals[ci].AssessmentLogs, a)
	}

	all := make([]string, 0, len(evals))
	for i := range evals {
		e := &evals[i]
		results := make([]string, 0, len(e.AssessmentLogs))
		parts := make([]string, 0, len(e.AssessmentLogs))
		for j := range e.AssessmentLogs {
			a := &e.AssessmentLogs[j]
			results = append(results, a.Result)
			parts = append(parts, a.Requirement.EntryID+": "+a.Result)
		}
		e.Result = aggregate(results)
		e.Message = strings.Join(parts, "; ")
		all = append(all, e.Result)
	}

	out := gemaraLog{
		Metadata: gemaraMetadata{
			ID:            fmt.Sprintf("scorecard:%s@%s:%s@%s:%s", r.Repo, r.Commit, rel.Catalog, rel.Version, r.Profile),
			Type:          "EvaluationLog",
			GemaraVersion: GemaraVersion,
			Date:          date,
			Description:   r.description(),
			Author: gemaraEntity{
				ID:      "scorecard",
				Name:    "OpenSSF Scorecard",
				Type:    "Software",
				Version: r.ScorecardVersion,
				URI:     scorecardCommitURI(r.ScorecardCommit),
			},
			MappingReferences: []gemaraMappingReference{{
				ID:      rel.Catalog,
				Title:   rel.Title,
				Version: rel.Version,
				URL:     rel.URL,
			}},
		},
		Target: gemaraEntity{
			ID:      r.Repo,
			Name:    r.Repo,
			Type:    "Software",
			Version: r.Commit,
			URI:     repoURI(r.Repo),
		},
		Result:      aggregate(all),
		Evaluations: evals,
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		return fmt.Errorf("encoding Gemara Evaluation Log: %w", err)
	}
	return nil
}

// description states what the log covers, so that a partial run is not
// read as conformance with the whole catalog.
func (r *Report) description() string {
	evaluated := 0
	for i := range r.Results {
		if r.Results[i].Result != NotApplicable {
			evaluated++
		}
	}
	return fmt.Sprintf("Experimental OpenSSF Scorecard evaluation of %s against %s %s (%s). "+
		"%d of the %d requirements in the profile were evaluated; the other %d were not evaluated "+
		"and are not in this log.",
		r.Repo, r.Release.Catalog, r.Release.Version, r.Profile,
		evaluated, evaluated+len(r.NotEvaluated), len(r.NotEvaluated))
}

// scorecardCommitURI links to the Scorecard commit that produced the log,
// so the log records the evaluator's commit as well as its version. It
// falls back to the repository URL when the build has no commit.
func scorecardCommitURI(commit string) string {
	if commit == "" || strings.Trim(commit, "0123456789abcdef") != "" {
		return scorecardURI
	}
	return scorecardURI + "/commit/" + commit
}

// repoURI returns a URI for a repository name such as
// "github.com/owner/repo". Names that already carry a scheme, such as
// local "file://" paths, are returned unchanged.
func repoURI(name string) string {
	if strings.Contains(name, "://") {
		return name
	}
	return "https://" + name
}
