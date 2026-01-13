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

package hasOSVVulnerabilities

import (
	"embed"
	"errors"
	"fmt"
	"net/url"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/checknames"
	"github.com/ossf/scorecard/v5/internal/probes"
	"github.com/ossf/scorecard/v5/probes/internal/utils/uerror"
)

func init() {
	probes.MustRegister(Probe, Run, []checknames.CheckName{checknames.Vulnerabilities})
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

	dedupVulns := group(raw.VulnerabilitiesResults.Vulnerabilities)

	for _, vuln := range dedupVulns {
		if vuln.ID == "" {
			return nil, Probe, errNoVulnID
		}
		f, err := finding.NewWith(fs, Probe,
			"Project contains OSV vulnerabilities", nil,
			finding.OutcomeTrue)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		vulnLink, err := url.JoinPath("https://osv.dev", vuln.ID)
		if err != nil {
			return nil, Probe, fmt.Errorf("create osv link: %w", err)
		}
		f = f.WithMessage("Project is vulnerable to: " + vulnLink)
		f = f.WithRemediationMetadata(map[string]string{
			"osvid": vuln.ID,
		})
		findings = append(findings, *f)
	}
	return findings, Probe, nil
}
