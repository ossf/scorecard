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

	"github.com/ossf/scorecard/v4/log"
)

// nolint:paralleltest
// because we are using t.Setenv.
func TestGetClients(t *testing.T) { //nolint:gocognit
	type args struct { //nolint:govet
		ctx      context.Context
		repoURI  string
		localURI string
		logger   *log.Logger
	}
	tests := []struct { //nolint:govet
		name                  string
		args                  args
		shouldOSSFuzzBeNil    bool
		shouldRepoClientBeNil bool
		shouldVulnClientBeNil bool
		shouldRepoBeNil       bool
		shouldCIIBeNil        bool
		wantErr               bool
		experimental          bool
	}{
		{
			name: "localURI is not empty",
			args: args{
				ctx:      context.Background(),
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
				ctx:      context.Background(),
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
			name: "repoURI is gitlab which is not supported",
			args: args{
				ctx:      context.Background(),
				repoURI:  "https://gitlab.com/ossf/scorecard",
				localURI: "",
			},
			shouldOSSFuzzBeNil:    false,
			shouldRepoClientBeNil: false,
			shouldVulnClientBeNil: false,
			shouldRepoBeNil:       true,
			wantErr:               true,
		},
		{
			name: "repoURI is gitlab and experimental is true",
			args: args{
				ctx:      context.Background(),
				repoURI:  "https://gitlab.com/ossf/scorecard",
				localURI: "",
			},
			shouldOSSFuzzBeNil:    false,
			shouldRepoBeNil:       false,
			shouldRepoClientBeNil: false,
			shouldVulnClientBeNil: false,
			shouldCIIBeNil:        false,
			wantErr:               false,
			experimental:          true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.experimental {
				t.Setenv("SCORECARD_EXPERIMENTAL", "true")
			}
			got, repoClient, ossFuzzClient, ciiClient, vulnsClient, err := GetClients(tt.args.ctx, tt.args.repoURI, tt.args.localURI, tt.args.logger) //nolint:lll
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
		})
	}
}
