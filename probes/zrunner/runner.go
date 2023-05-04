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

package zrunner

import (
	"errors"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	serrors "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes"
)

var errProbeRun = errors.New("probe run failure")

// Run runs the probes in probesToRun.
func Run(raw *checker.RawResults, probesToRun []probes.ProbeImpl) ([]finding.Finding, error) {
	var results []finding.Finding
<<<<<<< HEAD
	var errs []error
	for _, probeFunc := range probesToRun {
		findings, probeID, err := probeFunc(raw)
		if err != nil {
			errs = append(errs, err)
=======
	for _, probeFunc := range probesToRun {
		findings, probeID, err := probeFunc(raw)
		if err != nil {
>>>>>>> fbcf212a (update)
			results = append(results,
				finding.Finding{
					Probe:   probeID,
					Outcome: finding.OutcomeError,
					Message: serrors.WithMessage(serrors.ErrScorecardInternal, err.Error()).Error(),
				})
			continue
		}
		results = append(results, findings...)
	}
<<<<<<< HEAD
	if len(errs) > 0 {
		return results, fmt.Errorf("%w: %v", errProbeRun, errs)
	}
=======
>>>>>>> fbcf212a (update)
	return results, nil
}
