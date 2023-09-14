// Copyright 2020 OpenSSF Scorecard Authors
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

package checks

import (
	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/evaluation"
	"github.com/ossf/scorecard/v4/checks/raw"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/probes"
	"github.com/ossf/scorecard/v4/probes/zrunner"
)

// CheckDependencyUpdateTool is the exported name for Automatic-Depdendency-Update.
const CheckDependencyUpdateTool = "Dependency-Update-Tool"

// nolint
func init() {
	supportedRequestTypes := []checker.RequestType{
		checker.FileBased,
	}
	if err := registerCheck(CheckDependencyUpdateTool, DependencyUpdateTool, supportedRequestTypes); err != nil {
		// this should never happen
		panic(err)
	}
}

// DependencyUpdateTool checks if the repository uses a dependency update tool.
func DependencyUpdateTool(c *checker.CheckRequest) checker.CheckResult {
	rawData, err := raw.DependencyUpdateTool(c.RepoClient)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckDependencyUpdateTool, e)
	}

	// Set the raw results.
	pRawResults := getRawResults(c)
	pRawResults.DependencyUpdateToolResults = rawData

	// Evaluate the probes.
	findings, err := zrunner.Run(pRawResults, probes.DependencyToolUpdates)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckDependencyUpdateTool, e)
	}

	// Return the score evaluation.
	return evaluation.DependencyUpdateTool(CheckDependencyUpdateTool, findings, c.Dlogger)
}
