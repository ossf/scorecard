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

package checks

import (
	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/checks/evaluation"
	"github.com/ossf/scorecard/v5/checks/raw"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/probes"
	zrunner "github.com/ossf/scorecard/v5/probes/zrunner"
)

const CheckMTTUDependencies = "MTTUDependencies"

//nolint:gochecknoinits
func init() {
	supportedRequestTypes := []checker.RequestType{
		checker.FileBased,
	}
	if err := registerCheck(CheckMTTUDependencies, MTTUDependencies, supportedRequestTypes); err != nil {
		// this should never happen
		panic(err)
	}
}

// indirections to ease testing.
var runProbes = zrunner.Run

func MTTUDependencies(c *checker.CheckRequest) checker.CheckResult {
	// 1) get raw data
	rawData, err := raw.MTTUDependencies(c)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckMTTUDependencies, e)
	}

	// 2) set raw results
	pRawResults := getRawResults(c)
	pRawResults.MTTUDependenciesResults = rawData

	// 3) run probes
	findings, err := runProbes(pRawResults, probes.MTTUDependencies)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckMTTUDependencies, e)
	}

	// 4) evaluate
	ret := evaluation.MTTUDependencies(CheckMTTUDependencies, findings, c.Dlogger)
	return ret
}
