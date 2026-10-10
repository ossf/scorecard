// Copyright 2024 OpenSSF Scorecard Authors
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
	"os"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/checks/evaluation"
	"github.com/ossf/scorecard/v5/checks/raw"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/probes"
	"github.com/ossf/scorecard/v5/probes/zrunner"
)

// CheckChangelog is the registered name for Changelog.
const CheckChangelog = "Changelog"

//nolint:gochecknoinits
func init() {
	supportedRequestTypes := []checker.RequestType{
		checker.CommitBased,
		checker.FileBased,
	}
	if err := registerCheck(CheckChangelog, Changelog, supportedRequestTypes); err != nil {
		// this should never happen
		panic(err)
	}
}

// Changelog runs the Changelog check.
func Changelog(c *checker.CheckRequest) checker.CheckResult {
	_, enabled := os.LookupEnv("SCORECARD_EXPERIMENTAL")
	if !enabled {
		c.Dlogger.Warn(&checker.LogMessage{
			Text: "SCORECARD_EXPERIMENTAL is not set, not running the Changelog check",
		})

		e := sce.WithMessage(sce.ErrUnsupportedCheck, "SCORECARD_EXPERIMENTAL is not set, not running the Changelog check")
		return checker.CreateRuntimeErrorResult(CheckChangelog, e)
	}

	rawData, err := raw.Changelog(c)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckChangelog, e)
	}

	// Set the raw results.
	pRawResults := getRawResults(c)
	pRawResults.ChangelogResults = rawData

	// Evaluate the probes.
	findings, err := zrunner.Run(pRawResults, probes.Changelog)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckChangelog, e)
	}

	ret := evaluation.Changelog(CheckChangelog, findings, c.Dlogger)
	ret.Findings = findings
	return ret
}
