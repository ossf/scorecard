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

// Package options implements Scorecard options.
package options

import (
	"errors"
	"os"

	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/log"
)

// Options define common options for configuring scorecard.
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
	ChecksToRun []string
	Metadata    []string
	ShowDetails bool
}

// New creates a new instance of `Options`.
func New() *Options {
	return &Options{}
}

const (
	// DefaultCommit specifies the default commit reference to use.
	DefaultCommit = clients.HeadSHA

	// Formats.

	// FormatJSON specifies that results should be output in JSON format.
	FormatJSON = "json"
	// FormatSarif specifies that results should be output in SARIF format.
	FormatSarif = "sarif"
	// FormatDefault specifies that results should be output in default format.
	FormatDefault = "default"
	// FormatRaw specifies that results should be output in raw format.
	FormatRaw = "raw"

	// Environment variables.

	// EnvVarEnableSarif is the environment variable which controls enabling
	// SARIF logging.
	EnvVarEnableSarif = "ENABLE_SARIF"
	// EnvVarScorecardV6 is the environment variable which enables scorecard v6
	// options.
	EnvVarScorecardV6 = "SCORECARD_V6"
)

var (
	// DefaultLogLevel retrieves the default log level.
	DefaultLogLevel = log.DefaultLevel.String()

	errCommitIsEmpty            = errors.New("commit should be non-empty")
	errCommitOptionNotSupported = errors.New("commit option is not supported yet")
	errFormatNotSupported       = errors.New("unsupported format")
	errPolicyFileNotSupported   = errors.New("policy file is not supported yet")
	errRawOptionNotSupported    = errors.New("raw option is not supported yet")
	errRepoOptionMustBeSet      = errors.New(
		"exactly one of `repo`, `npm`, `pypi`, `rubygems` or `local` must be set",
	)
	errSARIFNotSupported = errors.New("SARIF format is not supported yet")
)

// Validate validates scorecard configuration options.
// TODO(options): Cleanup error messages.
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
			errRepoOptionMustBeSet,
		)
	}

	// Validate SARIF features are flag-guarded.
	if !IsSarifEnabled() {
		if o.Format == FormatSarif {
			errs = append(
				errs,
				errSARIFNotSupported,
			)
		}
		if o.PolicyFile != "" {
			errs = append(
				errs,
				errPolicyFileNotSupported,
			)
		}
	}

	// Validate V6 features are flag-guarded.
	if !isV6Enabled() {
		if o.Format == FormatRaw {
			errs = append(
				errs,
				errRawOptionNotSupported,
			)
		}
		if o.Commit != clients.HeadSHA {
			errs = append(
				errs,
				errCommitOptionNotSupported,
			)
		}
	}

	// Validate format.
	if !validateFormat(o.Format) {
		errs = append(
			errs,
			errFormatNotSupported,
		)
	}

	// Validate `commit` is non-empty.
	if o.Commit == "" {
		errs = append(
			errs,
			errCommitIsEmpty,
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

// IsSarifEnabled returns true if `EnvVarEnableSarif` is specified.
// TODO(options): This probably doesn't need to be exported.
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
