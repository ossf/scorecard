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
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/ossf/scorecard/v4/cmd/internal/scdiff/app/format"
	"github.com/ossf/scorecard/v4/cmd/internal/scdiff/app/runner"
	"github.com/ossf/scorecard/v4/pkg"
)

//nolint:gochecknoinits // common for cobra apps
func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.PersistentFlags().StringVarP(&repoFile, "repos", "r", "", "path to newline-delimited repo file")
	generateCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "write to file instead of stdout")
}

var (
	repoFile   string
	outputFile string

	generateCmd = &cobra.Command{
		Use:   "generate [flags] repofile",
		Short: "Generate Scorecard results for diffing",
		Long:  `Generate Scorecard results for diffing`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input, err := os.Open(repoFile)
			if err != nil {
				return fmt.Errorf("unable to open repo file: %w", err)
			}
			defer input.Close()
			var output io.Writer = os.Stdout
			if outputFile != "" {
				outputF, err := os.Create(outputFile)
				if err != nil {
					return fmt.Errorf("unable to create output file: %w", err)
				}
				defer outputF.Close()
				output = outputF
			}
			r := runner.New()
			return generate(&r, input, output)
		},
	}
)

type scorecardRunner interface {
	Run(repo string) (pkg.ScorecardResult, error)
}

// Runs scorecard on each newline-delimited repo in repos, and writes the output
func generate(scRunner scorecardRunner, repos io.Reader, output io.Writer) error {
	scanner := bufio.NewScanner(repos)
	for scanner.Scan() {
		results, err := scRunner.Run(scanner.Text())
		if err != nil {
			return fmt.Errorf("running scorecard on %s: %w", scanner.Text(), err)
		}
		// TODO pretty print?
		err = format.JSON(&results, output)
		if err != nil {
			return fmt.Errorf("formatting results: %w", err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading repo file: %w", err)
	}
	return nil
}
