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
package pinsDependencies

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/fileparser"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/internal/probes"
	"github.com/ossf/scorecard/v4/probes/internal/utils/uerror"
)

func init() {
	probes.MustRegister(Probe, Run, []probes.CheckName{probes.PinnedDependencies})
}

//go:embed *.yml
var fs embed.FS

const (
	Probe      = "pinsDependencies"
	DepTypeKey = "dependencyType"
)

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	var findings []finding.Finding

	r := raw.PinningDependenciesResults

	for i := range r.ProcessingErrors {
		e := r.ProcessingErrors[i]
		f, err := finding.NewWith(fs, Probe, generateTextIncompleteResults(e),
			&e.Location, finding.OutcomeError)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	}

	for i := range r.Dependencies {
		rr := r.Dependencies[i]
		loc := rr.Location.Location()
		f, err := finding.NewWith(fs, Probe, "", loc, finding.OutcomeNotSupported)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		if rr.Location == nil {
			if rr.Msg == nil {
				e := sce.WithMessage(sce.ErrScorecardInternal, "empty File field")
				return findings, Probe, e
			}
			f = f.WithMessage(*rr.Msg).WithOutcome(finding.OutcomeNotSupported)
			findings = append(findings, *f)
			continue
		}
		if rr.Msg != nil {
			f = f.WithMessage(*rr.Msg).WithOutcome(finding.OutcomeNotSupported)
			findings = append(findings, *f)
			continue
		}
		if rr.Pinned == nil {
			f = f.WithMessage(fmt.Sprintf("%s has empty Pinned field", rr.Type)).
				WithOutcome(finding.OutcomeNotSupported)
			findings = append(findings, *f)
			continue
		}
		if !*rr.Pinned {
			f = f.WithMessage(generateTextUnpinned(&rr)).
				WithOutcome(finding.OutcomeFalse)
			if rr.Remediation != nil {
				f.Remediation = rr.Remediation
			}
			f = f.WithValues(map[string]string{
				DepTypeKey: string(rr.Type),
			})
			findings = append(findings, *f)
		} else {
			f = f.WithMessage("").WithOutcome(finding.OutcomeTrue)
			f = f.WithValues(map[string]string{
				DepTypeKey: string(rr.Type),
			})
			findings = append(findings, *f)
		}
	}

	if len(findings) == 0 {
		f, err := finding.NewWith(fs, Probe, "no dependencies found", nil, finding.OutcomeNotApplicable)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		return []finding.Finding{*f}, Probe, nil
	}

	return findings, Probe, nil
}

func generateTextIncompleteResults(e checker.ElementError) string {
	return fmt.Sprintf("Possibly incomplete results: %s", e.Err)
}

func generateTextUnpinned(rr *checker.Dependency) string {
	if rr.Type == checker.DependencyUseTypeGHAction {
		// Check if we are dealing with a GitHub action or a third-party one.
		gitHubOwned := fileparser.IsGitHubOwnedAction(rr.Location.Snippet)
		owner := generateOwnerToDisplay(gitHubOwned)
		return fmt.Sprintf("%s not pinned by hash", owner)
	}

	return fmt.Sprintf("%s not pinned by hash", rr.Type)
}

func generateOwnerToDisplay(gitHubOwned bool) string {
	if gitHubOwned {
		return fmt.Sprintf("GitHub-owned %s", checker.DependencyUseTypeGHAction)
	}
	return fmt.Sprintf("third-party %s", checker.DependencyUseTypeGHAction)
}
