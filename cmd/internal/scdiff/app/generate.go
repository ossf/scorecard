// Copyright 2023 OpenSSF Scorecard Authors
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

package app

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

//nolint:gochecknoinits // common for cobra apps
func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.PersistentFlags().StringVarP(&repoFile, "repos", "r", "", "path to newline-delimited repo file")
}

var (
	repoFile string

	generateCmd = &cobra.Command{
		Use:   "generate [flags] repofile",
		Short: "Generate Scorecard results for diffing",
		Long:  `Generate Scorecard results for diffing`,
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := os.Open(repoFile)
			if err != nil {
				return fmt.Errorf("unable to open repo file: %w", err)
			}
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				fmt.Fprintf(os.Stdout, "running for repo: %v\n", scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("reading repo file: %w", err)
			}
			return nil
		},
	}
)
