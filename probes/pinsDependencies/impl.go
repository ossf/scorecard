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
	"github.com/ossf/scorecard/v4/finding/probe"
	"github.com/ossf/scorecard/v4/probes/internal/utils/uerror"
	"github.com/ossf/scorecard/v4/rule"
)

//go:embed *.yml
var fs embed.FS

const (
	Probe      = "pinsDependencies"
	DepTypeKey = "dependencyType"
)

var dependencyTypes = map[checker.DependencyUseType]int{
	checker.DependencyUseTypeGHAction:                 0,
	checker.DependencyUseTypeDockerfileContainerImage: 1,
	checker.DependencyUseTypeDownloadThenRun:          2,
	checker.DependencyUseTypeGoCommand:                3,
	checker.DependencyUseTypeChocoCommand:             4,
	checker.DependencyUseTypeNpmCommand:               5,
	checker.DependencyUseTypePipCommand:               6,
	checker.DependencyUseTypeNugetCommand:             7,
}

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	var findings []finding.Finding

	r := raw.PinningDependenciesResults

	for i := range r.ProcessingErrors {
		e := r.ProcessingErrors[i]
		f := finding.Finding{
			Message:  generateTextIncompleteResults(e),
			Location: &e.Location,
			Outcome:  finding.OutcomeError,
		}
		findings = append(findings, f)
	}

	for i := range r.Dependencies {
		rr := r.Dependencies[i]
		if rr.Location == nil {
			if rr.Msg == nil {
				e := sce.WithMessage(sce.ErrScorecardInternal, "empty File field")
				return findings, Probe, e
			}
			f := &finding.Finding{
				Probe:   Probe,
				Outcome: finding.OutcomeNotApplicable,
				Message: *rr.Msg,
			}
			findings = append(findings, *f)
			continue
		}
		if rr.Msg != nil {
			loc := &finding.Location{
				Type:      rr.Location.Type,
				Path:      rr.Location.Path,
				LineStart: &rr.Location.Offset,
				LineEnd:   &rr.Location.EndOffset,
				Snippet:   &rr.Location.Snippet,
			}
			f := &finding.Finding{
				Probe:    Probe,
				Outcome:  finding.OutcomeNotApplicable,
				Message:  *rr.Msg,
				Location: loc,
			}
			findings = append(findings, *f)
			continue
		}
		if rr.Pinned == nil {
			loc := &finding.Location{
				Type:      rr.Location.Type,
				Path:      rr.Location.Path,
				LineStart: &rr.Location.Offset,
				LineEnd:   &rr.Location.EndOffset,
				Snippet:   &rr.Location.Snippet,
			}
			f := &finding.Finding{
				Probe:    Probe,
				Outcome:  finding.OutcomeNotApplicable,
				Message:  fmt.Sprintf("%s has empty Pinned field", rr.Type),
				Location: loc,
			}
			findings = append(findings, *f)
			continue
		}
		if !*rr.Pinned {
			loc := &finding.Location{
				Type:      rr.Location.Type,
				Path:      rr.Location.Path,
				LineStart: &rr.Location.Offset,
				LineEnd:   &rr.Location.EndOffset,
				Snippet:   &rr.Location.Snippet,
			}
			f := &finding.Finding{
				Probe:    Probe,
				Outcome:  finding.OutcomeNegative,
				Message:  generateTextUnpinned(&rr),
				Location: loc,
			}
			if rr.Remediation != nil {
				f.Remediation = ruleRemToProbeRem(rr.Remediation)
			}
			f = f.WithValues(map[string]int{
				DepTypeKey: dependencyTypes[rr.Type],
			})
			findings = append(findings, *f)
		} else {
			loc := &finding.Location{
				Type:      rr.Location.Type,
				Path:      rr.Location.Path,
				LineStart: &rr.Location.Offset,
				LineEnd:   &rr.Location.EndOffset,
				Snippet:   &rr.Location.Snippet,
			}
			f := &finding.Finding{
				Probe:    Probe,
				Outcome:  finding.OutcomePositive,
				Location: loc,
			}
			f = f.WithValues(map[string]int{
				DepTypeKey: dependencyTypes[rr.Type],
			})
			findings = append(findings, *f)
		}
	}

	if len(findings) == 0 {
		f, err := finding.NewWith(fs, Probe,
			"no dependencies found", nil,
			finding.OutcomeNotAvailable)
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

func ruleRemToProbeRem(rem *rule.Remediation) *probe.Remediation {
	return &probe.Remediation{
		Patch:    rem.Patch,
		Text:     rem.Text,
		Markdown: rem.Markdown,
		Effort:   probe.RemediationEffort(rem.Effort),
	}
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
