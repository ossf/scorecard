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

package options

import (
	"github.com/spf13/cobra"
)

const (
	// FlagOwner is the flag name for specifying an repository owner.
	FlagOwner = "owner"

	// FlagRepos is the flag name for specifying a set of repositories.
	FlagRepos = "repos"
)

// Command is an interface for handling options for command-line utilities.
type Command interface {
	// AddFlags adds this options' flags to the cobra command.
	AddFlags(cmd *cobra.Command)
}

// AddFlags adds this options' flags to the cobra command.
func (o *Options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&o.Owner,
		FlagOwner,
		o.Owner,
		"org/owner to install the scorecard action for",
	)

	cmd.Flags().StringSliceVar(
		&o.Repositories,
		FlagRepos,
		o.Repositories,
		"repositories to install the scorecard action on",
	)
}
