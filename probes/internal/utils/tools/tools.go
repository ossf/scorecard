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

package tools

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
)

type toolMatcher interface {
	Name() string
	Matches(*checker.Tool) bool
}

// ToolsRun runs the probe for a tool.
// The function iterates thru the raw results and searches for a tool of interest that is used on a repository.
// The function uses 'matcher' to identify the tool of interest.
// If a tool is used in the repository, it creates a finding with the 'foundOutcome'.
// If not, it returns a finding with outcome 'notFoundOutcome'.
func Run(tools []checker.Tool, fs embed.FS, probeID string,
	foundOutcome, notFoundOutcome finding.Outcome, matcher toolMatcher,
) ([]finding.Finding, string, error) {
	var findings []finding.Finding
	for i := range tools {
		tool := &tools[i]
		if !matcher.Matches(tool) {
			continue
		}

		var loc *finding.Location
		if len(tool.Files) > 0 {
			loc = tool.Files[0].Location()
		}

		f, err := finding.NewWith(fs, probeID, fmt.Sprintf("tool '%s' is used", tool.Name),
			loc, foundOutcome)
		if err != nil {
			return nil, probeID, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	}

	// No tools found.
	if len(findings) == 0 {
		f, err := finding.NewWith(fs, probeID, fmt.Sprintf("tool '%s' is not used", matcher.Name()),
			nil, notFoundOutcome)
		if err != nil {
			return nil, probeID, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	}

	return findings, probeID, nil
}
