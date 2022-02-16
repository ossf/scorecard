// Copyright 2020 Security Scorecard Authors
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

// package options implements Scorecard options.
package options

import (
	"fmt"
	"os"

	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/log"
)

type Options struct {
	Repo        string
	Local       string
	Commit      string
	LogLevel    string
	Format      string
	NPM         string
	PyPI        string
	RubyGems    string
	PolicyFile  string
	ShowDetails bool
	ChecksToRun []string
	Metadata    []string
}

func New() *Options {
	return &Options{}
}

const (
	DefaultCommit = clients.HeadSHA

	// Formats
	FormatJSON    = "json"
	FormatSarif   = "sarif"
	FormatDefault = "default"
	FormatRaw     = "raw"

	// Environment variables
	EnvVarEnableSarif = "ENABLE_SARIF"
	EnvVarScorecardV6 = "SCORECARD_V6"
)

var (
	DefaultLogLevel = log.DefaultLevel.String()
)

// TODO(options): Create explicit error types
// TODO(options): Cleanup error messages
func (o *Options) Validate() []error {
	var errs []error

	// Validate exactly one of `--repo`, `--npm`, `--pypi`, `--rubygems`, `--local` is enabled.
	if boolSum(o.Repo != "",
		o.NPM != "",
		o.PyPI != "",
		o.RubyGems != "",
		o.Local != "") != 1 {
		errs = append(
			errs,
			fmt.Errorf("Exactly one of `--repo`, `--npm`, `--pypi`, `--rubygems` or `--local` must be set"),
		)
	}

	// Validate SARIF features are flag-guarded.
	if !IsSarifEnabled() {
		if o.Format == FormatSarif {
			errs = append(
				errs,
				fmt.Errorf("sarif format not supported yet"),
			)
		}
		if o.PolicyFile != "" {
			errs = append(
				errs,
				fmt.Errorf("policy file not supported yet"),
			)
		}
	}

	// Validate V6 features are flag-guarded.
	if !isV6Enabled() {
		if o.Format == FormatRaw {
			errs = append(
				errs,
				fmt.Errorf("raw option not supported yet"),
			)
		}
		if o.Commit != clients.HeadSHA {
			errs = append(
				errs,
				fmt.Errorf("--commit option not supported yet"),
			)
		}
	}

	// Validate format.
	if !validateFormat(o.Format) {
		errs = append(
			errs,
			fmt.Errorf("unsupported format '%s'", o.Format),
		)
	}

	// Validate `commit` is non-empty.
	if o.Commit == "" {
		errs = append(
			errs,
			fmt.Errorf("commit should be non-empty"),
		)
	}

	return errs
}

func boolSum(bools ...bool) int {
	sum := 0
	for _, b := range bools {
		if b {
			sum++
		}
	}
	return sum
}

// TODO(options): This probably doesn't need to be exported
func IsSarifEnabled() bool {
	// UPGRADEv4: remove.
	var sarifEnabled bool
	_, sarifEnabled = os.LookupEnv(EnvVarEnableSarif)
	return sarifEnabled
}

func isV6Enabled() bool {
	var v6 bool
	_, v6 = os.LookupEnv(EnvVarScorecardV6)
	return v6
}

func validateFormat(format string) bool {
	switch format {
	case FormatJSON, FormatSarif, FormatDefault, FormatRaw:
		return true
	default:
		return false
	}
}
