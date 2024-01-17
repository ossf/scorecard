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
package hasGitHubWorkflowPermissionUndeclared

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/internal/utils/permissions"
	"github.com/ossf/scorecard/v4/probes/internal/utils/uerror"
)

//go:embed *.yml
var fs embed.FS

const Probe = "hasGitHubWorkflowPermissionUndeclared"

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	results := raw.TokenPermissionsResults
	var findings []finding.Finding

	if results.NumTokens == 0 {
		f, err := finding.NewWith(fs, Probe,
			"No token permissions found",
			nil, finding.OutcomeNotAvailable)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
		return findings, Probe, nil
	}

	for _, r := range results.TokenPermissions {
		if r.Type != checker.PermissionLevelUndeclared {
			continue
		}
		topLevel := 0
		jobLevel := 0
		if *r.LocationType == checker.PermissionLocationTop {
			topLevel = 1
		}
		if *r.LocationType == checker.PermissionLocationJob {
			jobLevel = 1
		}
		switch {
		case r.LocationType == nil:
			f, err := finding.NewWith(fs, Probe,
				"could not determine the location type",
				nil, finding.OutcomeNotAvailable)
			if err != nil {
				return nil, Probe, fmt.Errorf("create finding: %w", err)
			}
			findings = append(findings, *f)
		case *r.LocationType == checker.PermissionLocationTop,
			*r.LocationType == checker.PermissionLocationJob:
			// Create finding
			f, err := permissions.CreateNegativeFinding(r, Probe, fs)
			if err != nil {
				return nil, Probe, fmt.Errorf("create finding: %w", err)
			}
			f = f.WithValues(map[string]int{
				"topLevel": topLevel,
				"jobLevel": jobLevel,
			})
			findings = append(findings, *f)
		default:
			f, err := finding.NewWith(fs, Probe,
				"could not determine the location type",
				nil, finding.OutcomeNotApplicable)
			if err != nil {
				return nil, Probe, fmt.Errorf("create finding: %w", err)
			}
			findings = append(findings, *f)
		}
	}

	if len(findings) == 0 {
		f, err := finding.NewWith(fs, Probe,
			"project has no workflows with undeclared permissions",
			nil, finding.OutcomePositive)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	}
	return findings, Probe, nil
}
