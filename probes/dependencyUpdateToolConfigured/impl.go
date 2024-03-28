// Copyright 2024 OpenSSF Scorecard Authors
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
package dependencyUpdateToolConfigured

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/internal/probes"
	"github.com/ossf/scorecard/v4/probes/internal/utils/uerror"
)

func init() {
	probes.MustRegister(Probe, Run, []probes.CheckName{probes.DependencyUpdateTool})
}

//go:embed *.yml
var fs embed.FS

const (
	Probe   = "dependencyUpdateToolConfigured"
	ToolKey = "tool"
)

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, Probe, fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	tools := raw.DependencyUpdateToolResults.Tools
	if len(tools) == 0 {
		f, err := finding.NewNegative(fs, Probe, "no dependency update tool configurations found", nil)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		return []finding.Finding{*f}, Probe, nil
	}

	var findings []finding.Finding
	for i := range tools {
		tool := &tools[i]

		var loc *finding.Location
		if len(tool.Files) > 0 {
			loc = tool.Files[0].Location()
		}

		f, err := finding.NewPositive(fs, Probe, "detected update tool: "+tool.Name, loc)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		f = f.WithValue(ToolKey, tool.Name)
		findings = append(findings, *f)
	}

	return findings, Probe, nil
}
