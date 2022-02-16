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

	"github.com/spf13/cobra"

	"github.com/ossf/scorecard/v4/pkg"
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
		fmt.Printf("GitVersion:\t%s\n", pkg.GetTagVersion())
		fmt.Printf("GitCommit:\t%s\n", pkg.GetCommit())
		fmt.Printf("GitTreeState:\t%s\n", pkg.GetTreeState())
		fmt.Printf("BuildDate:\t%s\n", pkg.GetBuildDate())
		fmt.Printf("GoVersion:\t%s\n", pkg.GetGoVersion())
		fmt.Printf("Compiler:\t%s\n", pkg.GetCompiler())
		fmt.Printf("Platform:\t%s/%s\n", pkg.GetOS(), pkg.GetArch())
	},
}
