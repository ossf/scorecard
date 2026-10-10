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

package securityPolicyEnablesPrivateReporting

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/checknames"
	"github.com/ossf/scorecard/v5/internal/probes"
	"github.com/ossf/scorecard/v5/probes/internal/utils/uerror"
)

func init() {
	probes.MustRegister(Probe, Run, []checknames.CheckName{checknames.SecurityPolicy})
}

//go:embed *.yml
var fs embed.FS

const Probe = "securityPolicyEnablesPrivateReporting"

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	pvr := raw.SecurityPolicyResults.PrivateVulnerabilityReportingEnabled

	var findings []finding.Finding

	if pvr == nil {
		// Unknown — platform does not expose this signal.
		f, err := finding.NewWith(fs, Probe, "private vulnerability reporting status unknown",
			nil, finding.OutcomeNotApplicable)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		f = f.WithRemediationMetadata(raw.Metadata.Metadata)
		findings = append(findings, *f)
	} else if *pvr {
		// Enabled — good practice.
		f, err := finding.NewWith(fs, Probe, "private vulnerability reporting is enabled",
			nil, finding.OutcomeTrue)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		f = f.WithRemediationMetadata(raw.Metadata.Metadata)
		findings = append(findings, *f)
	} else {
		// Disabled — not enabled.
		f, err := finding.NewWith(fs, Probe, "private vulnerability reporting is not enabled",
			nil, finding.OutcomeFalse)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		f = f.WithRemediationMetadata(raw.Metadata.Metadata)
		findings = append(findings, *f)
	}

	return findings, Probe, nil
}
