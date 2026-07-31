// Package scorecardref parses and validates repository references of the form
// platform/owner/repo.
package scorecardref

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/ossf/scorecard/v5/cmd/internal/mcp/model"
)

// Supported platforms.
const (
	PlatformGitHub = "github.com"
	PlatformGitLab = "gitlab.com"
)

// Sentinel errors returned by this package.
var (
	errInvalidRef          = errors.New("invalid repository reference")
	errUnsupportedPlatform = errors.New("unsupported platform")
	errInvalidCommit       = errors.New("invalid commit")

	supportedPlatforms = map[string]bool{PlatformGitHub: true, PlatformGitLab: true}
	commitSHA          = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)
)

// Parse parses a "platform/owner/repo" reference. The platform is optional and
// defaults to github.com; gitlab.com is also supported. A leading scheme and a
// trailing slash are tolerated. Unsupported platforms and malformed references
// return an error.
func Parse(s string) (model.RepoRef, error) {
	orig := s
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	s = strings.Trim(s, "/")
	if s == "" {
		return model.RepoRef{}, fmt.Errorf("%w: expected platform/owner/repo", errInvalidRef)
	}

	parts := strings.Split(s, "/")
	var ref model.RepoRef
	switch len(parts) {
	case 2:
		ref = model.RepoRef{Platform: PlatformGitHub, Org: parts[0], Name: parts[1]}
	case 3:
		ref = model.RepoRef{Platform: strings.ToLower(parts[0]), Org: parts[1], Name: parts[2]}
	default:
		return model.RepoRef{}, fmt.Errorf(
			"%w %q: expected platform/owner/repo (platform optional, defaults to %s)",
			errInvalidRef, orig, PlatformGitHub)
	}

	if ref.Org == "" || ref.Name == "" {
		return model.RepoRef{}, fmt.Errorf("%w %q: owner and repository must be non-empty", errInvalidRef, orig)
	}
	if !supportedPlatforms[ref.Platform] {
		return model.RepoRef{}, fmt.Errorf(
			"%w %q: supported platforms are %s and %s",
			errUnsupportedPlatform, ref.Platform, PlatformGitHub, PlatformGitLab)
	}
	return ref, nil
}

// ValidateCommit returns an error if commit is non-empty and is not a
// 40-character hexadecimal SHA.
func ValidateCommit(commit string) error {
	if commit == "" {
		return nil
	}
	if !commitSHA.MatchString(commit) {
		return fmt.Errorf("%w %q: expected a 40-character hexadecimal SHA", errInvalidCommit, commit)
	}
	return nil
}
