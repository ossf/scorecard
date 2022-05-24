// Copyright 2022 OpenSSF Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ossf/scorecard-action/install"
	"github.com/ossf/scorecard-action/install/options"
)

const (
	cmdUsage     = `--owner example_org [--repos <repo1,repo2,repo3>]`
	cmdDescShort = "Scorecard GitHub Action installer"
	cmdDescLong  = `
The Scorecard GitHub Action installer simplifies the installation of the
scorecard GitHub Action by creating pull requests through the command line.`
)

// New creates a new instance of the scorecard action installation command.
func New(o *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdUsage,
		Short: cmdDescShort,
		Long:  cmdDescLong,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate()
			if err != nil {
				return fmt.Errorf("validating options: %w", err)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return rootCmd(o)
		},
	}

	o.AddFlags(cmd)
	return cmd
}

// rootCmd runs scorecard checks given a set of arguments.
func rootCmd(o *options.Options) error {
	err := install.Run(o)
	if err != nil {
		return fmt.Errorf("running scorecard installation: %w", err)
	}

	return nil
}
