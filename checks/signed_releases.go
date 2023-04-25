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

// CheckSignedReleases is the registered name for SignedReleases.
const CheckSignedReleases = "Signed-Releases"

//nolint:gochecknoinits
func init() {
	if err := registerCheck(CheckSignedReleases, SignedReleases, nil); err != nil {
		// this should never happen
		panic(err)
	}
}

// SignedReleases runs Signed-Releases check.
func SignedReleases(c *checker.CheckRequest) checker.CheckResult {
	rawData, err := raw.SignedReleases(c)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckSignedReleases, e)
	}

	// Return raw results.
	if c.RawResults != nil {
		c.RawResults.SignedReleasesResults = rawData
	}

	// Return the score evaluation.
	return evaluation.SignedReleases(CheckSignedReleases, c.Dlogger, &rawData)
}
