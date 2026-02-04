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

package tagsCannotDuplicateBranchNames

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
	probes.MustRegister(
		Probe,
		Run,
		[]checknames.CheckName{checknames.TagProtection},
	)
}

//go:embed *.yml
var fs embed.FS

const (
	Probe              = "tagsCannotDuplicateBranchNames"
	BranchNameKey      = "branchName"
	AccessLevelKey     = "minAccessLevel"
	PatternKey         = "matchingPattern"
	ProtectionLevelKey = "protectionLevel"

	// Access level values from GitLab API.
	AccessLevelNone       = 0  // No one
	AccessLevelDeveloper  = 30 // Developer
	AccessLevelMaintainer = 40 // Maintainer
	AccessLevelOwner      = 50 // Owner
)

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, Probe, fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	r := raw.TagProtectionResults

	// This probe is GitLab-specific
	if len(r.GitLabBranches) == 0 {
		// Not a GitLab repository or no branches
		f, err := finding.NewWith(
			fs,
			Probe,
			"not a GitLab repository or no branches found",
			nil,
			finding.OutcomeNotApplicable,
		)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		return []finding.Finding{*f}, Probe, nil
	}

	var findings []finding.Finding

	// Check each branch
	for _, branchName := range r.GitLabBranches {
		// Find matching patterns for this branch name
		matchingPatterns := findMatchingPatterns(branchName, r.GitLabProtectedTags)

		if len(matchingPatterns) == 0 {
			// No protection for this branch name
			f, err := finding.NewWith(
				fs,
				Probe,
				fmt.Sprintf("branch '%s' can have a tag created with the same name by anyone with write access", branchName),
				nil,
				finding.OutcomeFalse,
			)
			if err != nil {
				return nil, Probe, fmt.Errorf("create finding: %w", err)
			}
			f = f.WithValue(BranchNameKey, branchName)
			f = f.WithValue(ProtectionLevelKey, "none")
			findings = append(findings, *f)
			continue
		}

		// Get minimum access level from all matching patterns
		minAccessLevel := getMinimumAccessLevel(matchingPatterns)

		// Determine protection level
		var protectionLevel string
		var outcome finding.Outcome
		var text string

		switch {
		case minAccessLevel == AccessLevelNone:
			protectionLevel = "strongest"
			outcome = finding.OutcomeTrue
			text = fmt.Sprintf("branch '%s' is fully protected: no one can create a tag with this name", branchName)
		case minAccessLevel >= AccessLevelMaintainer:
			protectionLevel = "strong"
			outcome = finding.OutcomeTrue
			text = fmt.Sprintf("branch '%s' is protected: only maintainers+ can create a tag with this name", branchName)
		default:
			protectionLevel = "weak"
			outcome = finding.OutcomeFalse
			text = fmt.Sprintf("branch '%s' has weak protection: developers can create a tag with this name", branchName)
		}

		f, err := finding.NewWith(fs, Probe, text, nil, outcome)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		f = f.WithValue(BranchNameKey, branchName)
		f = f.WithValue(AccessLevelKey, fmt.Sprintf("%d", minAccessLevel))
		f = f.WithValue(ProtectionLevelKey, protectionLevel)
		findings = append(findings, *f)
	}

	return findings, Probe, nil
}

// findMatchingPatterns returns all protected tag patterns that match
// the given tag name.
func findMatchingPatterns(tagName string, patterns []checker.GitLabProtectedTagInfo) []checker.GitLabProtectedTagInfo {
	var matches []checker.GitLabProtectedTagInfo
	for _, pattern := range patterns {
		if pattern.Pattern == "" {
			continue
		}
		// Exact match
		if pattern.Pattern == tagName {
			matches = append(matches, pattern)
			continue
		}
		// Wildcard match
		if matchesWildcard(pattern.Pattern, tagName) {
			matches = append(matches, pattern)
		}
	}
	return matches
}

// matchesWildcard is a simple wildcard matcher for common patterns.
func matchesWildcard(pattern, name string) bool {
	// Handle the * wildcard (matches any string)
	if pattern == "*" {
		return true
	}
	// Handle prefix matching (pattern*)
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(name, prefix)
	}
	// Handle suffix matching (*pattern)
	if strings.HasPrefix(pattern, "*") {
		suffix := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(name, suffix)
	}
	return false
}

// getMinimumAccessLevel returns the minimum (most permissive) access level
// from a list of patterns.
func getMinimumAccessLevel(patterns []checker.GitLabProtectedTagInfo) int {
	if len(patterns) == 0 {
		return AccessLevelDeveloper // Default to developer if no patterns
	}

	minLevel := AccessLevelOwner + 1 // Start with a value higher than any real level
	for _, pattern := range patterns {
		if pattern.CreateAccessLevel < minLevel {
			minLevel = pattern.CreateAccessLevel
		}
	}

	return minLevel
}
