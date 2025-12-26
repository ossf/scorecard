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

package blocksForcePushOnTags

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
	probes.MustRegister(
		Probe,
		Run,
		[]checknames.CheckName{checknames.TagProtection},
	)
}

//go:embed *.yml
var fs embed.FS

const (
	Probe      = "blocksForcePushOnTags"
	TagNameKey = "tagName"
)

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	r := raw.TagProtectionResults
	var findings []finding.Finding

	if len(r.Tags) == 0 {
		f, err := finding.NewWith(
			fs,
			Probe,
			"no release tags found",
			nil,
			finding.OutcomeNotApplicable,
		)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
		return findings, Probe, nil
	}

	for i := range r.Tags {
		tag := &r.Tags[i]

		protected := (tag.Protected != nil && *tag.Protected)
		if !protected {
			f, err := finding.NewWith(fs, Probe,
				fmt.Sprintf("tag '%s' is not protected", *tag.Name),
				nil, finding.OutcomeFalse)
			if err != nil {
				return nil, Probe, fmt.Errorf("create finding: %w", err)
			}
			f = f.WithValue(TagNameKey, *tag.Name)
			findings = append(findings, *f)
			continue
		}

		allowsForcePush := tag.TagProtectionRule.AllowForcePushes != nil &&
			*tag.TagProtectionRule.AllowForcePushes
		var text string
		var outcome finding.Outcome
		if !allowsForcePush {
			text = fmt.Sprintf("tag '%s' blocks force-pushes", *tag.Name)
			outcome = finding.OutcomeTrue
		} else {
			text = fmt.Sprintf("tag '%s' allows force-pushes", *tag.Name)
			outcome = finding.OutcomeFalse
		}

		f, err := finding.NewWith(fs, Probe, text, nil, outcome)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		f = f.WithValue(TagNameKey, *tag.Name)
		findings = append(findings, *f)
	}
	return findings, Probe, nil
}
