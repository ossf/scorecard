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
	errMissingInputFiles = errors.New("must provide at least two files")
	errMismatchedResults = errors.New("results differ")
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
	xScanner := bufio.NewScanner(x)
	yScanner := bufio.NewScanner(y)
	for {
		xMore := xScanner.Scan()
		yMore := yScanner.Scan()
		if xMore != yMore {
			return errNumResults
		}
		if !xMore && !yMore {
			break
		}
		xResult, err := loadResult(xScanner.Text())
		if err != nil {
			return fmt.Errorf("parsing file1: %w", err)
		}
		yResult, err := loadResult(yScanner.Text())
		if err != nil {
			return fmt.Errorf("parsing file2: %w", err)
		}
		if !compare.Results(&xResult, &yResult) {
			// go-cmp says its not production ready. Is this a valid usage?
			// it certainly helps with readability.
			fmt.Fprintf(output, "%s\n", cmp.Diff(xResult, yResult))
			return errMismatchedResults
		}
	}
	if err := xScanner.Err(); err != nil {
		return fmt.Errorf("reading results: %w", err)
	} else if err := yScanner.Err(); err != nil {
		return fmt.Errorf("reading results: %w", err)
	}
	return nil
}

func loadResult(s string) (pkg.ScorecardResult, error) {
	reader := strings.NewReader(s)
	result, err := pkg.ExperimentalFromJSON2(reader)
	if err != nil {
		return pkg.ScorecardResult{}, fmt.Errorf("parsing result: %w", err)
	}
	format.Normalize(&result)
	return result, nil
}
