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
package hasBinaryArtifacts

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
	probes.MustRegister(Probe, Run, []checknames.CheckName{checknames.BinaryArtifacts})
}

//go:embed *.yml
var fs embed.FS

const Probe = "hasBinaryArtifacts"

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	r := raw.BinaryArtifactResults
	var findings []finding.Finding

	// Apply the policy evaluation.
	if len(r.Files) == 0 {
		f, err := finding.NewWith(fs, Probe,
			"Repository does not have any binary artifacts.", nil,
			finding.OutcomeFalse)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	}
	for i := range r.Files {
		file := &r.Files[i]
		f, err := finding.NewWith(fs, Probe, "binary artifact detected",
			nil, finding.OutcomeTrue)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		f = f.WithLocation(&finding.Location{
			Path:      file.Path,
			LineStart: &file.Offset,
			Type:      file.Type,
		})
		findings = append(findings, *f)
	}

	return findings, Probe, nil
}
