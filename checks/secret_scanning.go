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

// CheckSecretScanning is the registered name for the Secret-Scanning check.
const CheckSecretScanning = "SecretScanning"

//nolint:gochecknoinits
func init() {
	if err := registerCheck(CheckSecretScanning, SecretScanning, nil); err != nil {
		// this should never happen
		panic(err)
	}
}

// SecretScanning runs the Secret-Scanning check.
func SecretScanning(c *checker.CheckRequest) checker.CheckResult {
	// Collect raw data for this check.
	rawData, err := raw.SecretScanning(c)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckSecretScanning, e)
	}

	// Store raw results so probes can read them.
	pRawResults := getRawResults(c)
	pRawResults.SecretScanningResults = rawData

	// Run the probes associated with this check.
	findings, err := zrunner.Run(pRawResults, probes.SecretScanning)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckSecretScanning, e)
	}

	// Evaluate findings into a 0–10 score and reason.
	ret := evaluation.SecretScanning(
		CheckSecretScanning,
		findings,
		c.Dlogger,
		&rawData,
	)
	ret.Findings = findings
	return ret
}
