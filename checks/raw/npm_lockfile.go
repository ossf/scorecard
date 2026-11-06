// Copyright 2026 OpenSSF Scorecard Authors
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

package raw

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gobwas/glob"
)

var (
	errNpmLockfileMissingPackages     = errors.New("npm lockfile has no packages section")
	errNpmLockfileUnsupportedVersion  = errors.New("unsupported npm lockfile version")
	errNpmWorkspacePatternUnsupported = errors.New("unsupported npm workspace pattern")
)

type npmLockVerification struct {
	Invalid    []string
	Packages   []npmLockPackageResult
	Applicable int
}

type npmLockPackageResult struct {
	Path   string
	Pinned bool
}

type npmLockFile struct {
	Packages        map[string]npmPackage    `json:"packages"`
	Dependencies    map[string]npmDependency `json:"dependencies"`
	LockfileVersion int                      `json:"lockfileVersion"`
}

type npmPackage struct {
	Integrity string `json:"integrity"`
	Resolved  string `json:"resolved"`
	Link      bool   `json:"link"`
	InBundle  bool   `json:"inBundle"`
}

type npmDependency struct {
	Dependencies map[string]npmDependency `json:"dependencies"`
	Integrity    string                   `json:"integrity"`
	Version      string                   `json:"version"`
	Bundled      bool                     `json:"bundled"`
}

type pendingNpmDependency struct {
	name string
	dep  npmDependency
}

func verifyNpmLockFile(content []byte) (npmLockVerification, error) {
	var lock npmLockFile
	if err := json.Unmarshal(content, &lock); err != nil {
		return npmLockVerification{}, fmt.Errorf(
			"parse npm lockfile: %w",
			err,
		)
	}

	switch lock.LockfileVersion {
	case 1:
		return verifyNpmLockFileV1(lock.Dependencies), nil
	case 2, 3:
		if lock.Packages == nil {
			return npmLockVerification{}, fmt.Errorf(
				"%w: v%d",
				errNpmLockfileMissingPackages,
				lock.LockfileVersion,
			)
		}

		return verifyNpmLockFilePackages(lock.Packages), nil
	default:
		return npmLockVerification{}, fmt.Errorf(
			"%w: %d",
			errNpmLockfileUnsupportedVersion,
			lock.LockfileVersion,
		)
	}
}

func verifyNpmLockFilePackages(packages map[string]npmPackage) npmLockVerification {
	result := npmLockVerification{}

	linkedTargets := make(map[string]bool)
	for _, pkg := range packages {
		if pkg.Link && pkg.Resolved != "" {
			target := path.Clean(pkg.Resolved)
			if target != "." &&
				!path.IsAbs(target) &&
				target != ".." &&
				!strings.HasPrefix(target, "../") &&
				!strings.Contains("/"+target+"/", "/node_modules/") {
				linkedTargets[target] = true
			}
		}
	}

	for packagePath, pkg := range packages {
		if packagePath == "" || pkg.InBundle || linkedTargets[packagePath] {
			continue
		}

		if pkg.Link && pkg.Resolved != "" {
			if _, found := packages[path.Clean(pkg.Resolved)]; found {
				continue
			}
		}

		// A link without its target is incomplete, even with integrity.
		pinned := false
		if !pkg.Link {
			pinned = validNpmIntegrity(pkg.Integrity)
			if isNpmGitReference(pkg.Resolved) {
				pinned = pinnedNpmGitCommit(pkg.Resolved)
			}
		}

		result.Applicable++
		result.Packages = append(result.Packages, npmLockPackageResult{
			Path:   packagePath,
			Pinned: pinned,
		})
		if !pinned {
			result.Invalid = append(result.Invalid, packagePath)
		}
	}

	sort.Strings(result.Invalid)
	sort.Slice(result.Packages, func(i, j int) bool {
		return result.Packages[i].Path < result.Packages[j].Path
	})
	return result
}

func verifyNpmLockFileV1(
	dependencies map[string]npmDependency,
) npmLockVerification {
	result := npmLockVerification{}

	stack := make([]pendingNpmDependency, 0, len(dependencies))
	for name, dep := range dependencies {
		stack = append(stack, pendingNpmDependency{
			name: name,
			dep:  dep,
		})
	}

	for len(stack) > 0 {
		last := len(stack) - 1
		item := stack[last]
		stack = stack[:last]

		if !item.dep.Bundled &&
			!isNpmLocalDirectory(item.dep.Version) {
			pinned := validNpmIntegrity(item.dep.Integrity)
			if isNpmGitReference(item.dep.Version) {
				pinned = pinnedNpmGitCommit(item.dep.Version)
			}
			result.Applicable++
			result.Packages = append(result.Packages, npmLockPackageResult{
				Path:   item.name,
				Pinned: pinned,
			})
			if !pinned {
				result.Invalid = append(result.Invalid, item.name)
			}
		}

		for childName, child := range item.dep.Dependencies {
			stack = append(stack, pendingNpmDependency{
				name: item.name + "/node_modules/" + childName,
				dep:  child,
			})
		}
	}

	sort.Strings(result.Invalid)

	sort.Slice(result.Packages, func(i, j int) bool {
		return result.Packages[i].Path < result.Packages[j].Path
	})

	return result
}

func isNpmGitReference(value string) bool {
	return strings.HasPrefix(value, "git+") ||
		strings.HasPrefix(value, "git://")
}

func pinnedNpmGitCommit(value string) bool {
	_, ref, found := strings.Cut(value, "#")
	if !found || len(ref) != 40 {
		return false
	}

	_, err := hex.DecodeString(ref)
	return err == nil
}

func validNpmIntegrity(value string) bool {
	for _, token := range strings.Fields(value) {
		token, _, _ = strings.Cut(token, "?")

		algorithm, digest, ok := strings.Cut(token, "-")
		if !ok {
			continue
		}

		var expectedLen int

		switch algorithm {
		case "sha1":
			expectedLen = 20
		case "sha512":
			expectedLen = sha512.Size
		default:
			continue
		}

		decoded, err := base64.StdEncoding.DecodeString(digest)
		if err != nil {
			decoded, err = base64.RawStdEncoding.DecodeString(digest)
		}
		if err != nil {
			continue
		}

		if len(decoded) == expectedLen {
			return true
		}
	}

	return false
}

func isNpmLocalDirectory(version string) bool {
	if !strings.HasPrefix(version, "file:") {
		return false
	}

	localPath := strings.ToLower(strings.TrimPrefix(version, "file:"))
	return !strings.HasSuffix(localPath, ".tgz") &&
		!strings.HasSuffix(localPath, ".tar.gz") &&
		!strings.HasSuffix(localPath, ".tar")
}

func missingNpmLockDependencies(
	manifestContent, lockContent []byte,
) ([]string, error) {
	return missingNpmLockDependenciesAt(manifestContent, lockContent, "")
}

func missingNpmLockDependenciesAt(
	manifestContent, lockContent []byte,
	packagePath string,
) ([]string, error) {
	var lock npmLockFile
	if err := json.Unmarshal(lockContent, &lock); err != nil {
		return nil, fmt.Errorf("parse npm lockfile: %w", err)
	}

	return missingNpmDependenciesInLock(manifestContent, &lock, packagePath)
}

func missingNpmDependenciesInLock(
	manifestContent []byte,
	lock *npmLockFile,
	packagePath string,
) ([]string, error) {
	var manifest npmManifest
	if err := json.Unmarshal(manifestContent, &manifest); err != nil {
		return nil, fmt.Errorf("parse npm manifest: %w", err)
	}

	switch lock.LockfileVersion {
	case 1:
	case 2, 3:
		if lock.Packages == nil {
			return nil, errNpmLockfileMissingPackages
		}
	default:
		return nil, fmt.Errorf("%w: %d",
			errNpmLockfileUnsupportedVersion, lock.LockfileVersion)
	}

	declared := make(map[string]bool)
	for _, dependencies := range []map[string]string{
		manifest.Dependencies,
		manifest.DevDependencies,
		manifest.OptionalDependencies,
	} {
		for name := range dependencies {
			declared[name] = true
		}
	}

	var missing []string
	for name := range declared {
		var present bool
		if lock.LockfileVersion == 1 {
			_, present = lock.Dependencies[name]
		} else {
			present = npmLockContainsDependency(
				lock.Packages, packagePath, name,
			)
		}
		if !present {
			missing = append(missing, name)
		}
	}

	sort.Strings(missing)
	return missing, nil
}

func npmLockContainsDependency(
	packages map[string]npmPackage,
	packagePath, name string,
) bool {
	dir := path.Clean(packagePath)
	for {
		if path.Base(dir) != "node_modules" {
			if _, found := packages[path.Join(dir, "node_modules", name)]; found {
				return true
			}
		}
		if dir == "." || dir == "/" {
			return false
		}
		dir = path.Dir(dir)
	}
}

func missingNpmProjectDependencies(
	lockContent []byte,
	projectDir string,
	manifests map[string][]byte,
) ([]npmLockPackageResult, error) {
	var lock npmLockFile
	if err := json.Unmarshal(lockContent, &lock); err != nil {
		return nil, fmt.Errorf("parse npm lockfile: %w", err)
	}

	packagePaths := map[string]bool{"": true}
	for _, pkg := range lock.Packages {
		if !pkg.Link || pkg.Resolved == "" {
			continue
		}

		target := path.Clean(pkg.Resolved)
		if path.IsAbs(target) ||
			target == ".." ||
			strings.HasPrefix(target, "../") {
			continue
		}
		packagePaths[target] = true
	}

	declaredWorkspaces, err := npmDeclaredWorkspacePaths(projectDir, manifests)
	if err != nil {
		return nil, err
	}
	for _, workspace := range declaredWorkspaces {
		packagePaths[workspace] = true
	}

	var paths []string
	for packagePath := range packagePaths {
		paths = append(paths, packagePath)
	}
	sort.Strings(paths)

	var missing []npmLockPackageResult
	for _, workspace := range declaredWorkspaces {
		if _, found := lock.Packages[workspace]; !found {
			missing = append(missing, npmLockPackageResult{
				Path:   workspace,
				Pinned: false,
			})
		}
	}
	var comparisonErrors []error
	for _, packagePath := range paths {
		manifestDir := filepath.Join(
			projectDir, filepath.FromSlash(packagePath),
		)
		manifest, found := manifests[manifestDir]
		if !found {
			continue
		}

		names, err := missingNpmDependenciesInLock(
			manifest, &lock, packagePath,
		)
		if err != nil {
			comparisonErrors = append(comparisonErrors,
				fmt.Errorf("manifest %s: %w",
					filepath.Join(manifestDir, "package.json"), err))
			continue
		}

		for _, name := range names {
			missing = append(missing, npmLockPackageResult{
				Path:   path.Join(packagePath, "node_modules", name),
				Pinned: false,
			})
		}
	}
	return missing, errors.Join(comparisonErrors...)
}

func npmDeclaredWorkspacePaths(
	projectDir string,
	manifests map[string][]byte,
) ([]string, error) {
	content, found := manifests[projectDir]
	if !found {
		return nil, nil
	}

	var manifest npmManifest
	if err := json.Unmarshal(content, &manifest); err != nil {
		return nil, fmt.Errorf("parse npm manifest: %w", err)
	}
	if len(manifest.Workspaces) == 0 ||
		string(manifest.Workspaces) == "null" {
		return nil, nil
	}

	var patterns []string
	if err := json.Unmarshal(manifest.Workspaces, &patterns); err != nil {
		var object struct {
			Packages []string `json:"packages"`
		}
		if err := json.Unmarshal(manifest.Workspaces, &object); err != nil {
			return nil, fmt.Errorf("parse npm workspaces: %w", err)
		}
		patterns = object.Packages
	}

	var matchers []glob.Glob
	for _, pattern := range patterns {
		if strings.HasPrefix(pattern, "!") {
			return nil, fmt.Errorf(
				"%w: negated workspace pattern %q",
				errNpmWorkspacePatternUnsupported, pattern,
			)
		}
		normalized := pattern
		for strings.HasPrefix(normalized, "./") {
			normalized = strings.TrimPrefix(normalized, "./")
		}
		matcher, err := glob.Compile(strings.TrimSuffix(normalized, "/"), '/')
		if err != nil {
			return nil, fmt.Errorf("parse workspace pattern %q: %w",
				pattern, err)
		}
		matchers = append(matchers, matcher)
	}

	var workspaces []string
	for manifestDir := range manifests {
		relative, err := filepath.Rel(projectDir, manifestDir)
		if err != nil {
			return nil, fmt.Errorf("relative workspace path: %w", err)
		}
		relative = filepath.ToSlash(relative)
		if relative == "." || relative == ".." ||
			strings.HasPrefix(relative, "../") {
			continue
		}

		// Installed packages are not project workspaces.
		if strings.Contains("/"+relative+"/", "/node_modules/") {
			continue
		}

		for _, matcher := range matchers {
			if matcher.Match(relative) {
				workspaces = append(workspaces, relative)
				break
			}
		}
	}

	sort.Strings(workspaces)
	return workspaces, nil
}
