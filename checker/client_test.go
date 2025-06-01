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
package checker

import (
	"context"
	"testing"

	"github.com/ossf/scorecard/v5/log"
)

//nolint:gocognit
func TestGetClients(t *testing.T) {
	type args struct { //nolint:govet
		ctx      context.Context
		repoURI  string
		localURI string
		logger   *log.Logger
	}
	tests := []struct { //nolint:govet
		name                     string
		args                     args
		shouldOSSFuzzBeNil       bool
		shouldRepoClientBeNil    bool
		shouldVulnClientBeNil    bool
		shouldRepoBeNil          bool
		shouldCIIBeNil           bool
		shouldProjectClientBeNil bool
		wantErr                  bool
		experimental             bool
		isGhHost                 bool
	}{
		{
			name: "localURI is not empty",
			args: args{
				ctx:      t.Context(),
				repoURI:  "",
				localURI: "foo",
			},
			shouldOSSFuzzBeNil:    false,
			shouldRepoClientBeNil: false,
			shouldVulnClientBeNil: false,
			shouldRepoBeNil:       true,
			wantErr:               true,
		},
		{
			name: "repoURI is not empty",
			args: args{
				ctx:      t.Context(),
				repoURI:  "foo",
				localURI: "",
			},
			shouldOSSFuzzBeNil:    false,
			shouldRepoClientBeNil: false,
			shouldVulnClientBeNil: false,
			shouldRepoBeNil:       true,
			wantErr:               true,
		},
		{
			name: "repoURI is gitlab which is supported",
			args: args{
				ctx:      t.Context(),
				repoURI:  "https://gitlab.com/ossf-test/scorecard",
				localURI: "",
			},
			shouldOSSFuzzBeNil:    false,
			shouldRepoClientBeNil: false,
			shouldVulnClientBeNil: false,
			shouldRepoBeNil:       false,
			wantErr:               false,
		},
		{
			name: "repoURI is corp github host",
			args: args{
				ctx:      t.Context(),
				repoURI:  "https://github.corp.com/ossf/scorecard",
				localURI: "",
			},
			shouldOSSFuzzBeNil:    false,
			shouldRepoClientBeNil: false,
			shouldVulnClientBeNil: false,
			shouldRepoBeNil:       false,
			shouldCIIBeNil:        false,
			wantErr:               false,
			isGhHost:              true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.experimental {
				t.Setenv("SCORECARD_EXPERIMENTAL", "true")
			}
			if tt.isGhHost {
				t.Setenv("GH_HOST", "github.corp.com")
				t.Setenv("GH_TOKEN", "PAT")
			}
			got, repoClient, ossFuzzClient, ciiClient, vulnsClient, projectClient, err := GetClients(tt.args.ctx, tt.args.repoURI, tt.args.localURI, tt.args.logger)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetClients() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.shouldRepoBeNil != (got == nil) {
				t.Errorf("GetClients() got = %v", got)
			}
			if repoClient != nil && tt.shouldRepoClientBeNil {
				t.Errorf("GetClients() repoClient = %v ", repoClient)
			}
			if ossFuzzClient != nil && tt.shouldOSSFuzzBeNil {
				t.Errorf("GetClients() ossFuzzClient = %v ", ossFuzzClient)
			}
			if ciiClient != nil && tt.shouldCIIBeNil {
				t.Errorf("GetClients() ciiClient = %v", ciiClient)
			}
			if vulnsClient != nil && tt.shouldVulnClientBeNil {
				t.Errorf("GetClients() vulnsClient = %v", vulnsClient)
			}
			if projectClient != nil && tt.shouldProjectClientBeNil {
				t.Errorf("GetClients() projectClient = %v", projectClient)
			}
		})
	}
}
