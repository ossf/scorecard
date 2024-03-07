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

package permissions

import (
	"embed"
	"fmt"
	"strings"

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
)

func createText(t checker.TokenPermission) (string, error) {
	// By default, use the message already present.
	if t.Msg != nil {
		return *t.Msg, nil
	}

	// Ensure there's no implementation bug.
	if t.LocationType == nil {
		return "", sce.WithMessage(sce.ErrScorecardInternal, "locationType is nil")
	}

	// Use a different text depending on the type.
	if t.Type == checker.PermissionLevelUndeclared {
		return fmt.Sprintf("no %s permission defined", *t.LocationType), nil
	}

	if t.Value == nil {
		return "", sce.WithMessage(sce.ErrScorecardInternal, "Value fields is nil")
	}

	if t.Name == nil {
		return fmt.Sprintf("%s permissions set to '%v'", *t.LocationType,
			*t.Value), nil
	}

	return fmt.Sprintf("%s '%v' permission set to '%v'", *t.LocationType,
		*t.Name, *t.Value), nil
}

func CreateNegativeFinding(r checker.TokenPermission,
	probe string,
	fs embed.FS,
) (*finding.Finding, error) {
	// Create finding
	text, err := createText(r)
	if err != nil {
		return nil, fmt.Errorf("create finding: %w", err)
	}
	f, err := finding.NewWith(fs, probe,
		text, nil, finding.OutcomeNegative)
	if err != nil {
		return nil, fmt.Errorf("create finding: %w", err)
	}

	// Create Location
	var loc *finding.Location
	if r.File != nil {
		loc = &finding.Location{
			Type:      r.File.Type,
			Path:      r.File.Path,
			LineStart: newUint(r.File.Offset),
		}
		if r.File.Snippet != "" {
			loc.Snippet = newStr(r.File.Snippet)
		}
		f = f.WithLocation(loc)
		f = f.WithRemediationMetadata(map[string]string{
			"repo":     r.Remediation.Repo,
			"branch":   r.Remediation.Branch,
			"workflow": strings.TrimPrefix(f.Location.Path, ".github/workflows/"),
		})
	}
	if r.LocationType != nil {
		f = f.WithValue("permissionLocation", string(*r.LocationType))
	}
	if r.Name != nil {
		f = f.WithValue("tokenName", *r.Name)
	}
	f = f.WithValue("permissionLevel", string(r.Type))
	return f, nil
}

// avoid memory aliasing by returning a new copy.
func newUint(u uint) *uint {
	return &u
}

// avoid memory aliasing by returning a new copy.
func newStr(s string) *string {
	return &s
}

func ReadPositiveLevelFinding(probe string, fs embed.FS, r checker.TokenPermission) (*finding.Finding, error) {
	f, err := finding.NewWith(fs, probe,
		"no workflows with 'read' permissions",
		nil, finding.OutcomePositive)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	var loc *finding.Location
	if r.File != nil {
		loc = &finding.Location{
			Type:      r.File.Type,
			Path:      r.File.Path,
			LineStart: newUint(r.File.Offset),
		}
		if r.File.Snippet != "" {
			loc.Snippet = newStr(r.File.Snippet)
		}
		f = f.WithLocation(loc)
		f = f.WithRemediationMetadata(map[string]string{
			"repo":     r.Remediation.Repo,
			"branch":   r.Remediation.Branch,
			"workflow": strings.TrimPrefix(f.Location.Path, ".github/workflows/"),
		})
	}
	f = f.WithValue("permissionLevel", "read")
	return f, nil
}

func CreateNoneFinding(probe string, fs embed.FS, r checker.TokenPermission) (*finding.Finding, error) {
	// Create finding
	f, err := finding.NewWith(fs, probe,
		"no workflows with 'none' permissions",
		nil, finding.OutcomeNegative)
	if err != nil {
		return nil, fmt.Errorf("create finding: %w", err)
	}
	var loc *finding.Location
	if r.File != nil {
		loc = &finding.Location{
			Type:      r.File.Type,
			Path:      r.File.Path,
			LineStart: newUint(r.File.Offset),
		}
		if r.File.Snippet != "" {
			loc.Snippet = newStr(r.File.Snippet)
		}
		f = f.WithLocation(loc)
		f = f.WithRemediationMetadata(map[string]string{
			"repo":     r.Remediation.Repo,
			"branch":   r.Remediation.Branch,
			"workflow": strings.TrimPrefix(f.Location.Path, ".github/workflows/"),
		})
	}
	f = f.WithValue("permissionLevel", string(r.Type))
	return f, nil
}
