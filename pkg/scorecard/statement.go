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
	"encoding/json"
	"fmt"
	"io"

	intoto "github.com/in-toto/attestation/go/v1"

	docs "github.com/ossf/scorecard/v5/docs/checks"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/log"
)

const (
	InTotoPredicateType = "https://scorecard.dev/result/v0.1"
)

type statement struct {
	Predicate InTotoPredicate `json:"predicate"`
	intoto.Statement
}

// Predicate overrides JSONScorecardResultV2 with a nullable Repo field.
type InTotoPredicate struct {
	Repo *jsonRepoV2 `json:"repo,omitempty"`
	JSONScorecardResultV2
}

// AsInTotoResultOption wraps AsJSON2ResultOption preparing it for export as an
// intoto statement.
type AsInTotoResultOption struct {
	AsJSON2ResultOption
}

// AsStatement converts the results as an in-toto statement.
func (r *Result) AsStatement(writer io.Writer, checkDocs docs.Doc, opt *AsInTotoResultOption) error {
	// Build the attestation subject from the result Repo.
	subject := intoto.ResourceDescriptor{
		Name: r.Repo.Name,
		Uri:  fmt.Sprintf("git+https://%s@%s", r.Repo.Name, r.Repo.CommitSHA),
		Digest: map[string]string{
			"gitCommit": r.Repo.CommitSHA,
		},
	}

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

	out := statement{
		Statement: intoto.Statement{
			Type: intoto.StatementTypeUri,
			Subject: []*intoto.ResourceDescriptor{
				&subject,
			},
			PredicateType: InTotoPredicateType,
		},
		Predicate: InTotoPredicate{
			JSONScorecardResultV2: json2,
			Repo:                  nil,
		},
	}

	encoder := json.NewEncoder(writer)
	if err := encoder.Encode(&out); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("encoder.Encode: %v", err))
	}

	return nil
}
