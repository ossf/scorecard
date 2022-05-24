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
	"errors"
	"path/filepath"
)

const (
	configDir      = "starter-workflows/code-scanning"
	configFilename = "scorecards.yml"
)

var errOwnerNotSpecified = errors.New("owner not specified")

// Options are installation options for the scorecard action.
type Options struct {
	// Scorecard GitHub Action configuration path
	ConfigPath string

	// GitHub org/repo owner
	Owner string

	// Repositories
	Repositories []string
}

// New creates a new instance of installation options.
func New() *Options {
	opts := &Options{}
	opts.ConfigPath = GetConfigPath()
	return opts
}

// Validate checks if the installation options specified are valid.
func (o *Options) Validate() error {
	if o.Owner == "" {
		return errOwnerNotSpecified
	}

	return nil
}

// GetConfigPath returns the local path for the scorecard action config file.
// TODO: Consider making this configurable.
func GetConfigPath() string {
	return filepath.Join(configDir, configFilename)
}
