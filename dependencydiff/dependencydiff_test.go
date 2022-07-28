// Copyright 2022 Security Scorecard Authors
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

package dependencydiff

import (
	"context"
	"path"
	"testing"

	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/log"
	sclog "github.com/ossf/scorecard/v4/log"
)

// Test_fetchRawDependencyDiffData is a test function for fetchRawDependencyDiffData.
func Test_fetchRawDependencyDiffData(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name      string
		dCtx      dependencydiffContext
		wantEmpty bool
		wantErr   bool
	}{
		{
			name: "error response",
			dCtx: dependencydiffContext{
				logger:    log.NewLogger(log.InfoLevel),
				ctx:       context.Background(),
				ownerName: "no_such_owner",
				repoName:  "repo_not_exist",
				base:      "main",
				head:      clients.HeadSHA,
			},
			wantEmpty: true,
			wantErr:   true,
		},
		// Considering of the token usage, normal responses are tested in the e2e test.
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := fetchRawDependencyDiffData(&tt.dCtx)
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchRawDependencyDiffData() error = {%v}, want error: %v", err, tt.wantErr)
				return
			}
			lenResults := len(tt.dCtx.dependencydiffs)
			if (lenResults == 0) != tt.wantEmpty {
				t.Errorf("want empty results: %v, got len of results:%d", tt.wantEmpty, lenResults)
				return
			}

		})
	}
}

func Test_initRepoAndClientByChecks(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name                           string
		dCtx                           dependencydiffContext
		srcRepo                        string
		wantRepoClient, wantFuzzClient bool
		wantVulnClient, wantCIIClient  bool
		wantErr                        bool
	}{
		{
			name: "error creating repo",
			dCtx: dependencydiffContext{
				logger:          log.NewLogger(log.InfoLevel),
				ctx:             context.Background(),
				checkNamesToRun: []string{},
			},
			srcRepo: path.Join(
				"host_not_exist.com",
				"owner_not_exist",
				"repo_not_exist",
			),
			wantRepoClient: false,
			wantFuzzClient: false,
			wantVulnClient: false,
			wantCIIClient:  false,
			wantErr:        true,
		},
		// Same as the above, putting the normal responses to the e2e test.
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := initRepoAndClientByChecks(&tt.dCtx, tt.srcRepo)
			if (err != nil) != tt.wantErr {
				t.Errorf("initClientByChecks() error = {%v}, want error: %v", err, tt.wantErr)
				return
			}
			if (tt.dCtx.ghRepoClient != nil) != tt.wantRepoClient {
				t.Errorf("init repo error, wantRepoClient: %v, got %v", tt.wantRepoClient, tt.dCtx.ghRepoClient)
				return
			}
			if (tt.dCtx.ossFuzzClient != nil) != tt.wantFuzzClient {
				t.Errorf("init repo error, wantFuzzClient: %v, got %v", tt.wantFuzzClient, tt.dCtx.ossFuzzClient)
				return
			}
			if (tt.dCtx.vulnsClient != nil) != tt.wantVulnClient {
				t.Errorf("init repo error, wantVulnClient: %v, got %v", tt.wantVulnClient, tt.dCtx.vulnsClient)
				return
			}
			if (tt.dCtx.ciiClient != nil) != tt.wantCIIClient {
				t.Errorf("init repo error, wantCIIClient: %v, got %v", tt.wantCIIClient, tt.dCtx.ciiClient)
				return
			}
		})
	}
}

func Test_getScorecardCheckResults(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name    string
		dCtx    dependencydiffContext
		wantErr bool
	}{
		{
			name: "empty response",
			dCtx: dependencydiffContext{
				ctx:       context.Background(),
				logger:    log.NewLogger(log.InfoLevel),
				ownerName: "owner_not_exist",
				repoName:  "repo_not_exist",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := getScorecardCheckResults(&tt.dCtx)
			if (err != nil) != tt.wantErr {
				t.Errorf("getScorecardCheckResults() error = {%v}, want error: %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_ecosystemMapping(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name            string
		deps            []dependency
		ecosystemWanted []string
	}{
		{
			name: "case1",
			deps: []dependency{
				{
					Name:      "dependency_1",
					Ecosystem: asPointer("nuget"),
				},
				{
					Name:      "dependency_2",
					Ecosystem: asPointer("gomod"),
				},
			},
			ecosystemWanted: []string{"NuGet", "Go"},
		},
		{
			name: "case2",
			deps: []dependency{
				{
					Name:      "dependency_3",
					Ecosystem: asPointer("actions"),
				},
			},
			ecosystemWanted: []string{"actions"},
		},
		{
			name: "case3",
			deps: []dependency{
				{
					Name:      "dependency_4",
					Ecosystem: asPointer("rubygems"),
				},
				{
					Name:      "dependency_5",
					Ecosystem: asPointer("pip"),
				},
				{
					Name:      "dependency_6",
					Ecosystem: asPointer("cargo"),
				},
				{
					Name:      "dependency_6",
					Ecosystem: nil,
				},
			},
			ecosystemWanted: []string{"RubyGems", "PyPI", "crates.io"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mapDependencyEcosystemNaming(sclog.NewLogger(sclog.DefaultLevel), tt.deps)
			for i, d := range tt.deps {
				if d.Ecosystem != nil {
					if *d.Ecosystem != tt.ecosystemWanted[i] {
						t.Errorf(
							"%v: want: %s, got %s",
							errMappingNotExpected, tt.ecosystemWanted[i], *d.Ecosystem,
						)
					}
				}
			}
		})
	}
}
