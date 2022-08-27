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
	"fmt"
	"os"

	"github.com/caarlos0/env/v6"

	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/log"
)

// Options define common options for configuring scorecard.
type Options struct {
	Repo       string
	Local      string
	Commit     string
	LogLevel   string
	Format     string
	NPM        string
	PyPI       string
	RubyGems   string
	PolicyFile string
	// TODO(action): Add logic for writing results to file
	ResultsFile string
	ChecksToRun []string
	Metadata    []string
	ShowDetails bool

	// Feature flags.
	EnableSarif       bool `env:"ENABLE_SARIF"`
	EnableScorecardV6 bool `env:"SCORECARD_V6"`
	EnableScorecardExperimental bool `env:"SCORECARD_EXPERIMENTAL"`
}

// New creates a new instance of `Options`.
func New() *Options {
	opts := &Options{}
	if err := env.Parse(opts); err != nil {
		fmt.Printf("could not parse env vars, using default options: %v", err)
	}

	// Defaulting.
	// TODO(options): Consider moving this to a separate function/method.
	if opts.Commit == "" {
		opts.Commit = DefaultCommit
	}
	if opts.Format == "" {
		opts.Format = FormatDefault
	}
	if opts.LogLevel == "" {
		opts.LogLevel = DefaultLogLevel
	}

	return opts
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
	// EnvVarScorecardExperimental is the environment variable which enables
	// scorecard experimental features.
	EnvVarScorecardExperimental = "SCORECARD_EXPERIMENTAL"
)

var (
	// DefaultLogLevel retrieves the default log level.
	DefaultLogLevel = log.DefaultLevel.String()

	errCommitIsEmpty          = errors.New("commit should be non-empty")
	errFormatNotSupported     = errors.New("unsupported format")
	errPolicyFileNotSupported = errors.New("policy file is not supported yet")
	errRawOptionNotSupported  = errors.New("raw option is not supported yet")
	errRepoOptionMustBeSet    = errors.New(
		"exactly one of `repo`, `npm`, `pypi`, `rubygems` or `local` must be set",
	)
	errSARIFNotSupported    = errors.New("SARIF format is not supported yet")
	errValidate             = errors.New("some options could not be validated")
	errExperimentalDisabled = errors.New("scorecard experimental features are disabled")
)

// Validate validates scorecard configuration options.
// TODO(options): Cleanup error messages.
func (o *Options) Validate() error {
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
	if !o.isSarifEnabled() {
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
	if !o.isV6Enabled() {
		if o.Format == FormatRaw {
			errs = append(
				errs,
				errRawOptionNotSupported,
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

	if len(errs) != 0 {
		return fmt.Errorf(
			"%w: %+v",
			errValidate,
			errs,
		)
	}

	return nil
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

// Feature flags.

// isSarifEnabled returns true if SARIF format was specified in options or via
// environment variable.
func (o *Options) isSarifEnabled() bool {
	// UPGRADEv4: remove.
	_, enabled := os.LookupEnv(EnvVarEnableSarif)
	return o.EnableSarif || enabled
}

// isV6Enabled returns true if v6 functionality was specified in options or via
// environment variable.
func (o *Options) isV6Enabled() bool {
	_, enabled := os.LookupEnv(EnvVarScorecardV6)
	return o.EnableScorecardV6 || enabled
}

func validateFormat(format string) bool {
	switch format {
	case FormatJSON, FormatSarif, FormatDefault, FormatRaw:
		return true
	default:
		return false
	}
}

// ValidateExperimental returns true if SCORECARD_EXPERIMENTAL was specified in options or via
// environment variable.
func (o *Options) ValidateExperimental() error {
	_, enabled := os.LookupEnv(EnvVarScorecardExperimental)
	if !(o.EnableScorecardExperimental || enabled) {
		return fmt.Errorf(
			"cannot use this feature: %w",
			errExperimentalDisabled,
		)
	}
	return nil
}
