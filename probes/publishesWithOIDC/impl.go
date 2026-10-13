// Copyright 2026 OpenSSF Scorecard Authors
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

package publishesWithOIDC

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
	probes.MustRegister(Probe, Run, []checknames.CheckName{checknames.Packaging})
}

//go:embed *.yml
var fs embed.FS

const Probe = "publishesWithOIDC"

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	var findings []finding.Finding
	for _, pkg := range raw.PackagingResults.Packages {
		if pkg.Msg != nil {
			continue
		}
		var outcome finding.Outcome
		var message string
		switch pkg.PublishAuth {
		case checker.PublishAuthOIDC:
			outcome = finding.OutcomeTrue
			message = fmt.Sprintf("%s publishing uses OpenID Connect", pkg.PublishEcosystem)
		case checker.PublishAuthCredential:
			outcome = finding.OutcomeFalse
			message = fmt.Sprintf("%s publishing uses a long-lived credential", pkg.PublishEcosystem)
		default:
			continue
		}
		f, err := finding.NewWith(fs, Probe, message, nil, outcome)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		loc := &finding.Location{}
		if pkg.File != nil {
			loc.Path = pkg.File.Path
			loc.Type = pkg.File.Type
			loc.LineStart = &pkg.File.Offset
		}
		f = f.WithLocation(loc)
		findings = append(findings, *f)
	}
	if len(findings) > 0 {
		return findings, Probe, nil
	}

	f, err := finding.NewWith(fs, Probe,
		"no trusted-publishing workflow detected", nil, finding.OutcomeNotApplicable)
	if err != nil {
		return nil, Probe, fmt.Errorf("create finding: %w", err)
	}
	return []finding.Finding{*f}, Probe, nil
}
