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

package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
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
	// Build date in ISO8601 format, output of $(date -u +'%Y-%m-%dT%H:%M:%SZ').
	buildDate = "unknown"
)

//nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// not using logger, since it prints timing info, etc
		fmt.Printf("GitVersion:\t%s\n", gitVersion)
		fmt.Printf("GitCommit:\t%s\n", gitCommit)
		fmt.Printf("GitTreeState:\t%s\n", gitTreeState)
		fmt.Printf("BuildDate:\t%s\n", buildDate)
		fmt.Printf("GoVersion:\t%s\n", runtime.Version())
		fmt.Printf("Compiler:\t%s\n", runtime.Compiler)
		fmt.Printf("Platform:\t%s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}
