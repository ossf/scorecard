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

package MTTUDependenciesIsLow

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
const Probe = "MTTUDependenciesIsLow"

// Run emits OutcomeTrue iff 2 weeks <= mean < 6 months.
func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	meantime, evaluated, problems, err := mttu.MeanTimeSinceFirstNewer(raw)
	if err != nil {
		return nil, Probe, fmt.Errorf("computing mean MTTU: %w", err)
	}

	thresholdLow := 14 * 24 * time.Hour
	thresholdHigh := 180 * 24 * time.Hour

	days := int(meantime.Hours() / 24)
	months := days / 30
	text := fmt.Sprintf("meantime is %dd (~%d months) based on %d dependencies", days, months, evaluated)

	if len(problems) > 0 {
		text += fmt.Sprintf(" (issues: %d dependencies skipped)", len(problems))
	}

	outcome := finding.OutcomeFalse
	if meantime >= thresholdLow && meantime < thresholdHigh {
		outcome = finding.OutcomeTrue
	}

	f, err := finding.NewWith(fs, Probe, text, nil, outcome)
	if err != nil {
		return nil, Probe, fmt.Errorf("create finding: %w", err)
	}
	return []finding.Finding{*f}, Probe, nil
}
