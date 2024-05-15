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
package hasReleaseSBOM

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/probes"
	"github.com/ossf/scorecard/v5/probes/internal/utils/uerror"
)

func init() {
	probes.MustRegister(Probe, Run, []probes.CheckName{probes.SBOM})
}

//go:embed *.yml
var fs embed.FS

const (
	Probe           = "hasReleaseSBOM"
	AssetNameKey    = "assetName"
	AssetURLKey     = "assetURL"
	missingSbom     = "Project is not publishing an SBOM file as part of a release or CICD"
	releaseLookBack = 5
)

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	var findings []finding.Finding
	var msg string

	SBOMFiles := raw.SBOMResults.SBOMFiles

	for i := range SBOMFiles {
		if i >= releaseLookBack {
			break
		}

		SBOMFile := SBOMFiles[i]

		if SBOMFile.File.Type != finding.FileTypeURL {
			continue
		}

		loc := SBOMFile.File.Location()
		msg = "Project publishes an SBOM file as part of a release or CICD"
		f, err := finding.NewTrue(fs, Probe, msg, loc)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		f.Values = map[string]string{
			AssetNameKey: SBOMFile.Name,
			AssetURLKey:  SBOMFile.File.Path,
		}
		findings = append(findings, *f)
	}

	if len(findings) == 0 {
		msg = missingSbom
		f, err := finding.NewFalse(fs, Probe, msg, nil)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	}

	return findings, Probe, nil
}
