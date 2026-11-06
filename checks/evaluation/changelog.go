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

package evaluation

import (
	"fmt"
	"math"
	"strconv"

	"github.com/ossf/scorecard/v5/checker"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/hasChangelogFile"
	"github.com/ossf/scorecard/v5/probes/releasesHaveChangelog"
)

const (
	changelogFileScore    = 3
	changelogReleaseScore = 7
)

// Changelog applies the score policy for the Changelog check.
//
// Scoring:
//   - Path A (has changelog file): 3 pts for file + up to 7 pts proportional
//     based on how many of the last 5 releases have entries in the changelog.
//   - Path B (no file, has releases with notes): up to 10 pts proportional
//     based on how many of the last 5 releases have substantive body text.
//   - No file and no releases: inconclusive.
func Changelog(name string,
	findings []finding.Finding,
	dl checker.DetailLogger,
) checker.CheckResult {
	expectedProbes := []string{
		hasChangelogFile.Probe,
		releasesHaveChangelog.Probe,
	}

	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	score := 0
	hasFile := false
	hasReleases := false
	var logLevel checker.DetailType
	for i := range findings {
		f := &findings[i]
		switch f.Outcome {
		case finding.OutcomeTrue:
			logLevel = checker.DetailInfo
			switch f.Probe {
			case hasChangelogFile.Probe:
				if !hasFile {
					hasFile = true
					score += changelogFileScore
				}
			case releasesHaveChangelog.Probe:
				hasReleases = true
				score += releaseScore(f, hasFile)
			}
		case finding.OutcomeFalse:
			logLevel = checker.DetailWarn
			if f.Probe == releasesHaveChangelog.Probe {
				hasReleases = true
				score += releaseScore(f, hasFile)
			}
		case finding.OutcomeNotApplicable:
			logLevel = checker.DetailDebug
		default:
			continue
		}
		checker.LogFinding(dl, f, logLevel)
	}

	if !hasFile && !hasReleases {
		return checker.CreateInconclusiveResult(name, "no changelog file or releases found")
	}

	if hasFile {
		return checker.CreateResultWithScore(name, "changelog file detected", score)
	}

	return checker.CreateResultWithScore(name,
		fmt.Sprintf("%d out of %d release(s) have descriptive logs",
			releaseCount(findings), releaseTotal(findings)),
		score)
}

// releaseScore returns the proportional score for the release probe.
// When a changelog file exists, releases are worth 7 points.
// When no file exists, releases are worth the full 10 points.
func releaseScore(f *finding.Finding, hasFile bool) int {
	maxScore := changelogReleaseScore
	if !hasFile {
		maxScore = checker.MaxResultScore
	}
	withChangelog, err := strconv.Atoi(f.Values[releasesHaveChangelog.ReleasesWithChangelogKey])
	if err != nil {
		return 0
	}
	total, err := strconv.Atoi(f.Values[releasesHaveChangelog.ReleasesTotalKey])
	if err != nil || total == 0 {
		return 0
	}
	return int(math.Floor(float64(maxScore) * float64(withChangelog) / float64(total)))
}

func releaseCount(findings []finding.Finding) int {
	for i := range findings {
		if findings[i].Probe == releasesHaveChangelog.Probe {
			n, err := strconv.Atoi(findings[i].Values[releasesHaveChangelog.ReleasesWithChangelogKey])
			if err != nil {
				return 0
			}
			return n
		}
	}
	return 0
}

func releaseTotal(findings []finding.Finding) int {
	for i := range findings {
		if findings[i].Probe == releasesHaveChangelog.Probe {
			n, err := strconv.Atoi(findings[i].Values[releasesHaveChangelog.ReleasesTotalKey])
			if err != nil {
				return 0
			}
			return n
		}
	}
	return 0
}
