// Copyright 2025 OpenSSF Scorecard Authors
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
	"fmt"
	"io"

	intoto "github.com/in-toto/attestation/go/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	docs "github.com/ossf/scorecard/v5/docs/checks"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/log"
)

const (
	InTotoPredicateType = "https://scorecard.dev/result/v0.1"
)

// InTotoPredicate overrides JSONScorecardResultV2 with a nullable Repo field.
type InTotoPredicate struct {
	Repo *jsonRepoV2 `json:"repo,omitempty"`
	JSONScorecardResultV2
}

// AsInTotoResultOption wraps AsJSON2ResultOption preparing it for export as an
// intoto statement.
type AsInTotoResultOption struct {
	AsJSON2ResultOption
}

// AsInToto writes the results as an in-toto attestation statement.
func (r *Result) AsInToto(writer io.Writer, checkDocs docs.Doc, opt *AsInTotoResultOption) error {
	if opt == nil {
		opt = &AsInTotoResultOption{
			AsJSON2ResultOption{
				LogLevel:    log.DefaultLevel,
				Details:     false,
				Annotations: false,
			},
		}
	}

	json2, err := r.resultsToJSON2(checkDocs, &opt.AsJSON2ResultOption)
	if err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	}

	predicate, err := inTotoPredicateToStruct(&json2)
	if err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	}

	stmt := &intoto.Statement{
		Type: intoto.StatementTypeUri,
		// Build the attestation subject from the result Repo.
		Subject: []*intoto.ResourceDescriptor{
			{
				Name: r.Repo.Name,
				Uri:  fmt.Sprintf("git+https://%s@%s", r.Repo.Name, r.Repo.CommitSHA),
				Digest: map[string]string{
					"gitCommit": r.Repo.CommitSHA,
				},
			},
		},
		PredicateType: InTotoPredicateType,
		Predicate:     predicate,
	}
	if err := stmt.Validate(); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("invalid in-toto statement: %v", err))
	}

	// The statement types are generated from protobuf definitions, so they must
	// be serialized with protojson to get the field names defined by the in-toto
	// spec (_type, predicateType, etc).
	data, err := protojson.Marshal(stmt)
	if err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("protojson.Marshal: %v", err))
	}

	// protojson intentionally randomizes its whitespace, so compact the output to
	// make it deterministic.
	var buf bytes.Buffer
	if err := json.Compact(&buf, data); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("json.Compact: %v", err))
	}
	buf.WriteByte('\n')

	if _, err := writer.Write(buf.Bytes()); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("writer.Write: %v", err))
	}

	return nil
}

// inTotoPredicateToStruct converts the JSON results into a protobuf Struct so
// they can be embedded as the predicate of an in-toto statement.
func inTotoPredicateToStruct(json2 *JSONScorecardResultV2) (*structpb.Struct, error) {
	predicate := InTotoPredicate{
		JSONScorecardResultV2: *json2,
		Repo:                  nil,
	}

	// Round trip through JSON to honor the custom marshaling of the result types.
	data, err := json.Marshal(&predicate)
	if err != nil {
		return nil, fmt.Errorf("marshaling predicate: %w", err)
	}

	s := &structpb.Struct{}
	if err := protojson.Unmarshal(data, s); err != nil {
		return nil, fmt.Errorf("converting predicate to struct: %w", err)
	}

	return s, nil
}
