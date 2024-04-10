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

func CreateFalseFinding(r checker.TokenPermission,
	probe string,
	fs embed.FS,
	metadata map[string]string,
) (*finding.Finding, error) {
	// Create finding
	text, err := createText(r)
	if err != nil {
		return nil, fmt.Errorf("create finding: %w", err)
	}
	f, err := finding.NewWith(fs, probe,
		text, nil, finding.OutcomeFalse)
	if err != nil {
		return nil, fmt.Errorf("create finding: %w", err)
	}

	if r.File != nil {
		f = f.WithLocation(r.File.Location())
		workflowPath := strings.TrimPrefix(f.Location.Path, ".github/workflows/")
		f = f.WithRemediationMetadata(map[string]string{"workflow": workflowPath})
	}
	if metadata != nil {
		f = f.WithRemediationMetadata(metadata)
	}

	if r.Name != nil {
		f = f.WithValue("tokenName", *r.Name)
	}
	f = f.WithValue("permissionLevel", string(r.Type))
	return f, nil
}

func ReadTrueLevelFinding(probe string,
	fs embed.FS,
	r checker.TokenPermission,
	metadata map[string]string,
) (*finding.Finding, error) {
	f, err := finding.NewWith(fs, probe,
		"found token with 'read' permissions",
		nil, finding.OutcomeTrue)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	if r.File != nil {
		f = f.WithLocation(r.File.Location())
		workflowPath := strings.TrimPrefix(f.Location.Path, ".github/workflows/")
		f = f.WithRemediationMetadata(map[string]string{"workflow": workflowPath})
	}
	if metadata != nil {
		f = f.WithRemediationMetadata(metadata)
	}

	f = f.WithValue("permissionLevel", "read")
	return f, nil
}

func CreateNoneFinding(probe string,
	fs embed.FS,
	r checker.TokenPermission,
	metadata map[string]string,
) (*finding.Finding, error) {
	// Create finding
	f, err := finding.NewWith(fs, probe,
		"found token with 'none' permissions",
		nil, finding.OutcomeFalse)
	if err != nil {
		return nil, fmt.Errorf("create finding: %w", err)
	}
	if r.File != nil {
		f = f.WithLocation(r.File.Location())
		workflowPath := strings.TrimPrefix(f.Location.Path, ".github/workflows/")
		f = f.WithRemediationMetadata(map[string]string{"workflow": workflowPath})
	}
	if metadata != nil {
		f = f.WithRemediationMetadata(metadata)
	}

	f = f.WithValue("permissionLevel", string(r.Type))
	return f, nil
}

func CreateUndeclaredFinding(probe string,
	fs embed.FS,
	r checker.TokenPermission,
	metadata map[string]string,
) (*finding.Finding, error) {
	var f *finding.Finding
	var err error
	switch {
	case r.LocationType == nil:
		f, err = finding.NewWith(fs, probe,
			"could not determine the location type",
			nil, finding.OutcomeNotApplicable)
		if err != nil {
			return nil, fmt.Errorf("create finding: %w", err)
		}
	case *r.LocationType == checker.PermissionLocationTop,
		*r.LocationType == checker.PermissionLocationJob:
		// Create finding
		f, err = CreateFalseFinding(r, probe, fs, metadata)
		if err != nil {
			return nil, fmt.Errorf("create finding: %w", err)
		}
	default:
		f, err = finding.NewWith(fs, probe,
			"could not determine the location type",
			nil, finding.OutcomeError)
		if err != nil {
			return nil, fmt.Errorf("create finding: %w", err)
		}
	}
	f = f.WithValue("permissionLevel", string(r.Type))
	return f, nil
}
