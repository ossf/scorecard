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
	"os"
	"testing"
)

// Cannot run parallel tests because of the ENV variables.
//nolint
func TestOptions_Validate(t *testing.T) {
	type fields struct {
		Repo              string
		Local             string
		Commit            string
		LogLevel          string
		Format            string
		NPM               string
		PyPI              string
		RubyGems          string
		PolicyFile        string
		ResultsFile       string
		ChecksToRun       []string
		Metadata          []string
		ShowDetails       bool
		PublishResults    bool
		EnableSarif       bool
		EnableScorecardV5 bool
		EnableScorecardV6 bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "No options are turned on",
			fields:  fields{},
			wantErr: true,
		},
		{
			name: "format sarif but the enable sarif flag is not set",
			fields: fields{
				Format: "sarif",
			},
			wantErr: true,
		},
		{
			name: "format sarif and the enable sarif flag is set",
			fields: fields{
				Repo:        "github.com/oss/scorecard",
				Commit:      "HEAD",
				Format:      "sarif",
				EnableSarif: true,
				PolicyFile:  "testdata/policy.yaml",
			},
			wantErr: false,
		},
		{
			name: "format sarif and the disabled but the policy file is set",
			fields: fields{
				Repo:       "github.com/oss/scorecard",
				Commit:     "HEAD",
				PolicyFile: "testdata/policy.yaml",
			},
			wantErr: true,
		},
		{
			name: "format raw is not supported when V6 is not enabled",
			fields: fields{
				Repo:   "github.com/oss/scorecard",
				Commit: "HEAD",
				Format: "raw",
			},
			wantErr: true,
		},
		{
			name: "invalid for local repo",
			fields: fields{
				Local:       "testdata/repo",
				ChecksToRun: []string{"Branch-Protection"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			o := &Options{
				Repo:              tt.fields.Repo,
				Local:             tt.fields.Local,
				Commit:            tt.fields.Commit,
				LogLevel:          tt.fields.LogLevel,
				Format:            tt.fields.Format,
				NPM:               tt.fields.NPM,
				PyPI:              tt.fields.PyPI,
				RubyGems:          tt.fields.RubyGems,
				PolicyFile:        tt.fields.PolicyFile,
				ResultsFile:       tt.fields.ResultsFile,
				ChecksToRun:       tt.fields.ChecksToRun,
				Metadata:          tt.fields.Metadata,
				ShowDetails:       tt.fields.ShowDetails,
				PublishResults:    tt.fields.PublishResults,
				EnableSarif:       tt.fields.EnableSarif,
				EnableScorecardV5: tt.fields.EnableScorecardV5,
				EnableScorecardV6: tt.fields.EnableScorecardV6,
			}
			if o.EnableSarif {
				os.Setenv(EnvVarEnableSarif, "1")
				defer os.Unsetenv(EnvVarEnableSarif)
			}

			if err := o.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Options.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOptions_isCheckValidForLocalRepo(t *testing.T) {
	t.Parallel()
	type fields struct {
		Local       string
		ChecksToRun []string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "no checks to run",
			fields:  fields{},
			wantErr: false,
		},
		{
			name: "no local repo with branch protection",
			fields: fields{
				ChecksToRun: []string{"Branch-Protection"},
			},
			wantErr: false,
		},
		{
			name: "local repo with branch protection",
			fields: fields{
				Local:       "testdata/repo",
				ChecksToRun: []string{"Branch-Protection"},
			},
			wantErr: true,
		},
		{
			name: "local repo with no branch protection",
			fields: fields{
				Local:       "testdata/repo",
				ChecksToRun: []string{"No-Branch-Protection"},
			},
			wantErr: true,
		},
		{
			name: "local repo with CII Best Practices",
			fields: fields{
				Local:       "testdata/repo",
				ChecksToRun: []string{"CII-Best-Practices"},
			},
			wantErr: true,
		},
		{
			name: "local repo with Signed release",
			fields: fields{
				Local:       "testdata/repo",
				ChecksToRun: []string{"Signed-Releases"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			o := &Options{
				Local:       tt.fields.Local,
				ChecksToRun: tt.fields.ChecksToRun,
			}
			if err := o.isValidForLocalRepo(); (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v for %v", err, tt.wantErr, tt.name)
			}
		})
	}
}
