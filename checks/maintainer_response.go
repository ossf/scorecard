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
	"github.com/ossf/scorecard/v5/probes/zrunner"
)

// CheckMaintainerResponse is the registered name for this check.
const CheckMaintainerResponse = "Maintainer-Response-BugSecurity"

//nolint:gochecknoinits
func init() {
	if err := registerCheck(CheckMaintainerResponse, MaintainerResponse, nil); err != nil {
		// this should never happen
		panic(err)
	}
}

// MaintainerResponse executes the maintainer-response check.
// Flow: build raw -> store raw -> run probes -> evaluate findings.
func MaintainerResponse(c *checker.CheckRequest) checker.CheckResult {
	// 1) Build raw data (issues + label intervals + reactions).
	rawData, err := raw.MaintainerResponse(c)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckMaintainerResponse, e)
	}

	// 2) Attach raw to request context so probes can read it.
	pRawResults := getRawResults(c)
	pRawResults.MaintainerResponseResults = rawData

	// 3) Run the probe(s) for this check to produce findings.
	findings, err := zrunner.Run(pRawResults, probes.MaintainersRespondToBugIssues)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckMaintainerResponse, e)
	}

	// 4) Evaluate the findings and return the scored result.
	ret := evaluation.MaintainerResponse(CheckMaintainerResponse, findings, c.Dlogger)
	ret.Findings = findings
	return ret
}
