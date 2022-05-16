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

package raw

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
	scut "github.com/ossf/scorecard/v4/utests"
)

// Test_checkOSSFuzz is a test function for checkOSSFuzz.
func Test_checkOSSFuzz(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name        string
		want        bool
		response    clients.SearchResponse
		wantErr     bool
		wantFuzzErr bool
	}{
		{
			name:     "Test_checkOSSFuzz failure",
			want:     false,
			response: clients.SearchResponse{},
			wantErr:  false,
		},
		{
			name: "Test_checkOSSFuzz success",
			want: true,
			response: clients.SearchResponse{
				Hits: 1,
			},
			wantErr: false,
		},
		{
			name:    "error",
			want:    false,
			wantErr: true,
		},
		{
			name:        "Test_checkOSSFuzz fuzz error",
			want:        false,
			wantFuzzErr: true,
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

			req := checker.CheckRequest{
				RepoClient:  mockFuzz,
				OssFuzzRepo: mockFuzz,
			}
			if tt.wantFuzzErr {
				req.OssFuzzRepo = nil
			}

			got, err := checkOSSFuzz(&req)

			if (err != nil) != tt.wantErr {
				t.Errorf("checkOSSFuzz() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("checkOSSFuzz() = %v, want %v for %v", got, tt.want, tt.name)
			}
		})
	}
}

// Test_checkCFLite is a test function for checkCFLite.
func Test_checkCFLite(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name        string
		want        bool
		wantErr     bool
		fileName    []string
		fileContent string
	}{
		{
			name:        "Test_checkCFLite success",
			want:        false,
			wantErr:     false,
			fileName:    []string{"docker-compose.yml"},
			fileContent: `# .clusterfuzzlite/Dockerfile`,
		},
		{
			name:     "Test_checkCFLite failure",
			want:     false,
			wantErr:  true,
			fileName: []string{"docker-compose.yml"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockFuzz := mockrepo.NewMockRepoClient(ctrl)
			mockFuzz.EXPECT().ListFiles(gomock.Any()).Return(tt.fileName, nil).AnyTimes()
			mockFuzz.EXPECT().GetFileContent(gomock.Any()).DoAndReturn(func(f string) (string, error) {
				if tt.wantErr {
					//nolint
					return "", errors.New("error")
				}
				return tt.fileContent, nil
			}).AnyTimes()
			req := checker.CheckRequest{
				RepoClient: mockFuzz,
			}
			got, err := checkCFLite(&req)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkCFLite() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("checkCFLite() = %v, want %v for test %v", got, tt.want, tt.name)
			}
		})
	}
}
