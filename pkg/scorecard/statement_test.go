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

	intoto "github.com/in-toto/attestation/go/v1"
	"google.golang.org/protobuf/encoding/protojson"

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

	// The output must be stable across runs.
	var w2 bytes.Buffer
	if err := result.AsInToto(&w2, jsonMockDocRead(), nil); err != nil {
		t.Error("unexpected error: ", err)
	}
	if !bytes.Equal(w.Bytes(), w2.Bytes()) {
		t.Error("statement output is not deterministic")
	}

	// The statement must use the field names defined by the in-toto spec,
	// not the names encoding/json derives from the generated protobuf types.
	raw := map[string]json.RawMessage{}
	if err := json.Unmarshal(w.Bytes(), &raw); err != nil {
		t.Fatal("error unmarshaling statement to map", err)
	}
	for _, key := range []string{"_type", "subject", "predicateType", "predicate"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("statement is missing the %q field", key)
		}
	}
	for _, key := range []string{"type", "predicate_type"} {
		if _, ok := raw[key]; ok {
			t.Errorf("statement has unexpected field %q", key)
		}
	}

	// Unmarshal the written json to an in-toto statement
	stmt := &intoto.Statement{}
	if err := protojson.Unmarshal(w.Bytes(), stmt); err != nil {
		t.Fatal("error unmarshaling statement", err)
	}
	if err := stmt.Validate(); err != nil {
		t.Error("statement failed validation", err)
	}

	// Check the data
	if stmt.GetType() != intoto.StatementTypeUri {
		t.Error("incorrect statement type", stmt.GetType())
	}
	if len(stmt.GetSubject()) != 1 {
		t.Fatal("unexpected statement subject length")
	}
	if stmt.GetSubject()[0].GetDigest()["gitCommit"] != result.Repo.CommitSHA {
		t.Error("mismatched statement subject digest")
	}
	if stmt.GetSubject()[0].GetName() != result.Repo.Name {
		t.Error("mismatched statement subject name")
	}

	if stmt.GetPredicateType() != InTotoPredicateType {
		t.Error("incorrect predicate type", stmt.GetPredicateType())
	}

	// Check the predicate
	predicateJSON, err := protojson.Marshal(stmt.GetPredicate())
	if err != nil {
		t.Fatal("error marshaling predicate", err)
	}
	predicate := InTotoPredicate{}
	if err := json.Unmarshal(predicateJSON, &predicate); err != nil {
		t.Fatal("error unmarshaling predicate", err)
	}
	if predicate.Scorecard.Commit != result.Scorecard.CommitSHA {
		t.Error("mismatch in scorecard commit")
	}
	if predicate.Scorecard.Version != result.Scorecard.Version {
		t.Error("mismatch in scorecard version")
	}
	if predicate.Repo != nil {
		t.Error("repo should be null")
	}
	if !slices.Equal(predicate.Metadata, result.Metadata) {
		t.Error("mismatched metadata")
	}
}
