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
)

// CheckBranchProtection is the exported name for Branch-Protected check.
const CheckBranchProtection = "Branch-Protection"

//nolint:gochecknoinits
func init() {
	if err := registerCheck(CheckBranchProtection, BranchProtection, nil); err != nil {
		// this should never happen
		panic(err)
	}
}

// BranchProtection runs the Branch-Protection check.
func BranchProtection(c *checker.CheckRequest) checker.CheckResult {
	rawData, err := raw.BranchProtection(c.RepoClient)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckBranchProtection, e)
	}

	// Return raw results.
	if c.RawResults != nil {
		c.RawResults.BranchProtectionResults = rawData
	}

	// Return the score evaluation.
	return evaluation.BranchProtection(CheckBranchProtection, c.Dlogger, &rawData)
}
