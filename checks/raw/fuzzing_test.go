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
	"path"
	"regexp"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
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

func Test_fuzzFileAndFuncMatchPattern(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name              string
		expectedFileMatch bool
		expectedFuncMatch bool
		lang              clients.LanguageName
		fileName          string
		fileContent       string
		wantErr           bool
	}{
		{
			name:              "Test_fuzzFuncRegex file success & func success",
			expectedFileMatch: true,
			expectedFuncMatch: true,
			lang:              clients.LanguageName("go"),
			fileName:          "FOOoo_fOOff_BaRRR_test.go",
			fileContent:       `func FuzzSomething (fOo_bAR_1234 *testing.F)`,
			wantErr:           false,
		},
		{
			name:              "Test_fuzzFuncRegex file success & func failure",
			expectedFileMatch: true,
			expectedFuncMatch: false,
			lang:              clients.LanguageName("go"),
			fileName:          "a_unit_test.go",
			fileContent:       `func TestSomethingUnitTest (t *testing.T)`,
			wantErr:           true,
		},
		{
			name:              "Test_fuzzFuncRegex file failure & func failure",
			expectedFileMatch: false,
			expectedFuncMatch: false,
			lang:              clients.LanguageName("go"),
			fileName:          "not_a_fuzz_test_file.go",
			fileContent:       `func main (t *testing.T)`,
			wantErr:           true,
		},
		{
			name:              "Test_fuzzFuncRegex not a support language",
			expectedFileMatch: false,
			expectedFuncMatch: false,
			lang:              clients.LanguageName("not_a_supported_one"),
			fileName:          "a_fuzz_test.py",
			fileContent:       `def NotSupported (foo)`,
			wantErr:           true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			langSpecs, ok := languageFuzzSpecs[tt.lang]
			if !ok && !tt.wantErr {
				t.Errorf("retrieve supported language error")
			}
			fileMatchPattern := langSpecs.filePattern
			fileMatch, err := path.Match(fileMatchPattern, tt.fileName)
			if (fileMatch != tt.expectedFileMatch || err != nil) && !tt.wantErr {
				t.Errorf("fileMatch = %v, want %v for %v", fileMatch, tt.expectedFileMatch, tt.name)
			}
			funcRegexPattern := langSpecs.funcPattern
			r := regexp.MustCompile(funcRegexPattern)
			found := r.MatchString(tt.fileContent)
			if (found != tt.expectedFuncMatch) && !tt.wantErr {
				t.Errorf("funcMatch = %v, want %v for %v", fileMatch, tt.expectedFileMatch, tt.name)
			}
		})
	}
}

func Test_checkFuzzFunc(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name        string
		want        bool
		wantErr     bool
		langs       []clients.Language
		fileName    []string
		fileContent string
	}{
		{
			name:    "Test_checkFuzzFunc failure",
			want:    false,
			wantErr: false,
			fileName: []string{
				"foo_test.go",
				"main.go",
			},
			langs: []clients.Language{
				{
					Name:     clients.Go,
					NumLines: 100,
				},
			},
			fileContent: "func TestFoo (t *testing.T)",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockClient := mockrepo.NewMockRepoClient(ctrl)
			mockClient.EXPECT().ListFiles(gomock.Any()).Return(tt.fileName, nil).AnyTimes()
			mockClient.EXPECT().GetFileContent(gomock.Any()).DoAndReturn(func(f string) (string, error) {
				if tt.wantErr {
					//nolint
					return "", errors.New("error")
				}
				return tt.fileContent, nil
			}).AnyTimes()
			req := checker.CheckRequest{
				RepoClient: mockClient,
			}
			for _, l := range tt.langs {
				got, _, err := checkFuzzFunc(&req, l.Name)
				if (got != tt.want || err != nil) && !tt.wantErr {
					t.Errorf("checkFuzzFunc() = %v, want %v for %v", got, tt.want, tt.name)
				}
			}
		})
	}
}

func Test_getProminentLanguages(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name      string
		languages []clients.Language
		expected  []clients.LanguageName
	}{
		{
			name: "case1",
			languages: []clients.Language{
				{
					Name:     clients.Go,
					NumLines: 1000,
				},
				{
					Name:     clients.Python,
					NumLines: 40,
				}, {
					Name:     clients.JavaScript,
					NumLines: 800,
				},
			},
			expected: []clients.LanguageName{
				clients.Go, clients.JavaScript,
			},
		},
		{
			// This test case simulates the situation when the GitHub language API returns
			// duplicated languages, but we can still drop them and get the correct result.
			name: "case2: drop duplicates",
			languages: []clients.Language{
				{
					Name:     clients.Go,
					NumLines: 1000,
				},
				{
					Name:     clients.Python,
					NumLines: 40,
				}, {
					Name:     clients.JavaScript,
					NumLines: 800,
				},
				{
					Name:     clients.Go,
					NumLines: 1000,
				},
				{
					Name:     clients.Python,
					NumLines: 40,
				}, {
					Name:     clients.JavaScript,
					NumLines: 800,
				},
				{
					Name:     clients.Go,
					NumLines: 1000,
				},
				{
					Name:     clients.Python,
					NumLines: 40,
				}, {
					Name:     clients.JavaScript,
					NumLines: 800,
				},
			},
			expected: []clients.LanguageName{
				clients.Go, clients.JavaScript,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := getProminentLanguages(tt.languages)
			if !unorderedEqual(got, tt.expected) {
				t.Errorf(
					"got (%s) != expected (%s)",
					got, tt.expected,
				)
			}
		})
	}
}

func unorderedEqual(l1, l2 []clients.LanguageName) bool {
	if len(l1) != len(l2) {
		return false
	}
	l1Map, l2Map := map[clients.LanguageName]bool{}, map[clients.LanguageName]bool{}
	for _, l := range l1 {
		l1Map[l] = true
	}
	for _, l := range l2 {
		l2Map[l] = true
		if !l1Map[l] {
			return false
		}
	}
	for k := range l1Map {
		if !l2Map[k] {
			return false
		}
	}
	return true
}
