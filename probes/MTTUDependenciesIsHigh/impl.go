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

package MTTUDependenciesIsHigh

import (
	"embed"
	"fmt"
	"time"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/internal/utils/mttu"
)

//go:embed *.yml
var fs embed.FS

// Probe identifies this probe in emitted findings.
const Probe = "MTTUDependenciesIsHigh"

// Run computes the mean time since the oldest newer release and emits a single Finding.
// OutcomeTrue iff mean >= 6 months (defined as 180 days).
func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	meantime, evaluated, problems, err := mttu.MeanTimeSinceFirstNewer(raw)
	if err != nil {
		return nil, Probe, fmt.Errorf("computing mean MTTU: %w", err)
	}

	const daysPerMonth = 30
	threshold := time.Duration(6*daysPerMonth) * 24 * time.Hour

	days := int(meantime.Hours() / 24)
	months := days / 30
	text := fmt.Sprintf("meantime is %dd (~%d months) based on %d dependencies", days, months, evaluated)

	if len(problems) > 0 && len(problems) <= 3 {
		text += fmt.Sprintf(" (issues: %d dependencies skipped)", len(problems))
	} else if len(problems) > 3 {
		text += fmt.Sprintf(" (issues: %d dependencies skipped)", len(problems))
	}

	outcome := finding.OutcomeFalse
	if meantime >= threshold {
		outcome = finding.OutcomeTrue
	}

	f, err := finding.NewWith(fs, Probe, text, nil, outcome)
	if err != nil {
		return nil, Probe, fmt.Errorf("create finding: %w", err)
	}
	return []finding.Finding{*f}, Probe, nil
}
