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

// nolint:stylecheck
package hasDangerousWorkflowScriptInjection

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/internal/utils/uerror"
)

//go:embed *.yml
var fs embed.FS

const Probe = "hasDangerousWorkflowScriptInjection"

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	r := raw.DangerousWorkflowResults

	if r.NumWorkflows == 0 {
		f, err := finding.NewWith(fs, Probe,
			"Project does not have any workflows.", nil,
			finding.OutcomeNotApplicable)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		return []finding.Finding{*f}, Probe, nil
	}

	var findings []finding.Finding
	for _, e := range r.Workflows {
		e := e
		if e.Type == checker.DangerousWorkflowScriptInjection {
			f, err := finding.NewWith(fs, Probe,
				fmt.Sprintf("script injection with untrusted input '%v'", e.File.Snippet),
				nil, finding.OutcomeNegative)
			if err != nil {
				return nil, Probe, fmt.Errorf("create finding: %w", err)
			}
			f = f.WithLocation(&finding.Location{
				Path:      e.File.Path,
				Type:      e.File.Type,
				LineStart: &e.File.Offset,
				Snippet:   &e.File.Snippet,
			})
			findings = append(findings, *f)
		}
	}

	if len(findings) == 0 {
		return positiveOutcome()
	}

	return findings, Probe, nil
}

func positiveOutcome() ([]finding.Finding, string, error) {
	f, err := finding.NewWith(fs, Probe,
		"Project does not have dangerous workflow(s) with possibility of script injection.", nil,
		finding.OutcomePositive)
	if err != nil {
		return nil, Probe, fmt.Errorf("create finding: %w", err)
	}
	return []finding.Finding{*f}, Probe, nil
}
