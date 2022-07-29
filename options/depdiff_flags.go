// Copyright 2022 Security Scorecard Authors
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

// Package options implements Scorecard options.
package options

import (
	"github.com/spf13/cobra"
)

const (
	// FlagBase is the flag name for specifying a dependency-diff base.
	FlagBase = "base"

	// FlagHead is the flag name for specifying a dependency-diff head.
	FlagHead = "head"

	// FlagChangeTypes is the flag name for specifying the change type for
	// which dependency-diff surfaces the scorecard check results.
	FlagChangeTypes = "change-types"
)

// AddDepdiffFlags adds flags to the dependency-diff cobra command.
func (depOptions *DependencydiffOptions) AddDepdiffFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&depOptions.Base,
		FlagBase,
		depOptions.Base,
		`The base code branch name or the base commitSHA to check. Valid input examples: 
		main (using a branch name), SHA_VALUE_1 (using a commitSHA)`,
	)
	cmd.Flags().StringVar(
		&depOptions.Head,
		FlagHead,
		depOptions.Head,
		`The head code branch name or the head commitSHA to check. Valid input examples: 
		dev (using a branch name), SHA_VALUE_2 (using a commitSHA)`,
	)
	cmd.Flags().StringSliceVar(
		&depOptions.ChangeTypes,
		FlagChangeTypes,
		depOptions.ChangeTypes,
		`Dependency change types for surfacing the scorecard results. This is not a required 
		input and can be null. Possible values are: added,removed`,
	)
}
