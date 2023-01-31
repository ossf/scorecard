// Copyright 2021 OpenSSF Scorecard Authors
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

package evaluation

import (
	"embed"
	"path"

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
)

//go:embed *.yml
var rules embed.FS

// BinaryArtifacts applies the score policy for the Binary-Artifacts check.
func BinaryArtifacts(name string, dl checker.DetailLogger,
	r *checker.BinaryArtifactData,
) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	reportedRuleResults := map[string]bool{
		"BinaryGraddleWrapperSafe":  false,
		"BinaryArtifactsNotPresent": false,
	}

	// Apply the policy evaluation.
	if r.Files == nil || len(r.Files) == 0 {
		if err := logDefaultFindings(dl, reportedRuleResults); err != nil {
			return checker.CreateRuntimeErrorResult(name, err)
		}
		return checker.CreateMaxScoreResult(name, "no binaries found in the repo")
	}

	score := checker.MaxResultScore
	for _, f := range r.Files {
		loc := checker.LocationFromFile(&f)
		switch {
		// BinaryGraddleWrapperSafe case.
		case path.Base(f.Path) == "graddle-wrapper.jar":
			if err := checker.LogFinding(rules, "BinaryGraddleWrapperSafe",
				"unsafe graddle-wrapper.jar found",
				loc, finding.OutcomeNegative, dl); err != nil {
				return checker.CreateRuntimeErrorResult(name, err)
			}
			reportedRuleResults["BinaryGraddleWrapperSafe"] = true
		// Default case BinaryArtifactsNotPresent.
		default:
			if err := checker.LogFinding(rules, "BinaryArtifactsNotPresent",
				"binary found",
				loc, finding.OutcomeNegative, dl); err != nil {
				return checker.CreateRuntimeErrorResult(name, err)
			}
			reportedRuleResults["BinaryArtifactsNotPresent"] = true
		}

		// We remove one point for each binary.
		score--
	}

	if err := logDefaultFindings(dl, reportedRuleResults); err != nil {
		return checker.CreateRuntimeErrorResult(name, err)
	}

	if score < checker.MinResultScore {
		score = checker.MinResultScore
	}

	return checker.CreateResultWithScore(name, "binaries present in source code", score)
}

func logDefaultFindings(dl checker.DetailLogger, r map[string]bool) error {
	if !r["BinaryArtifactsNotPresent"] {
		if err := checker.LogFinding(rules, "BinaryArtifactsNotPresent",
			"no binaries found",
			nil, finding.OutcomePositive, dl); err != nil {
			return err
		}
	}

	if !r["BinaryGraddleWrapperSafe"] {
		if err := checker.LogFinding(rules, "BinaryGraddleWrapperSafe",
			"no unsafe graddle wrapper binaries found",
			nil, finding.OutcomeNotApplicable, dl); err != nil {
			return err
		}
	}
	return nil
}
