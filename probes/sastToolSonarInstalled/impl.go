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
package sastToolSonarInstalled

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/internal/utils/uerror"
)

//go:embed *.yml
var fs embed.FS

const Probe = "sastToolSonarInstalled"

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	r := raw.SASTResults

	for _, wf := range r.Workflows {
		if wf.Type != checker.SonarWorkflow {
			continue
		}
		wf := wf
		loc := &finding.Location{
			Type:      wf.File.Type,
			Path:      wf.File.Path,
			LineStart: &wf.File.Offset,
			LineEnd:   &wf.File.EndOffset,
			Snippet:   &wf.File.Snippet,
		}
		f, err := finding.NewWith(fs, Probe,
			"SAST tool detected: Sonar", loc,
			finding.OutcomePositive)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		return []finding.Finding{*f}, Probe, nil
	}
	f, err := finding.NewWith(fs, Probe,
		"Sonar tool not detected", nil,
		finding.OutcomeNegative)
	if err != nil {
		return nil, Probe, fmt.Errorf("create finding: %w", err)
	}
	return []finding.Finding{*f}, Probe, nil
}
