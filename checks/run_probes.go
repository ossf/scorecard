// Copyright 2023 OpenSSF Scorecard Authors
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
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes"
	"github.com/ossf/scorecard/v4/probes/zrunner"
)

// evaluateProbes runs the probes in probesToRun.
func evaluateProbes(rawResults *checker.RawResults,
	probesToRun []probes.ProbeImpl,
) ([]finding.Finding, error) {
	// Run the probes.
	findings, err := zrunner.Run(rawResults, probesToRun)
	if err != nil {
		return nil, fmt.Errorf("zrunner.Run: %w", err)
	}
	return findings, nil
}

// getRawResults returns a pointer to the raw results in the CheckRequest
// if the pointer is not nil. Else, it creates a new raw result.
func getRawResults(c *checker.CheckRequest) *checker.RawResults {
	if c.RawResults != nil {
		return c.RawResults
	}
	return &checker.RawResults{}
}
