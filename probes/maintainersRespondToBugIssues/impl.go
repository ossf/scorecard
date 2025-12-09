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

package maintainersRespondToBugIssues

import (
	"fmt"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
)

// Probe is the stable ID for this probe.
const Probe = "maintainersRespondToBugIssues"

// Threshold in days for violation.
const thresholdDays = 180

// Run consumes the raw intervals computed by checks/raw/maintainer_response.go
// and emits exactly one finding per issue:
//
//   - OutcomeTrue  => maintainers responded within 180 days
//   - OutcomeFalse => maintainers did NOT respond within 180 days (violation)
//   - OutcomeNotApplicable => issue never had tracked labels (excluded from denominator)
//
// "Response" means a comment from a maintainer after label was applied.
// Tracked labels are defined in checks/raw/maintainer_response.TrackedLabels.
func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	data := raw.MaintainerResponseResults
	var out []finding.Finding

	for _, it := range data.Items {
		// No tracked labeling at all → NotApplicable (not in denominator).
		if len(it.HadLabelIntervals) == 0 {
			out = append(out, finding.Finding{
				Probe:   Probe,
				Outcome: finding.OutcomeNotApplicable,
				Message: fmt.Sprintf("issue #%d had no tracked labels", it.IssueNumber),
				Location: &finding.Location{
					Path: it.IssueURL, // include URL if available
				},
			})
			continue
		}

		// Scan intervals for violation and track worst duration for messaging.
		violates := false
		worst := 0
		for _, iv := range it.HadLabelIntervals {
			if iv.DurationDays > worst {
				worst = iv.DurationDays
			}
			// Violation iff: label interval ≥ 180 days AND no reaction in that interval.
			if iv.DurationDays >= thresholdDays && iv.ResponseAt == nil {
				violates = true
			}
		}

		if violates {
			out = append(out, finding.Finding{
				Probe:   Probe,
				Outcome: finding.OutcomeFalse,
				Message: fmt.Sprintf("issue #%d exceeded %d days without maintainer response (worst %d days)",
					it.IssueNumber, thresholdDays, worst),
				Location: &finding.Location{
					Path: it.IssueURL,
				},
			})
		} else {
			out = append(out, finding.Finding{
				Probe:   Probe,
				Outcome: finding.OutcomeTrue,
				Message: fmt.Sprintf("issue #%d received maintainer response within %d days (worst %d days)",
					it.IssueNumber, thresholdDays, worst),
				Location: &finding.Location{
					Path: it.IssueURL,
				},
			})
		}
	}

	return out, Probe, nil
}
