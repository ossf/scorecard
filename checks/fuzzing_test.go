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
	"io"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	sce "github.com/ossf/scorecard/v5/errors"
	scut "github.com/ossf/scorecard/v5/utests"
)

// TestFuzzing is a test function for Fuzzing.
func TestFuzzing(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		fileContent string
		langs       []clients.Language
		fileName    []string
		response    clients.SearchResponse
		expected    scut.TestReturn
		wantErr     bool
		wantFuzzErr bool
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
			expected: scut.TestReturn{
				Error:         nil,
				NumberOfWarn:  1,
				NumberOfDebug: 0,
				NumberOfInfo:  0,
				Score:         0,
			},
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
			expected: scut.TestReturn{
				NumberOfWarn:  0,
				NumberOfDebug: 0,
				NumberOfInfo:  1,
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
			expected: scut.TestReturn{
				Error:         nil,
				NumberOfWarn:  1,
				NumberOfDebug: 0,
				NumberOfInfo:  0,
				Score:         0,
			},
		},
		{
			name:        "error",
			wantFuzzErr: true,
			expected: scut.TestReturn{
				Error:         nil,
				NumberOfWarn:  1,
				NumberOfDebug: 0,
				NumberOfInfo:  0,
				Score:         0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockFuzz := mockrepo.NewMockRepoClient(ctrl)
			mockFuzz.EXPECT().URI().Return("github.com/ossf/scorecard").AnyTimes()
			mockFuzz.EXPECT().Search(gomock.Any()).
				DoAndReturn(func(q clients.SearchRequest) (clients.SearchResponse, error) {
					if tt.wantErr {
						return clients.SearchResponse{}, errors.New("error")
					}
					return tt.response, nil
				}).AnyTimes()
			mockFuzz.EXPECT().ListProgrammingLanguages().Return(tt.langs, nil).AnyTimes()
			mockFuzz.EXPECT().ListFiles(gomock.Any()).Return(tt.fileName, nil).AnyTimes()
			mockFuzz.EXPECT().GetFileReader(gomock.Any()).DoAndReturn(func(f string) (io.ReadCloser, error) {
				if tt.wantErr {
					return nil, errors.New("error")
				}
				rc := io.NopCloser(strings.NewReader(tt.fileContent))
				return rc, nil
			}).AnyTimes()
			dl := scut.TestDetailLogger{}
			raw := checker.RawResults{}
			req := checker.CheckRequest{
				RepoClient:  mockFuzz,
				OssFuzzRepo: mockFuzz,
				Dlogger:     &dl,
				RawResults:  &raw,
			}

			if tt.wantFuzzErr {
				req.OssFuzzRepo = nil
			}

			result := Fuzzing(&req)
			if (result.Error != nil) != tt.wantErr {
				t.Errorf("Fuzzing() error = %v, wantErr %v", result.Error, tt.wantErr)
				return
			}

			scut.ValidateTestReturn(t, tt.name, &tt.expected, &result, &dl)
		})
	}
}
