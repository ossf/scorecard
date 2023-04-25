// Copyright 2020 OpenSSF Scorecard Authors
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

package checks

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
	sce "github.com/ossf/scorecard/v4/errors"
	scut "github.com/ossf/scorecard/v4/utests"
)

// TestFuzzing is a test function for Fuzzing.
func TestFuzzing(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name        string
		want        checker.CheckResult
		langs       []clients.Language
		response    clients.SearchResponse
		wantErr     bool
		wantFuzzErr bool
		fileName    []string
		fileContent string
		expected    scut.TestReturn
	}{
		{
			name:     "empty response",
			response: clients.SearchResponse{},
			langs: []clients.Language{
				{
					Name:     clients.Go,
					NumLines: 300,
				},
			},
			wantErr: false,
		},
		{
			name: "hits 1",
			response: clients.SearchResponse{
				Hits: 1,
			},
			langs: []clients.Language{
				{
					Name:     clients.Go,
					NumLines: 100,
				},
				{
					Name:     clients.Java,
					NumLines: 70,
				},
			},
			wantErr: false,
			want:    checker.CheckResult{Score: 10},
			expected: scut.TestReturn{
				NumberOfWarn:  0,
				NumberOfDebug: 0,
				NumberOfInfo:  0,
				Score:         10,
			},
		},
		{
			name: "nil response",
			langs: []clients.Language{
				{
					Name:     clients.Python,
					NumLines: 256,
				},
			},
			wantErr: true,
			want:    checker.CheckResult{Score: -1},
			expected: scut.TestReturn{
				Error:         sce.ErrScorecardInternal,
				NumberOfWarn:  0,
				NumberOfDebug: 0,
				NumberOfInfo:  0,
				Score:         -1,
			},
		},
		{
			name: "min score since lang not supported",
			langs: []clients.Language{
				{
					Name:     clients.LanguageName("a_not_supported_lang"),
					NumLines: 500,
				},
			},
			wantFuzzErr: false,
			want:        checker.CheckResult{Score: 0},
		},
		{
			name:        "error",
			wantFuzzErr: true,
			want:        checker.CheckResult{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockFuzz := mockrepo.NewMockRepoClient(ctrl)
			mockFuzz.EXPECT().URI().Return("github.com/ossf/scorecard").AnyTimes()
			mockFuzz.EXPECT().Search(gomock.Any()).
				DoAndReturn(func(q clients.SearchRequest) (clients.SearchResponse, error) {
					if tt.wantErr {
						//nolint
						return clients.SearchResponse{}, errors.New("error")
					}
					return tt.response, nil
				}).AnyTimes()
			mockFuzz.EXPECT().ListProgrammingLanguages().Return(tt.langs, nil).AnyTimes()
			mockFuzz.EXPECT().ListFiles(gomock.Any()).Return(tt.fileName, nil).AnyTimes()
			mockFuzz.EXPECT().GetFileContent(gomock.Any()).DoAndReturn(func(f string) (string, error) {
				if tt.wantErr {
					//nolint
					return "", errors.New("error")
				}
				return tt.fileContent, nil
			}).AnyTimes()
			dl := scut.TestDetailLogger{}
			req := checker.CheckRequest{
				RepoClient:  mockFuzz,
				OssFuzzRepo: mockFuzz,
				Dlogger:     &dl,
			}
			if tt.wantFuzzErr {
				req.OssFuzzRepo = nil
			}

			result := Fuzzing(&req)
			if (result.Error != nil) != tt.wantErr {
				t.Errorf("Fuzzing() error = %v, wantErr %v", result.Error, tt.wantErr)
				return
			}

			if !scut.ValidateTestReturn(t, tt.name, &tt.expected, &result, &dl) {
				t.Fatalf(tt.name, tt.expected)
			}
		})
	}
}
