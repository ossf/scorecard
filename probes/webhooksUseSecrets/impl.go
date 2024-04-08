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

//nolint:stylecheck
package webhooksUseSecrets

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/internal/probes"
	"github.com/ossf/scorecard/v4/probes/internal/utils/uerror"
)

func init() {
	probes.MustRegister(Probe, Run, []probes.CheckName{probes.Webhooks})
}

//go:embed *.yml
var fs embed.FS

const Probe = "webhooksUseSecrets"

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	r := raw.WebhookResults
	var findings []finding.Finding

	if len(r.Webhooks) == 0 {
		f, err := finding.NewWith(fs, Probe,
			"Repository does not have webhooks.", nil,
			finding.OutcomeNotApplicable)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
		return findings, Probe, nil
	}

	for _, hook := range r.Webhooks {
		if hook.UsesAuthSecret {
			msg := "Webhook with token authorization found."
			f, err := finding.NewWith(fs, Probe,
				msg, nil, finding.OutcomeTrue)
			if err != nil {
				return nil, Probe, fmt.Errorf("create finding: %w", err)
			}
			f = f.WithLocation(&finding.Location{
				Path: hook.Path,
			})
			findings = append(findings, *f)
		} else {
			msg := "Webhook without token authorization found."
			f, err := finding.NewWith(fs, Probe,
				msg, nil, finding.OutcomeFalse)
			if err != nil {
				return nil, Probe, fmt.Errorf("create finding: %w", err)
			}
			f = f.WithLocation(&finding.Location{
				Path: hook.Path,
			})
			findings = append(findings, *f)
		}
	}

	return findings, Probe, nil
}
