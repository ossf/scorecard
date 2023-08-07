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
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ossf/scorecard/v4/cmd/internal/scdiff/app/compare"
	"github.com/ossf/scorecard/v4/pkg"
)

//nolint:gochecknoinits // common for cobra apps
func init() {
	rootCmd.AddCommand(compareCmd)
}

var (
	errMissingInputFiles = errors.New("must provide at least two files")

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
			r1, err := pkg.ExperimentalFromJSON2(f1)
			if err != nil {
				return fmt.Errorf("parsing %q: %w", args[0], err)
			}
			r2, err := pkg.ExperimentalFromJSON2(f2)
			if err != nil {
				return fmt.Errorf("parsing %q: %w", args[1], err)
			}
			compare.Results(&r1, &r2)
			return nil
		},
	}
)
