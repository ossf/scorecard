// Copyright 2023 OpenSSF Scorecard Authors
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

var (
	binaryGradleSafe          = "BinaryGradleWrapperSafe"
	binaryArtifactsNotPresent = "BinaryArtifactsNotPresent"
)

// BinaryArtifacts applies the score policy for the Binary-Artifacts check.
func BinaryArtifacts(name string, dl checker.DetailLogger,
	r *checker.BinaryArtifactData,
) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	// Keep track of reported results.
	reportedRuleResults := make(map[string]bool)

	// Apply the policy evaluation.
	if r.Files == nil || len(r.Files) == 0 {
		// Report findings for all rules.
		if err := logDefaultFindings(dl, reportedRuleResults); err != nil {
			return checker.CreateRuntimeErrorResult(name, err)
		}
		return checker.CreateMaxScoreResult(name, "no binaries found in the repo")
	}

	score := checker.MaxResultScore
	for _, f := range r.Files {
		//nolint:gosec
		loc := checker.LocationFromFile(&f)
		metadata := map[string]string{"path": f.Path}
		switch {
		// BinaryGradleWrapperSafe case.
		case path.Base(f.Path) == "gradle-wrapper.jar":
			if err := checker.LogFinding(rules, binaryGradleSafe,
				"unsafe gradle-wrapper.jar found",
				loc, finding.OutcomeNegative, metadata, dl); err != nil {
				return checker.CreateRuntimeErrorResult(name, err)
			}
			reportedRuleResults[binaryGradleSafe] = true
		// Default case BinaryArtifactsNotPresent.
		default:
			if err := checker.LogFinding(rules, binaryArtifactsNotPresent,
				"binary found",
				loc, finding.OutcomeNegative, metadata, dl); err != nil {
				return checker.CreateRuntimeErrorResult(name, err)
			}
			reportedRuleResults[binaryArtifactsNotPresent] = true
		}

		// We remove one point for each binary.
		score--
	}

	// Report findings for all rules.
	if err := logDefaultFindings(dl, reportedRuleResults); err != nil {
		return checker.CreateRuntimeErrorResult(name, err)
	}

	if score < checker.MinResultScore {
		score = checker.MinResultScore
	}

	return checker.CreateResultWithScore(name, "binaries present in source code", score)
}

func logDefaultFindings(dl checker.DetailLogger, r map[string]bool) error {
	// We always report at least one finding for each rule.
	if !r[binaryArtifactsNotPresent] {
		if err := checker.LogFinding(rules, binaryArtifactsNotPresent,
			"no binaries found",
			nil, finding.OutcomePositive, nil, dl); err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		}
	}
	if !r[binaryGradleSafe] {
		if err := checker.LogFinding(rules, binaryGradleSafe,
			"no unsafe gradle wrapper binaries found",
			// No wrapper binary found, so the rule is not applicable.
			nil, finding.OutcomeNotApplicable, nil, dl); err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		}
	}
	return nil
}
