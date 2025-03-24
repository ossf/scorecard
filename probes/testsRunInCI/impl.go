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
package testsRunInCI

import (
	"embed"
	"fmt"
	"strings"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/checknames"
	"github.com/ossf/scorecard/v5/internal/probes"
	"github.com/ossf/scorecard/v5/probes/internal/utils/uerror"
)

func init() {
	probes.MustRegister(Probe, Run, []checknames.CheckName{checknames.CITests})
}

//go:embed *.yml
var fs embed.FS

const (
	Probe   = "testsRunInCI"
	success = "success"
)

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	var findings []finding.Finding

	c := raw.CITestResults

	if len(c.CIInfo) == 0 {
		f, err := finding.NewWith(fs, Probe,
			"no pull requests found", nil,
			finding.OutcomeNotApplicable)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
		return findings, Probe, nil
	}

	for i := range c.CIInfo {
		r := c.CIInfo[i]
		// GitHub Statuses.
		prSuccessStatus, f, err := prHasSuccessStatus(r)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		if prSuccessStatus {
			findings = append(findings, *f)
			continue
		}

		// GitHub Check Runs.
		prCheckSuccessful, f, err := prHasSuccessfulCheck(r)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		if prCheckSuccessful {
			findings = append(findings, *f)
		}

		if !prSuccessStatus && !prCheckSuccessful {
			f, err := finding.NewWith(fs, Probe,
				fmt.Sprintf("merged PR %d without CI test at HEAD: %s", r.PullRequestNumber, r.HeadSHA),
				nil, finding.OutcomeFalse)
			if err != nil {
				return nil, Probe, fmt.Errorf("create finding: %w", err)
			}
			findings = append(findings, *f)
		}
	}

	return findings, Probe, nil
}

// PR has a status marked 'success' and a CI-related context.
//
//nolint:unparam
func prHasSuccessStatus(r checker.RevisionCIInfo) (bool, *finding.Finding, error) {
	for _, status := range r.Statuses {
		if status.State != success {
			continue
		}
		if isTest(status.Context) || isTest(status.TargetURL) {
			msg := fmt.Sprintf("CI test found: pr: %s, context: %s", r.HeadSHA,
				status.Context)

			f, err := finding.NewWith(fs, Probe,
				msg, nil,
				finding.OutcomeTrue)
			if err != nil {
				return false, nil, fmt.Errorf("create finding: %w", err)
			}

			loc := &finding.Location{
				Path: status.URL,
				Type: finding.FileTypeURL,
			}
			f = f.WithLocation(loc)
			return true, f, nil
		}
	}
	return false, nil, nil
}

// PR has a successful CI-related check.
//
//nolint:unparam
func prHasSuccessfulCheck(r checker.RevisionCIInfo) (bool, *finding.Finding, error) {
	for _, cr := range r.CheckRuns {
		if cr.Status != "completed" {
			continue
		}
		if cr.Conclusion != success {
			continue
		}
		if isTest(cr.App.Slug) {
			msg := fmt.Sprintf("CI test found: pr: %d, context: %s", r.PullRequestNumber,
				cr.App.Slug)

			f, err := finding.NewWith(fs, Probe,
				msg, nil,
				finding.OutcomeTrue)
			if err != nil {
				return false, nil, fmt.Errorf("create finding: %w", err)
			}

			loc := &finding.Location{
				Path: cr.URL,
				Type: finding.FileTypeURL,
			}
			f = f.WithLocation(loc)
			return true, f, nil
		}
	}
	return false, nil, nil
}

// isTest returns true if the given string is a CI test.
func isTest(s string) bool {
	l := strings.ToLower(s)

	// Add more patterns here!
	for _, pattern := range []string{
		"appveyor", "buildkite", "circleci", "e2e", "github-actions", "jenkins",
		"mergeable", "packit-as-a-service", "semaphoreci", "test", "travis-ci",
		"flutter-dashboard", "cirrus-ci", "Cirrus CI", "azure-pipelines", "ci/woodpecker",
		"vstfs:///build/build",
	} {
		if strings.Contains(l, pattern) {
			return true
		}
	}
	return false
}
