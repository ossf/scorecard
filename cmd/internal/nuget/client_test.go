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

// Package nuget implements Nuget API client.
package nuget

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"golang.org/x/exp/slices"

	pmc "github.com/ossf/scorecard/v4/cmd/internal/packagemanager"
)

type resultPackagePage struct {
	url      string
	response string
}
type nugetTestArgs struct {
	inputPackageName               string
	expectedPackageName            string
	resultIndex                    string
	resultPackageRegistrationIndex string
	resultPackageSpec              string
	version                        string
	resultPackageRegistrationPages []resultPackagePage
}
type nugetTest struct {
	name    string
	want    string
	args    nugetTestArgs
	wantErr bool
}

func Test_fetchGitRepositoryFromNuget(t *testing.T) {
	t.Parallel()

	tests := []nugetTest{
		{
			name: "find latest version in single page",

			args: nugetTestArgs{
				inputPackageName:               "nuget-package",
				resultIndex:                    "index.json",
				resultPackageRegistrationIndex: "package_registration_index_single.json",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "package_spec.xml",
				version:                        "4.0.1",
			},
			want:    "https://github.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "find by lowercase package name",

			args: nugetTestArgs{
				inputPackageName:               "Nuget-Package",
				expectedPackageName:            "nuget-package",
				resultIndex:                    "index.json",
				resultPackageRegistrationIndex: "package_registration_index_single.json",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "package_spec.xml",
				version:                        "4.0.1",
			},
			want:    "https://github.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "find and remove trailing slash",

			args: nugetTestArgs{
				inputPackageName:               "Nuget-Package",
				expectedPackageName:            "nuget-package",
				resultIndex:                    "index.json",
				resultPackageRegistrationIndex: "package_registration_index_single.json",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "package_spec_trailing_slash.xml",
				version:                        "4.0.1",
			},
			want:    "https://github.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "find and remove git ending",

			args: nugetTestArgs{
				inputPackageName:               "nuget-package",
				resultIndex:                    "index.json",
				resultPackageRegistrationIndex: "package_registration_index_single.json",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "package_spec_git_ending.xml",
				version:                        "4.0.1",
			},
			want:    "https://github.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "find and handle four digit version",

			args: nugetTestArgs{
				inputPackageName:               "nuget-package",
				resultIndex:                    "index.json",
				resultPackageRegistrationIndex: "package_registration_index_four_digit_version.json",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "package_spec_four_digit_version.xml",
				version:                        "1.60.0.2981",
			},
			want:    "https://github.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "skip semver metadata",

			args: nugetTestArgs{
				inputPackageName:               "nuget-package",
				resultIndex:                    "index.json",
				resultPackageRegistrationIndex: "package_registration_index_metadata_version.json",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "package_spec.xml",
				version:                        "4.0.1",
			},
			want:    "https://github.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "skip pre release",

			args: nugetTestArgs{
				inputPackageName:               "nuget-package",
				resultIndex:                    "index.json",
				resultPackageRegistrationIndex: "package_registration_index_pre_release_version.json",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "package_spec.xml",
				version:                        "4.0.1",
			},
			want:    "https://github.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "skip pre release and metadata",

			args: nugetTestArgs{
				inputPackageName:               "nuget-package",
				resultIndex:                    "index.json",
				resultPackageRegistrationIndex: "package_registration_index_pre_release_and_metadata_version.json",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "package_spec.xml",
				version:                        "4.0.1",
			},
			want:    "https://github.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "find in project url if repository missing",

			args: nugetTestArgs{
				inputPackageName:               "nuget-package",
				resultIndex:                    "index.json",
				resultPackageRegistrationIndex: "package_registration_index_single.json",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "package_spec_project_url.xml",
				version:                        "4.0.1",
			},
			want:    "https://github.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "get github project url without git ending",

			args: nugetTestArgs{
				inputPackageName:               "nuget-package",
				resultIndex:                    "index.json",
				resultPackageRegistrationIndex: "package_registration_index_single.json",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "package_spec_project_url_git_ending.xml",
				version:                        "4.0.1",
			},
			want:    "https://github.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "get gitlab project url if repository url missing",

			args: nugetTestArgs{
				inputPackageName:               "nuget-package",
				resultIndex:                    "index.json",
				resultPackageRegistrationIndex: "package_registration_index_single.json",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "package_spec_project_url_gitlab.xml",
				version:                        "4.0.1",
			},
			want:    "https://gitlab.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "error if project url is not gitlab or github",

			args: nugetTestArgs{
				inputPackageName:               "nuget-package",
				resultIndex:                    "index.json",
				resultPackageRegistrationIndex: "package_registration_index_single.json",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "package_spec_project_url_not_supported.xml",
				version:                        "4.0.1",
			},
			want:    "internal error: source repo is not defined for nuget package nuget-package",
			wantErr: true,
		},
		{
			name: "find latest version in first of multiple pages",

			args: nugetTestArgs{
				inputPackageName:               "nuget-package",
				resultIndex:                    "index.json",
				resultPackageRegistrationIndex: "package_registration_index_multiple.json",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "package_spec.xml",
				version:                        "4.0.1",
			},
			want:    "https://github.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "find latest version in first of multiple remote pages",

			args: nugetTestArgs{
				inputPackageName:               "nuget-package",
				resultIndex:                    "index.json",
				resultPackageRegistrationIndex: "package_registration_index_multiple_remote.json",
				resultPackageRegistrationPages: []resultPackagePage{
					{
						url:      "https://api.nuget.org/v3/registration5-semver1/Foo.NET/page1/index.json",
						response: "package_registration_page_one.json",
					},
					{
						url:      "https://api.nuget.org/v3/registration5-semver1/Foo.NET/page2/index.json",
						response: "package_registration_page_two.json",
					},
				},
				resultPackageSpec: "package_spec.xml",
				version:           "4.0.1",
			},
			want:    "https://github.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "find latest version in last of multiple pages",

			args: nugetTestArgs{
				inputPackageName:               "nuget-package",
				resultIndex:                    "index.json",
				resultPackageRegistrationIndex: "package_registration_index_multiple_last.json",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "package_spec.xml",
				version:                        "4.0.1",
			},
			want:    "https://github.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "find latest version in last of remote multiple pages",

			args: nugetTestArgs{
				inputPackageName:               "nuget-package",
				resultIndex:                    "index.json",
				resultPackageRegistrationIndex: "package_registration_index_multiple_remote.json",
				resultPackageRegistrationPages: []resultPackagePage{
					{
						url:      "https://api.nuget.org/v3/registration5-semver1/Foo.NET/page1/index.json",
						response: "package_registration_page_one.json",
					},
					{
						url:      "https://api.nuget.org/v3/registration5-semver1/Foo.NET/page2/index.json",
						response: "package_registration_page_two_not_listed.json",
					},
				},
				resultPackageSpec: "package_spec.xml",
				version:           "3.5.2",
			},
			want:    "https://github.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "find latest version with default listed value true",

			args: nugetTestArgs{
				inputPackageName:               "nuget-package",
				resultIndex:                    "index.json",
				resultPackageRegistrationIndex: "package_registration_index_default_listed_true.json",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "package_spec.xml",
				version:                        "4.0.1",
			},
			want:    "https://github.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "skip not listed versions",

			args: nugetTestArgs{
				inputPackageName:               "nuget-package",
				resultIndex:                    "index.json",
				resultPackageRegistrationIndex: "package_registration_index_with_not_listed.json",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "package_spec.xml",
				version:                        "3.5.8",
			},
			want:    "https://github.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "error if no listed version",

			args: nugetTestArgs{
				inputPackageName:               "nuget-package",
				resultIndex:                    "index.json",
				resultPackageRegistrationIndex: "package_registration_index_all_not_listed.json",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "",
				version:                        "",
			},
			want:    "internal error: failed to get a listed version for package",
			wantErr: true,
		},
		{
			name: "error no index",

			args: nugetTestArgs{
				inputPackageName:               "nuget-package",
				resultIndex:                    "",
				resultPackageRegistrationIndex: "",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "",
			},
			want:    "internal error: failed to get nuget index json: error",
			wantErr: true,
		},
		{
			name: "error bad index",

			args: nugetTestArgs{
				inputPackageName:               "nuget-package",
				resultIndex:                    "text",
				resultPackageRegistrationIndex: "",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "",
			},
			want:    "internal error: failed to parse nuget index json: invalid character 'e' in literal true (expecting 'r')",
			wantErr: true,
		},
		{
			name: "error package registration index",

			args: nugetTestArgs{
				inputPackageName:               "nuget-package",
				resultIndex:                    "index.json",
				resultPackageRegistrationIndex: "",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "",
			},
			want:    "internal error: failed to get nuget package registration index json: error",
			wantErr: true,
		},
		{
			name: "error bad package index",

			args: nugetTestArgs{
				inputPackageName:               "nuget-package",
				resultIndex:                    "index.json",
				resultPackageRegistrationIndex: "text",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "",
			},
			//nolint
			want:    "internal error: failed to parse nuget package registration index json: invalid character 'e' in literal true (expecting 'r')",
			wantErr: true,
		},
		{
			name: "error package registration page",
			args: nugetTestArgs{
				inputPackageName:               "nuget-package",
				resultIndex:                    "index.json",
				resultPackageRegistrationIndex: "package_registration_index_multiple_remote.json",
				resultPackageRegistrationPages: []resultPackagePage{
					{
						url:      "https://api.nuget.org/v3/registration5-semver1/Foo.NET/page1/index.json",
						response: "",
					},
					{
						url:      "https://api.nuget.org/v3/registration5-semver1/Foo.NET/page2/index.json",
						response: "",
					},
				},
				resultPackageSpec: "",
			},
			want:    "internal error: failed to get nuget package registration page: error",
			wantErr: true,
		},
		{
			name: "error in package spec",
			args: nugetTestArgs{
				inputPackageName:               "nuget-package",
				resultIndex:                    "index.json",
				resultPackageRegistrationIndex: "package_registration_index_single.json",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "",
				version:                        "4.0.1",
			},
			want:    "internal error: failed to get nuget package spec: error",
			wantErr: true,
		},
		{
			name: "error bad package spec",

			args: nugetTestArgs{
				inputPackageName:               "nuget-package",
				resultIndex:                    "index.json",
				resultPackageRegistrationIndex: "package_registration_index_multiple_remote.json",
				resultPackageRegistrationPages: []resultPackagePage{
					{
						url:      "https://api.nuget.org/v3/registration5-semver1/Foo.NET/page2/index.json",
						response: "text",
					},
				},
				resultPackageSpec: "",
			},
			//nolint
			want:    "internal error: failed to parse nuget package registration page: invalid character 'e' in literal true (expecting 'r')",
			wantErr: true,
		},
		{
			name: "error package spec",
			args: nugetTestArgs{
				inputPackageName:               "nuget-package",
				resultIndex:                    "index.json",
				resultPackageRegistrationIndex: "package_registration_index_single.json",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "text",
				version:                        "4.0.1",
			},
			want:    "internal error: failed to parse nuget package spec: EOF",
			wantErr: true,
		},
		{
			name: "bad remote package page",

			args: nugetTestArgs{
				inputPackageName:               "nuget-package",
				resultIndex:                    "index.json",
				resultPackageRegistrationIndex: "package_registration_index_multiple_remote.json",
				resultPackageRegistrationPages: []resultPackagePage{
					{
						url:      "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/1",
						response: "text",
					},
				},
				resultPackageSpec: "",
			},
			want:    "internal error: failed to get nuget package registration page: error",
			wantErr: true,
		},
		{
			name: "error no registration url",
			args: nugetTestArgs{
				inputPackageName:               "nuget-package",
				resultIndex:                    "index_bad_registration_base.json",
				resultPackageRegistrationIndex: "",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "",
				version:                        "4.0.1",
			},
			want:    "internal error: failed to find RegistrationsBaseUrl/3.6.0 URI at nuget index json",
			wantErr: true,
		},
		{
			name: "error no package base url",
			args: nugetTestArgs{
				inputPackageName:               "nuget-package",
				resultIndex:                    "index_bad_package_base.json",
				resultPackageRegistrationIndex: "",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "",
				version:                        "4.0.1",
			},
			want:    "internal error: failed to find PackageBaseAddress/3.0.0 URI at nuget index json",
			wantErr: true,
		},
		{
			name: "error marhsal entry",
			args: nugetTestArgs{
				inputPackageName:               "nuget-package",
				resultIndex:                    "index.json",
				resultPackageRegistrationIndex: "package_registration_index_marshal_error.json",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "",
				version:                        "",
			},
			//nolint
			want:    "internal error: failed to parse nuget package registration index json: failed to unmarshal json: json: cannot unmarshal number into Go struct field Alias.listed of type bool",
			wantErr: true,
		},
		{
			name: "empty package spec",
			args: nugetTestArgs{
				inputPackageName:               "nuget-package",
				resultIndex:                    "index.json",
				resultPackageRegistrationIndex: "package_registration_index_single.json",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "package_spec_error.xml",
				version:                        "4.0.1",
			},
			want:    "internal error: source repo is not defined for nuget package nuget-package",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			p := pmc.NewMockClient(ctrl)
			p.EXPECT().GetURI(gomock.Any()).
				DoAndReturn(func(url string) (*http.Response, error) {
					return nugetIndexOrPageTestResults(url, &tt)
				}).AnyTimes()
			expectedPackageName := tt.args.expectedPackageName
			if strings.TrimSpace(expectedPackageName) == "" {
				expectedPackageName = tt.args.inputPackageName
			}

			p.EXPECT().Get(gomock.Any(), expectedPackageName).
				DoAndReturn(func(url, inputPackageName string) (*http.Response, error) {
					return nugetPackageIndexAndSpecResponse(t, url, &tt)
				}).AnyTimes()
			client := NugetClient{Manager: p}
			got, err := client.GitRepositoryByPackageName(tt.args.inputPackageName)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("fetchGitRepositoryFromNuget() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if err.Error() != tt.want {
					t.Errorf("fetchGitRepositoryFromNuget() err.Error() = %v, wanted %v", err.Error(), tt.want)
					return
				}
				return
			}

			if got != tt.want {
				t.Errorf("fetchGitRepositoryFromNuget() = %v, want %v", got, tt.want)
			}
		})
	}
}

func nugetIndexOrPageTestResults(url string, test *nugetTest) (*http.Response, error) {
	if url == "https://api.nuget.org/v3/index.json" {
		return testResult(test.wantErr, test.args.resultIndex)
	}
	urlResponseIndex := slices.IndexFunc(test.args.resultPackageRegistrationPages,
		func(page resultPackagePage) bool { return page.url == url })
	if urlResponseIndex == -1 {
		//nolint
		return nil, errors.New("error")
	}
	page := test.args.resultPackageRegistrationPages[urlResponseIndex]
	return testResult(test.wantErr, page.response)
}

func nugetPackageIndexAndSpecResponse(t *testing.T, url string, test *nugetTest) (*http.Response, error) {
	t.Helper()
	if strings.HasSuffix(url, "index.json") {
		return testResult(test.wantErr, test.args.resultPackageRegistrationIndex)
	} else if strings.HasSuffix(url, ".nuspec") {
		if strings.Contains(url, fmt.Sprintf("/%v/", test.args.version)) {
			return testResult(test.wantErr, test.args.resultPackageSpec)
		}
		t.Errorf("fetchGitRepositoryFromNuget() version = %v, expected version = %v", url, test.args.version)
	}
	//nolint
	return nil, errors.New("error")
}

func testResult(wantErr bool, responseFileName string) (*http.Response, error) {
	if wantErr && responseFileName == "" {
		//nolint
		return nil, errors.New("error")
	}
	if wantErr && responseFileName == "text" {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString("text")),
		}, nil
	}
	content, err := os.ReadFile("./testdata/" + responseFileName)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(string(content))),
	}, nil
}
