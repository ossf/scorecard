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
		CommitDepth       int64
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
		EnableSarif       bool
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
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			o := &Options{
				Repo:              tt.fields.Repo,
				Local:             tt.fields.Local,
				Commit:            tt.fields.Commit,
				CommitDepth:       tt.fields.CommitDepth,
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
				EnableSarif:       tt.fields.EnableSarif,
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
