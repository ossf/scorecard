// Copyright 2020 Security Scorecard Authors
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

// CheckSecurityPolicy is the registred name for SecurityPolicy.
const CheckSecurityPolicy = "Security-Policy"

//nolint:gochecknoinits
func init() {
	supportedRequestTypes := []checker.RequestType{
		checker.FileBased,
	}
	if err := registerCheck(CheckSecurityPolicy, SecurityPolicy, supportedRequestTypes); err != nil {
		// This should never happen.
		panic(err)
	}
}

// SecurityPolicy runs Security-Policy check.
func SecurityPolicy(c *checker.CheckRequest) checker.CheckResult {
	rawData, err := raw.SecurityPolicy(c)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckSecurityPolicy, e)
	}

	// Set the raw results.
	if c.RawResults != nil {
		c.RawResults.SecurityPolicyResults = rawData
		return checker.CheckResult{}
	}

	return evaluation.SecurityPolicy(CheckSecurityPolicy, c.Dlogger, &rawData)
}
