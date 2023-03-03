// Copyright 2022 OpenSSF Scorecard Authors
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
	"errors"
	"path"
	"testing"

	sclog "github.com/ossf/scorecard/v4/log"
	"github.com/ossf/scorecard/v4/pkg"
)

func Test_initRepoAndClientByChecks(t *testing.T) {
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
				logger:          sclog.NewLogger(sclog.InfoLevel),
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
				logger:    sclog.NewLogger(sclog.InfoLevel),
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

func Test_mapDependencyEcosystemNaming(t *testing.T) {
	//nolint
	tests := []struct {
		name      string
		deps      []dependency
		errWanted error
	}{
		{
			name: "error invalid github ecosystem",
			deps: []dependency{
				{
					Name:      "dependency_1",
					Ecosystem: asPointer("not_supported"),
				},
				{
					Name:      "dependency_2",
					Ecosystem: asPointer("gomod"),
				},
			},
			errWanted: errInvalid,
		},
		{
			name: "error cannot find mapping",
			deps: []dependency{
				{
					Name:      "dependency_3",
					Ecosystem: asPointer("foobar"),
				},
			},
			errWanted: errMappingNotFound,
		},
		{
			name: "correct mapping",
			deps: []dependency{
				{
					Name:      "dependency_4",
					Ecosystem: asPointer("gomod"),
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
					Name:      "dependency_7",
					Ecosystem: asPointer("actions"),
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			//nolint
			err := mapDependencyEcosystemNaming(tt.deps)
			if tt.errWanted != nil && errors.Is(tt.errWanted, err) {
				t.Errorf("not a wanted error, want:%v, got:%v", tt.errWanted, err)
				return
			}
		})
	}
}

func Test_isSpecifiedByUser(t *testing.T) {
	//nolint
	tests := []struct {
		name               string
		ct                 pkg.ChangeType
		changeTypesToCheck []string
		resultWanted       bool
	}{
		{
			name: "error invalid github ecosystem",
		},
		{
			name:               "added",
			ct:                 pkg.ChangeType("added"),
			changeTypesToCheck: nil,
			resultWanted:       false,
		},
		{
			name:               "ct is added but not specified",
			ct:                 pkg.ChangeType("added"),
			changeTypesToCheck: []string{"removed"},
			resultWanted:       false,
		},
		{
			name:               "removed",
			ct:                 pkg.ChangeType("added"),
			changeTypesToCheck: []string{"added", "removed"},
			resultWanted:       true,
		},
		{
			name:               "not_supported",
			ct:                 pkg.ChangeType("not_supported"),
			changeTypesToCheck: []string{"added", "removed"},
			resultWanted:       false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			//nolint
			result := isSpecifiedByUser(tt.ct, tt.changeTypesToCheck)
			if result != tt.resultWanted {
				t.Errorf("result (%v) != result wanted (%v)", result, tt.resultWanted)
				return
			}
		})
	}
}
