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

package evaluation

import (
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/fileparser"
	sce "github.com/ossf/scorecard/v4/errors"
)

// PinningDependencies applies the score policy for the Pinned-Dependencies check.
func PinningDependencies(name string, dl checker.DetailLogger,
	r *checker.PinningDependenciesData,
) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}
	for i := range r.Dependencies {
		rr := r.Dependencies[i]
		if rr.File == nil {
			e := sce.WithMessage(sce.ErrScorecardInternal, "empty File field")
			return checker.CreateRuntimeErrorResult(name, e)
		}

		text, err := generateText(&rr)
		if err != nil {
			return checker.CreateRuntimeErrorResult(name, err)
		}

		dl.Warn(&checker.LogMessage{
			Path:        rr.File.Path,
			Type:        rr.File.Type,
			Offset:      rr.File.Offset,
			EndOffset:   rr.File.EndOffset,
			Text:        text,
			Snippet:     rr.File.Snippet,
			Remediation: rr.Remediation,
		})
	}

	return checker.CreateMaxScoreResult(name, "TODO")
}

func generateText(rr *checker.Dependency) (string, error) {
	if rr.Type == checker.DependencyUseTypeGHAction {
		// Check if we are dealing with a GitHub action or a third-party one.
		gitHubOwned := fileparser.IsGitHubOwnedAction(rr.File.Snippet)
		owner := generateOwnerToDisplay(gitHubOwned)
		return fmt.Sprintf("%s %s not pinned by hash", owner, rr.Type), nil
	}

	return fmt.Sprintf("%s not pinned by hash", rr.Type), nil
}

func generateOwnerToDisplay(gitHubOwned bool) string {
	if gitHubOwned {
		return "GitHub-owned"
	}
	return "third-party"
}
