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
)

// CheckLicense is the registered name for License.
const CheckLicense = "License"

//nolint:gochecknoinits
func init() {
	supportedRequestTypes := []checker.RequestType{
		checker.CommitBased,
	}
	if err := registerCheck(CheckLicense, License, supportedRequestTypes); err != nil {
		// this should never happen
		panic(err)
	}
}

// License runs License check.
func License(c *checker.CheckRequest) checker.CheckResult {
	rawData, err := raw.License(c)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckLicense, e)
	}

	// Set the raw results.
	pRawResults := getRawResults(c)
	pRawResults.LicenseResults = rawData

	// Evaluate the probes.
	findings, err := evaluateProbes(c, pRawResults, probes.License)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckContributors, e)
	}

	return evaluation.License(CheckLicense, findings, c.Dlogger)
}
