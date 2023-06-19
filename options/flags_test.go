// Copyright 2023 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package options

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/cobra"
)

//nolint:gocognit
func TestOptions_AddFlags(t *testing.T) {
	t.Parallel()
	tests := []struct { //nolint:govet
		name string
		opts *Options
	}{
		{
			name: "custom options",
			opts: &Options{
				Repo:        "owner/repo",
				Local:       "/path/to/local",
				Commit:      "1234567890abcdef",
				LogLevel:    "debug",
				NPM:         "npm-package",
				PyPI:        "pypi-package",
				RubyGems:    "rubygems-package",
				Metadata:    []string{"key1=value1", "key2=value2"},
				ShowDetails: true,
				ChecksToRun: []string{"check1", "check2"},
				PolicyFile:  "policy-file",
				Format:      "json",
				ResultsFile: "resultsFile.log",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := &cobra.Command{}
			tt.opts.AddFlags(cmd)

			// check FlagRepo
			if cmd.Flag(FlagRepo).Value.String() != tt.opts.Repo {
				t.Errorf("expected FlagRepo to be %q, but got %q", tt.opts.Repo, cmd.Flag(FlagRepo).Value.String())
			}

			// check FlagLocal
			if cmd.Flag(FlagLocal).Value.String() != tt.opts.Local {
				t.Errorf("expected FlagLocal to be %q, but got %q", tt.opts.Local, cmd.Flag(FlagLocal).Value.String())
			}

			// check FlagCommit
			if cmd.Flag(FlagCommit).Value.String() != tt.opts.Commit {
				t.Errorf("expected FlagCommit to be %q, but got %q", tt.opts.Commit, cmd.Flag(FlagCommit).Value.String())
			}

			// check FlagLogLevel
			if cmd.Flag(FlagLogLevel).Value.String() != tt.opts.LogLevel {
				t.Errorf("expected FlagLogLevel to be %q, but got %q", tt.opts.LogLevel, cmd.Flag(FlagLogLevel).Value.String())
			}

			// check FlagNPM
			if cmd.Flag(FlagNPM).Value.String() != tt.opts.NPM {
				t.Errorf("expected FlagNPM to be %q, but got %q", tt.opts.NPM, cmd.Flag(FlagNPM).Value.String())
			}

			// check FlagPyPI
			if cmd.Flag(FlagPyPI).Value.String() != tt.opts.PyPI {
				t.Errorf("expected FlagPyPI to be %q, but got %q", tt.opts.PyPI, cmd.Flag(FlagPyPI).Value.String())
			}

			// check FlagRubyGems
			if cmd.Flag(FlagRubyGems).Value.String() != tt.opts.RubyGems {
				t.Errorf("expected FlagRubyGems to be %q, but got %q", tt.opts.RubyGems, cmd.Flag(FlagRubyGems).Value.String())
			}

			// check ResultsFile
			if cmd.Flag(FlagResultsFile).Value.String() != tt.opts.ResultsFile {
				t.Errorf("expected ResultsFile to be %q, but got %q", tt.opts.ResultsFile, cmd.Flag(FlagResultsFile).Value.String())
			}

			var e1 []string
			for _, f := range strings.Split(cmd.Flag(FlagChecks).Value.String(), ",") {
				f = strings.TrimPrefix(f, "[")
				f = strings.TrimSuffix(f, "]")
				e1 = append(e1, f)
			}
			if !cmp.Equal(e1, tt.opts.ChecksToRun) {
				t.Errorf("expected FlagChecks to be %q, but got %q", tt.opts.ChecksToRun, e1)
			}
			// check FlagFormat
			if cmd.Flag(FlagFormat).Value.String() != tt.opts.Format {
				t.Errorf("expected FlagFormat to be %q, but got %q", tt.opts.Format, cmd.Flag(FlagFormat).Value.String())
			}
		})
	}
}

func TestOptions_AddFlags_ChecksToRun(t *testing.T) {
	tests := []struct {
		name     string
		opts     *Options
		expected []string
	}{
		{
			name: "custom options",
			opts: &Options{
				ChecksToRun: []string{"check1", "check2"},
			},
			expected: []string{"check1", "check2"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := []string{}
			cmd := &cobra.Command{}
			tt.opts.AddFlags(cmd)
			for _, f := range strings.Split(cmd.Flag(FlagChecks).Value.String(), ",") {
				f = strings.TrimPrefix(f, "[")
				f = strings.TrimSuffix(f, "]")
				e = append(e, f)
			}
			if !cmp.Equal(e, tt.expected) {
				t.Errorf("expected FlagChecks to be %q, but got %q", tt.expected, e)
			}
		})
	}
}

func TestOptions_AddFlags_Format(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		opts     *Options
		expected []string
	}{
		{
			name:     "default options",
			opts:     &Options{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := &cobra.Command{}
			tt.opts.AddFlags(cmd)
			if !cmp.Equal(cmd.Flag(FlagFormat).Value.String(), strings.Join(tt.expected, ", ")) {
				t.Errorf("expected FlagFormat to be %q, but got %q", strings.Join(tt.expected, ", "), cmd.Flag(FlagFormat).Value.String()) //nolint:lll
			}
		})
	}
}
