// Copyright 2025 OpenSSF Scorecard Authors
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

package checks

import (
	"fmt"
	"os"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/checks/raw"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes"
	"github.com/ossf/scorecard/v5/probes/zrunner"
)

// CheckReleasesDirectDepsVulnFree is the registered name for this check.
const CheckReleasesDirectDepsVulnFree = "ReleasesDirectDepsVulnFree"

func releasesDepsDebug() bool {
	switch v := os.Getenv("RELEASES_DEPS_DEBUG"); v {
	case "1", "true", "TRUE", "True", "yes", "on", "ON":
		return true
	default:
		return false
	}
}

//nolint:gochecknoinits
func init() {
	// Register the check function. A panic here means programming error.
	if err := registerCheck(CheckReleasesDirectDepsVulnFree, ReleasesDirectDepsVulnFree, nil); err != nil {
		panic(err)
	}
}

// ReleasesDirectDepsVulnFree runs the end-to-end flow for this check.
func ReleasesDirectDepsVulnFree(c *checker.CheckRequest) checker.CheckResult {
	// 1) Collect raw data for last 10 releases using "known-at-release-time" policy by default.
	//    This only flags vulnerabilities that were published before/at the release date.
	rawData, err := raw.ReleasesDirectDepsVulnFree(c)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckReleasesDirectDepsVulnFree, e)
	}

	if releasesDepsDebug() {
		fmt.Fprintf(os.Stderr, "[releases-deps] raw collection finished: releases=%d\n", len(rawData.Releases))
	}

	// 2) Attach raw results for probes to consume.
	pRawResults := getRawResults(c)
	pRawResults.ReleaseDirectDepsVulnsResults = *rawData

	// 3) Run probe(s) registered for this check’s group.
	findings, err := zrunner.Run(pRawResults, probes.ReleasesDirectDepsAreVulnFree)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckReleasesDirectDepsVulnFree, e)
	}

	// 4) Evaluate: proportional score across releases.
	total := len(findings)
	if total == 0 {
		// No releases present -> inconclusive is more informative than 0/10.
		reason := "no releases found to evaluate"
		return checker.CreateInconclusiveResult(CheckReleasesDirectDepsVulnFree, reason)
	}

	clean := 0
	for i := range findings {
		if findings[i].Outcome == finding.OutcomeTrue {
			clean++
		}
	}

	// Debug logging: show vulnerability details for each release
	if releasesDepsDebug() {
		fmt.Fprintf(os.Stderr, "\n[releases-deps] Debug output enabled - showing details for %d releases:\n",
			len(rawData.Releases))
		for _, rel := range rawData.Releases {
			if len(rel.Findings) == 0 {
				fmt.Fprintf(os.Stderr, "[releases-deps] release %s (published: %s): CLEAN (no vulnerabilities)\n",
					rel.Tag, rel.PublishedAt.Format("2006-01-02"))
			} else {
				fmt.Fprintf(os.Stderr,
					"[releases-deps] release %s (published: %s): VULNERABLE (%d dependencies with issues)\n",
					rel.Tag, rel.PublishedAt.Format("2006-01-02"), len(rel.Findings))
				for _, vuln := range rel.Findings {
					fmt.Fprintf(os.Stderr, "[releases-deps]   - %s@%s (%s) has %d vulnerabilities: %v [manifest: %s]\n",
						vuln.Name, vuln.Version, vuln.Ecosystem, len(vuln.OSVIDs), vuln.OSVIDs, vuln.ManifestPath)
				}
			}
		}
		fmt.Fprintf(os.Stderr, "[releases-deps] Summary: clean=%d total=%d\n\n", clean, total)
	}

	reason := fmt.Sprintf(
		"%d/%d recent releases were free of known vulnerable direct dependencies at the time of release",
		clean, total)

	if releasesDepsDebug() {
		fmt.Fprintf(os.Stderr, "[releases-deps] evaluation: clean=%d total=%d => %s\n", clean, total, reason)
	}

	// Standard helper to compute proportional score.
	res := checker.CreateProportionalScoreResult(CheckReleasesDirectDepsVulnFree, reason, clean, total)
	res.Findings = findings
	return res
}
