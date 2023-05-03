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
	"github.com/ossf/scorecard/v4/checker"
	serrors "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes"
)

func Run(raw *checker.RawResults, probesToRun []probes.ProbeImpl) ([]finding.Finding, error) {
	var findings []finding.Finding
	for _, item := range probesToRun {
		f := item
		ff, probeID, err := f(raw)
		if err != nil {
			findings = append(findings,
				finding.Finding{
					Probe:   probeID,
					Outcome: finding.OutcomeError,
					Message: serrors.WithMessage(serrors.ErrScorecardInternal, err.Error()).Error(),
				})
			continue
		}
		if len(ff) > 0 {
			findings = append(findings, ff...)
		}
	}
	return findings, nil
}
