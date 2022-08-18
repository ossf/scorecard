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

package checks

import (
	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/evaluation"
	"github.com/ossf/scorecard/v4/checks/raw"
	sce "github.com/ossf/scorecard/v4/errors"
)

// CheckTokenPermissions is the exported name for Token-Permissions check.
const CheckTokenPermissions = "Token-Permissions"

//nolint:gochecknoinits
func init() {
	supportedRequestTypes := []checker.RequestType{
		checker.FileBased,
		checker.CommitBased,
	}
	if err := registerCheck(CheckTokenPermissions, TokenPermissions, supportedRequestTypes); err != nil {
		// This should never happen.
		panic(err)
	}
}

// TokenPermissions will run the Token-Permissions check.
func TokenPermissions(c *checker.CheckRequest) checker.CheckResult {
	rawData, err := raw.TokenPermissions(c)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckTokenPermissions, e)
	}

	// Return raw results.
	if c.RawResults != nil {
		c.RawResults.TokenPermissionsResults = rawData
	}

	// Return the score evaluation.
	return evaluation.TokenPermissions(CheckTokenPermissions, c.Dlogger, c.RawResults)
}
