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
	"strings"

	"github.com/spf13/cobra"

	"github.com/ossf/scorecard/v4/pkg"
)

//nolint:gochecknoinits // common for cobra apps
func init() {
	rootCmd.AddCommand(statsCmd)
}

var statsCmd = &cobra.Command{
	Use:   "stats [flags] FILE",
	Short: "Summarize stats for a golden file",
	Long:  `Summarize stats for a golden file`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errMissingInputFiles // TODO: generalize this?
		}
		f1, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("opening %q: %w", args[0], err)
		}
		defer f1.Close()
		return calcStats(f1, os.Stdout)
	},
}

func calcStats(input io.Reader, output io.Writer) error {
	var counts [12]int // [-1, 10] inclusive
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		result, err := pkg.ExperimentalFromJSON2(strings.NewReader(scanner.Text()))
		if err != nil {
			return fmt.Errorf("parsing result: %w", err)
		}
		score, err := result.GetAggregateScore(nil) // todo, read from the file?
		if score < -1 || score > 10 {
			return fmt.Errorf("invalid score") // todo sentinel
		}
		bucket := int(score) + 1
		counts[bucket]++
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("parsing golden file: %w", err)
	}
	summary(&counts, output)
	return nil
}

func summary(counts *[12]int, output io.Writer) {
	fmt.Fprintln(output, "Score Distribution:")
	for i, c := range counts {
		scoreBucket := i - 1
		fmt.Fprintf(output, "%d: %d", scoreBucket, c)
	}
}
