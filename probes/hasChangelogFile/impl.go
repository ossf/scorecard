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

package hasChangelogFile

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
	probes.MustRegister(Probe, Run, []checknames.CheckName{checknames.Changelog})
}

//go:embed *.yml
var fs embed.FS

const Probe = "hasChangelogFile"

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	var findings []finding.Finding

	changelogFiles := raw.ChangelogResults.ChangelogFiles
	if len(changelogFiles) == 0 {
		f, err := finding.NewFalse(fs, Probe, "project does not have a changelog file", nil)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
		return findings, Probe, nil
	}

	for i := range changelogFiles {
		loc := changelogFiles[i].Location()
		f, err := finding.NewTrue(fs, Probe, "project has a changelog file", loc)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	}
	return findings, Probe, nil
}
