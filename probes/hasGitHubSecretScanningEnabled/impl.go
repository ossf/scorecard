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

package hasGitHubSecretScanningEnabled

import (
	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/checknames"
	"github.com/ossf/scorecard/v5/internal/probes"
	"github.com/ossf/scorecard/v5/probes/internal/utils/uerror"
)

const Probe = "hasGitHubSecretScanningEnabled"

func init() {
	probes.MustRegister(Probe, Run, []checknames.CheckName{checknames.SecretScanning})
}

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", uerror.ErrNil
	}

	// Check if this is a GitHub repository
	if raw.SecretScanningResults.Platform != "github" {
		return []finding.Finding{{
			Probe:   Probe,
			Outcome: finding.OutcomeNotApplicable,
			Message: "Not a GitHub repository",
		}}, Probe, nil
	}

	outcome := finding.OutcomeFalse
	msg := "GitHub secret scanning is disabled"
	switch raw.SecretScanningResults.GHNativeEnabled {
	case checker.TriTrue:
		outcome = finding.OutcomeTrue
		msg = "GitHub secret scanning is enabled"
	case checker.TriFalse:
		outcome = finding.OutcomeFalse
		msg = "GitHub secret scanning is disabled"
	case checker.TriUnknown:
		// No OutcomeNegative in Scorecard. Use OutcomeFalse with an "inconclusive" message.
		outcome = finding.OutcomeFalse
		msg = "GitHub secret scanning status is inconclusive"
	}
	return []finding.Finding{{
		Probe:   Probe,
		Outcome: outcome,
		Message: msg,
	}}, Probe, nil
}
