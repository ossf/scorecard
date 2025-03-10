// Copyright 2024 OpenSSF Scorecard Authors
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

package scorecard

import (
	"bytes"
	"encoding/json"
	"slices"
	"testing"
	"time"

	"github.com/ossf/scorecard/v5/finding"
)

func TestInToto(t *testing.T) {
	t.Parallel()
	// The intoto statement generation relies on the same generation as
	// the json output, so here we just check for correct assignments
	result := Result{
		Repo: RepoInfo{
			Name:      "github.com/example/example",
			CommitSHA: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
		Scorecard: ScorecardInfo{
			Version:   "1.2.3",
			CommitSHA: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		},
		Date: time.Date(2024, time.February, 1, 13, 48, 0, 0, time.UTC),
		Findings: []finding.Finding{
			{
				Probe:   "check for X",
				Outcome: finding.OutcomeTrue,
				Message: "found X",
				Location: &finding.Location{
					Path: "some/path/to/file",
					Type: finding.FileTypeText,
				},
			},
			{
				Probe:   "check for Y",
				Outcome: finding.OutcomeFalse,
				Message: "did not find Y",
			},
		},
	}
	var w bytes.Buffer
	err := result.AsInToto(&w, jsonMockDocRead(), nil)
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	// Unmarshal the written json to a generic map
	stmt := statement{}
	if err := json.Unmarshal(w.Bytes(), &stmt); err != nil {
		t.Error("error unmarshaling statement", err)
		return
	}

	// Check the data
	if len(stmt.Subject) != 1 {
		t.Error("unexpected statement subject length")
	}
	if stmt.Subject[0].GetDigest()["gitCommit"] != result.Repo.CommitSHA {
		t.Error("mismatched statement subject digest")
	}
	if stmt.Subject[0].GetName() != result.Repo.Name {
		t.Error("mismatched statement subject name")
	}

	if stmt.PredicateType != InTotoPredicateType {
		t.Error("incorrect predicate type", stmt.PredicateType)
	}

	// Check the predicate
	if stmt.Predicate.Scorecard.Commit != result.Scorecard.CommitSHA {
		t.Error("mismatch in scorecard commit")
	}
	if stmt.Predicate.Scorecard.Version != result.Scorecard.Version {
		t.Error("mismatch in scorecard version")
	}
	if stmt.Predicate.Repo != nil {
		t.Error("repo should be null")
	}
	if !slices.Equal(stmt.Predicate.Metadata, result.Metadata) {
		t.Error("mismatched metadata")
	}
}
