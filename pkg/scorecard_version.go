// Copyright 2021 Security Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pkg

import (
	"runtime"
	"strings"
)

// Base version information.
//
// This is the fallback data used when version information from git is not
// provided via go ldflags in the Makefile. See version.mk.
var (
	// Output of "git describe". The prerequisite is that the branch should be
	// tagged using the correct versioning strategy.
	gitVersion = "unknown"
	// SHA1 from git, output of $(git rev-parse HEAD).
	gitCommit = "unknown"
	// State of git tree, either "clean" or "dirty".
	gitTreeState = "unknown"
	// Build date in ISO8601 format.
	buildDate = "unknown"
)

// GetTagVersion returns the scorecard version
// fr the release GitHub tag, i.e. v.X.Y.Z.
func GetTagVersion() string {
	return gitVersion
}

// GetSemanticVersion returns the semantic version,
// i.e., X.Y.Z.
func GetSemanticVersion() string {
	tv := GetTagVersion()
	if strings.HasPrefix(tv, "v") {
		return tv[1:]
	}
	return tv
}

// GetCommit returns the GitHub's commit hash that scorecard was built from.
func GetCommit() string {
	return gitCommit
}

// GetTreeState returns the git tree state.
func GetTreeState() string {
	return gitTreeState
}

// GetBuildDate returns the date scorecard was build.
func GetBuildDate() string {
	return buildDate
}

// GetGoVersion returns the Go version used to build scorecard.
func GetGoVersion() string {
	return runtime.Version()
}

// GetOS returns the OS the build can run on.
func GetOS() string {
	return runtime.GOOS
}

// GetArch returns the architecture (e.g., x86) the build can run on.
func GetArch() string {
	return runtime.GOARCH
}

// GetCompiler returns the compiler that was used to build scorecard.
func GetCompiler() string {
	return runtime.Compiler
}
