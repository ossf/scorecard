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

	"github.com/ossf/scorecard/v4/checks"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/log"
	"github.com/ossf/scorecard/v4/pkg"
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
				baseSHA:   "base",
				headSHA:   clients.HeadSHA,
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
		name                                       string
		dCtx                                       dependencydiffContext
		wantGhRepo, wantRepoClient, wantFuzzClient bool
		wantVulnClient, wantCIIClient              bool
		wantErr                                    bool
	}{
		{
			name: "error creating repo",
			dCtx: dependencydiffContext{
				logger:          log.NewLogger(log.InfoLevel),
				ctx:             context.Background(),
				ownerName:       path.Join("host_not_exist.com", "owner_not_exist"),
				repoName:        "repo_not_exist",
				checkNamesToRun: []string{},
			},
			wantGhRepo:     false,
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
			err := initRepoAndClientByChecks(&tt.dCtx)
			if (err != nil) != tt.wantErr {
				t.Errorf("initRepoAndClientByChecks() error = {%v}, want error: %v", err, tt.wantErr)
				return
			}
			if (tt.dCtx.ghRepo != nil) != tt.wantGhRepo {
				t.Errorf("init repo error, wantGhRepo: %v, got %v", tt.wantGhRepo, tt.dCtx.ghRepo)
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
		name           string
		dCtx           dependencydiffContext
		wantResultsLen int
		wantErr        bool
	}{
		{
			name: "empty response",
			dCtx: dependencydiffContext{
				ctx:       context.Background(),
				logger:    log.NewLogger(log.InfoLevel),
				ownerName: "ossf-tests",
				repoName:  "scorecard-depdiff",
			},
			wantResultsLen: 0,
		},
		{
			name: "normal response 1",
			dCtx: dependencydiffContext{
				ctx:       context.Background(),
				logger:    log.NewLogger(log.InfoLevel),
				ownerName: "ossf-tests",
				repoName:  "scorecard-depdiff",
				checkNamesToRun: []string{
					checks.CheckBranchProtection,
				},
				dependencydiffs: []dependency{
					{
						ChangeType:       (*pkg.ChangeType)(asPointer(string(pkg.Added))),
						Name:             "tensorflow",
						SourceRepository: asPointer("https://github.com/tensorflow/tensorflow"),
					},
					{
						ChangeType:       (*pkg.ChangeType)(asPointer(string(pkg.Updated))),
						Name:             "numpy",
						SourceRepository: nil, // To make it explicit here, this one doesn't have a valid srcRepo URL.
					},
					{
						ChangeType:       (*pkg.ChangeType)(asPointer(string(pkg.Removed))),
						Name:             "pytorch",
						SourceRepository: asPointer("https://github.com/pytorch/pytorch"),
					},
				},
			},
			wantResultsLen: 3, // we need 3 results since the second one without a srcRepo URL won't break down the entire func.
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := initRepoAndClientByChecks(&tt.dCtx)
			if err != nil {
				t.Errorf("init repo and client error")
				return
			}
			err = getScorecardCheckResults(&tt.dCtx)
			if (err != nil) != tt.wantErr {
				t.Errorf("getScorecardCheckResults() error = {%v}, want error: %v", err, tt.wantErr)
				return
			}
			lenResults := len(tt.dCtx.results)
			if lenResults != tt.wantResultsLen {
				t.Errorf("want empty results: %v, got len of results:%d", tt.wantResultsLen, lenResults)
				return
			}
		})
	}
}

func asPointer(s string) *string {
	return &s
}
