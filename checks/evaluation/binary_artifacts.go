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
	"github.com/ossf/scorecard/v3/checker"
	sce "github.com/ossf/scorecard/v3/errors"
)

// BinaryArtifacts applies the score policy for the Binary-Artifacts check.
func BinaryArtifacts(name string, dl checker.DetailLogger,
	r *checker.BinaryArtifactData) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	// Apply the policy evaluation.
	if r.Files == nil || len(r.Files) == 0 {
		return checker.CreateMaxScoreResult(name, "no binaries found in the repo")
	}

	for _, f := range r.Files {
		dl.Warn3(&checker.LogMessage{
			Path: f.Path, Type: checker.FileTypeBinary,
			Text: "binary detected",
		})
	}

	return checker.CreateMinScoreResult(name, "binaries present in source code")
}
