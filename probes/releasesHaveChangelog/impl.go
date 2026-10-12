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

package releasesHaveChangelog

import (
	"embed"
	"fmt"
	"strconv"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/checknames"
	"github.com/ossf/scorecard/v5/internal/probes"
	"github.com/ossf/scorecard/v5/probes/internal/utils/uerror"
)

func init() {
	probes.MustRegister(Probe, Run, []checknames.CheckName{checknames.Changelog})
}

//go:embed *.yml
var fs embed.FS

const (
	Probe                    = "releasesHaveChangelog"
	ReleasesWithChangelogKey = "releasesWithChangelog"
	ReleasesTotalKey         = "releasesTotal"
)

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	var findings []finding.Finding

	total := raw.ChangelogResults.TotalReleases
	withChangelog := raw.ChangelogResults.ReleasesWithChangelog

	if total == 0 {
		f, err := finding.NewNotApplicable(fs, Probe, "no releases found", nil)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
		return findings, Probe, nil
	}

	// Emit a single finding with the counts for the evaluation to use.
	var f *finding.Finding
	var err error
	if withChangelog == total {
		f, err = finding.NewTrue(fs, Probe,
			fmt.Sprintf("%d out of %d release(s) have a changelog entry", withChangelog, total), nil)
	} else {
		f, err = finding.NewFalse(fs, Probe,
			fmt.Sprintf("%d out of %d release(s) have a changelog entry", withChangelog, total), nil)
	}
	if err != nil {
		return nil, Probe, fmt.Errorf("create finding: %w", err)
	}
	f.WithValue(ReleasesWithChangelogKey, strconv.Itoa(withChangelog))
	f.WithValue(ReleasesTotalKey, strconv.Itoa(total))
	findings = append(findings, *f)

	return findings, Probe, nil
}
