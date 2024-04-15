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
package hasOSVVulnerabilities

import (
	"embed"
	"errors"
	"fmt"
	"strings"

	"github.com/google/osv-scanner/pkg/grouper"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/probes"
	"github.com/ossf/scorecard/v5/probes/internal/utils/uerror"
)

func init() {
	probes.MustRegister(Probe, Run, []probes.CheckName{probes.Vulnerabilities})
}

//go:embed *.yml
var fs embed.FS

const Probe = "hasOSVVulnerabilities"

var errNoVulnID = errors.New("no vuln ID")

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	var findings []finding.Finding

	// if no vulns were found
	if len(raw.VulnerabilitiesResults.Vulnerabilities) == 0 {
		f, err := finding.NewWith(fs, Probe,
			"Project does not contain OSV vulnerabilities", nil,
			finding.OutcomeFalse)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
		return findings, Probe, nil
	}

	aliasVulnerabilities := []grouper.IDAliases{}
	for _, vuln := range raw.VulnerabilitiesResults.Vulnerabilities {
		aliasVulnerabilities = append(aliasVulnerabilities, grouper.IDAliases(vuln))
	}

	IDs := grouper.Group(aliasVulnerabilities)

	for _, vuln := range IDs {
		if len(vuln.IDs) == 0 {
			return nil, Probe, errNoVulnID
		}
		f, err := finding.NewWith(fs, Probe,
			"Project contains OSV vulnerabilities", nil,
			finding.OutcomeTrue)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		f = f.WithMessage("Project is vulnerable to: " + strings.Join(vuln.IDs, " / "))
		f = f.WithRemediationMetadata(map[string]string{
			"osvid": vuln.IDs[0],
		})
		findings = append(findings, *f)
	}
	return findings, Probe, nil
}
