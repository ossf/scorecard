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
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/pkg"
)

//nolint:gochecknoinits // common for cobra apps
func init() {
	rootCmd.AddCommand(statsCmd)
	statsCmd.PersistentFlags().StringVarP(&statsCheck, "check", "c", "", "Analyze breakdown of a single check")
}

var (
	statsCheck string
	statsCmd   = &cobra.Command{
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

	errCheckNotPresent = errors.New("requested check not present")
	errInvalidScore    = errors.New("invalid score")
)

// countScores quantizes the scores into 12 buckets, from [-1, 10]
// If a check is provided, that check score is used, otherwise the aggregate score is used.
func countScores(input io.Reader, check string) ([12]int, error) {
	var counts [12]int // [-1, 10] inclusive
	var score int
	scanner := bufio.NewScanner(input)
	scanner.Buffer(nil, 1024*1024) // TODO, how big is big enough?
	for scanner.Scan() {
		result, aggregateScore, err := pkg.ExperimentalFromJSON2(strings.NewReader(scanner.Text()))
		if err != nil {
			return [12]int{}, fmt.Errorf("parsing result: %w", err)
		}
		if check == "" {
			score = int(aggregateScore)
		} else {
			i := slices.IndexFunc(result.Checks, func(c checker.CheckResult) bool {
				return strings.EqualFold(c.Name, check)
			})
			if i == -1 {
				return [12]int{}, errCheckNotPresent
			}
			score = result.Checks[i].Score
		}
		if score < -1 || score > 10 {
			return [12]int{}, errInvalidScore
		}
		bucket := score + 1 // score of -1 is index 0, score of 0 is index 1, etc.
		counts[bucket]++
	}
	if err := scanner.Err(); err != nil {
		return [12]int{}, fmt.Errorf("parsing golden file: %w", err)
	}
	return counts, nil
}

func calcStats(input io.Reader, output io.Writer) error {
	counts, err := countScores(input, statsCheck)
	if err != nil {
		return err
	}
	name := statsCheck
	if name == "" {
		name = "Aggregate"
	}
	summary(name, &counts, output)
	return nil
}

func summary(name string, counts *[12]int, output io.Writer) {
	const (
		minWidth = 0
		tabWidth = 4
		padding  = 1
		padchar  = ' '
		flags    = tabwriter.AlignRight
	)
	w := tabwriter.NewWriter(output, minWidth, tabWidth, padding, padchar, flags)
	fmt.Fprintf(w, "%s Score\tCount\t\n", name)
	for i, c := range counts {
		scoreBucket := i - 1
		fmt.Fprintf(w, "%d\t%d\t\n", scoreBucket, c)
	}
	w.Flush()
}
