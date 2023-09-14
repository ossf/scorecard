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

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/cobra"

	"github.com/ossf/scorecard/v4/cmd/internal/scdiff/app/compare"
	"github.com/ossf/scorecard/v4/cmd/internal/scdiff/app/format"
	"github.com/ossf/scorecard/v4/pkg"
)

//nolint:gochecknoinits // common for cobra apps
func init() {
	rootCmd.AddCommand(compareCmd)
}

var (
	errMissingInputFiles = errors.New("must provide at least two files from scdiff generate")
	errResultsDiffer     = errors.New("results differ")
	errNumResults        = errors.New("number of results being compared differ")

	compareCmd = &cobra.Command{
		Use:   "compare [flags] FILE1 FILE2",
		Short: "Compare Scorecard results",
		Long:  `Compare Scorecard results`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return errMissingInputFiles
			}
			f1, err := os.Open(args[0])
			if err != nil {
				return fmt.Errorf("opening %q: %w", args[0], err)
			}
			defer f1.Close()
			f2, err := os.Open(args[1])
			if err != nil {
				return fmt.Errorf("opening %q: %w", args[1], err)
			}
			defer f2.Close()
			cmd.SilenceUsage = true  // disables printing Usage
			cmd.SilenceErrors = true // disables the "Error: <err>" message
			return compareReaders(f1, f2, os.Stderr)
		},
	}
)

func compareReaders(x, y io.Reader, output io.Writer) error {
	// results are currently newline delimited
	xs := bufio.NewScanner(x)
	ys := bufio.NewScanner(y)
	for {
		if shouldContinue, err := advanceScanners(xs, ys); err != nil {
			return err
		} else if !shouldContinue {
			break
		}
		xResult, yResult, err := loadResults(xs, ys)
		if err != nil {
			return err
		}
		if !compare.Results(&xResult, &yResult) {
			// go-cmp says its not production ready. Is this a valid usage?
			// it certainly helps with readability.
			fmt.Fprintf(output, "%s\n", cmp.Diff(xResult, yResult))
			return errResultsDiffer
		}
	}
	return nil
}

func loadResults(x, y *bufio.Scanner) (pkg.ScorecardResult, pkg.ScorecardResult, error) {
	xResult, err := pkg.ExperimentalFromJSON2(strings.NewReader(x.Text()))
	if err != nil {
		return pkg.ScorecardResult{}, pkg.ScorecardResult{}, fmt.Errorf("parsing first result: %w", err)
	}
	yResult, err := pkg.ExperimentalFromJSON2(strings.NewReader(y.Text()))
	if err != nil {
		return pkg.ScorecardResult{}, pkg.ScorecardResult{}, fmt.Errorf("parsing second result: %w", err)
	}
	format.Normalize(&xResult)
	format.Normalize(&yResult)
	return xResult, yResult, nil
}

// advanceScanners is intended to expand the normal `for scanner.Scan()` semantics to two scanners,
// it keeps the scanners in sync, and determines if iteration should continue.
//
// Iteration should continue until any scanner reaches EOF, or any scanner encounters a non-EOF error.
func advanceScanners(x, y *bufio.Scanner) (shouldContinue bool, err error) {
	xContinue := x.Scan()
	yContinue := y.Scan()
	if err := x.Err(); err != nil {
		return false, fmt.Errorf("reading results: %w", err)
	}
	if err := y.Err(); err != nil {
		return false, fmt.Errorf("reading results: %w", err)
	}
	if xContinue != yContinue {
		return false, errNumResults
	}
	return xContinue, nil
}
