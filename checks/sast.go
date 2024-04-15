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
	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/checks/evaluation"
	"github.com/ossf/scorecard/v5/checks/raw"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/probes"
	"github.com/ossf/scorecard/v5/probes/zrunner"
)

// CheckSAST is the registered name for SAST.
const CheckSAST = "SAST"

//nolint:gochecknoinits
func init() {
	if err := registerCheck(CheckSAST, SAST, nil); err != nil {
		// This should never happen.
		panic(err)
	}
}

// SAST runs SAST check.
func SAST(c *checker.CheckRequest) checker.CheckResult {
	rawData, err := raw.SAST(c)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckSAST, e)
	}

	// Set the raw results.
	pRawResults := getRawResults(c)
	pRawResults.SASTResults = rawData

	// Evaluate the probes.
	findings, err := zrunner.Run(pRawResults, probes.SAST)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckSAST, e)
	}

	// Return the score evaluation.
	ret := evaluation.SAST(CheckSAST, findings, c.Dlogger)
	ret.Findings = findings
	return ret
}
