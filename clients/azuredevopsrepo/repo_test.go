// Copyright 2024 OpenSSF Scorecard Authors
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

package azuredevopsrepo

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRepo_parse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		inputURL string
		expected Repo
		wantErr  bool
	}{
		{
			name: "valid azuredevops project with scheme",
			expected: Repo{
				scheme:       "https",
				host:         "dev.azure.com",
				organization: "dnceng-public",
				project:      "public",
				name:         "public",
			},
			inputURL: "https://dev.azure.com/dnceng-public/public/_git/public",
			wantErr:  false,
		},
		{
			name: "valid azuredevops project without scheme",
			expected: Repo{
				scheme:       "https",
				host:         "dev.azure.com",
				organization: "dnceng-public",
				project:      "public",
				name:         "public",
			},
			inputURL: "dev.azure.com/dnceng-public/public/_git/public",
			wantErr:  false,
		},
		{
			name: "valid azuredevops project with user information",
			expected: Repo{
				scheme:       "https",
				host:         "dev.azure.com",
				organization: "dnceng-public",
				project:      "public",
				name:         "public",
			},
			inputURL: "https://dnceng-public@dev.azure.com/dnceng-public/public/_git/public?version=1#readme",
			wantErr:  false,
		},
		{
			name: "valid legacy project with scheme",
			expected: Repo{
				scheme:       "https",
				host:         "dev.azure.com",
				organization: "dnceng-public",
				project:      "public",
				name:         "public",
			},
			inputURL: "https://dnceng-public.visualstudio.com/public/_git/public",
			wantErr:  false,
		},
		{
			name: "valid legacy project without scheme",
			expected: Repo{
				scheme:       "https",
				host:         "dev.azure.com",
				organization: "dnceng-public",
				project:      "public",
				name:         "public",
			},
			inputURL: "dnceng-public.visualstudio.com/public/_git/public",
			wantErr:  false,
		},
		{
			name: "valid legacy DefaultCollection project with scheme",
			expected: Repo{
				scheme:       "https",
				host:         "dev.azure.com",
				organization: "dnceng-public",
				project:      "public",
				name:         "public",
			},
			inputURL: "https://dnceng-public.visualstudio.com/DefaultCollection/public/_git/public",
			wantErr:  false,
		},
		{
			name: "valid legacy DefaultCollection project without scheme",
			expected: Repo{
				scheme:       "https",
				host:         "dev.azure.com",
				organization: "dnceng-public",
				project:      "public",
				name:         "public",
			},
			inputURL: "dnceng-public.visualstudio.com/DefaultCollection/public/_git/public",
			wantErr:  false,
		},
		{
			name: "valid legacy project with case-insensitive markers",
			expected: Repo{
				scheme:       "https",
				host:         "dev.azure.com",
				organization: "dnceng-public",
				project:      "public",
				name:         "public",
			},
			inputURL: "https://DNCENG-PUBLIC.visualstudio.com/defaultcollection/public/_GIT/public",
			wantErr:  false,
		},
		{
			name:     "invalid azuredevops project missing repo",
			expected: Repo{},
			inputURL: "https://dev.azure.com/dnceng-public/public",
			wantErr:  true,
		},
		{
			name:     "invalid host",
			expected: Repo{},
			inputURL: "https://example.com/dnceng-public/public/_git/public",
			wantErr:  true,
		},
		{
			name:     "invalid nested legacy host",
			expected: Repo{},
			inputURL: "https://nested.dnceng-public.visualstudio.com/public/_git/public",
			wantErr:  true,
		},
		{
			name:     "invalid HTTP URL",
			expected: Repo{},
			inputURL: "http://dev.azure.com/dnceng-public/public/_git/public",
			wantErr:  true,
		},
		{
			name:     "invalid URL with port",
			expected: Repo{},
			inputURL: "https://dev.azure.com:8443/dnceng-public/public/_git/public",
			wantErr:  true,
		},
		{
			name:     "invalid git path segment",
			expected: Repo{},
			inputURL: "https://dev.azure.com/dnceng-public/public/git/public",
			wantErr:  true,
		},
		{
			name:     "invalid legacy project with extra path segment",
			expected: Repo{},
			inputURL: "https://dnceng-public.visualstudio.com/public/_git/public/extra",
			wantErr:  true,
		},
		{
			name:     "invalid DefaultCollection path",
			expected: Repo{},
			inputURL: "https://dnceng-public.visualstudio.com/DefaultCollection/public/public",
			wantErr:  true,
		},
		{
			name:     "invalid encoded path separator",
			expected: Repo{},
			inputURL: "https://dev.azure.com/dnceng-public%2Fpublic/_git/public",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := Repo{}
			if err := r.parse(tt.inputURL); (err != nil) != tt.wantErr {
				t.Errorf("repoURL.parse() error = %v", err)
			}
			if tt.wantErr {
				return
			}
			if diff := cmp.Diff(tt.expected, r, cmp.AllowUnexported(Repo{})); diff != "" {
				t.Errorf("Repo mismatch (-want +got):\n%s", diff)
			}
			if got, want := r.Host(), tt.expected.Host(); got != want {
				t.Errorf("Host() = %q, want %q", got, want)
			}
			if got, want := r.Path(), tt.expected.Path(); got != want {
				t.Errorf("Path() = %q, want %q", got, want)
			}
			if got, want := r.URI(), tt.expected.URI(); got != want {
				t.Errorf("URI() = %q, want %q", got, want)
			}
			if got, want := r.String(), tt.expected.String(); got != want {
				t.Errorf("String() = %q, want %q", got, want)
			}
		})
	}
}

func TestHasAzureDevOpsHost(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "canonical host",
			input: "https://dev.azure.com/dnceng-public/public/_git/public",
			want:  true,
		},
		{
			name:  "legacy host",
			input: "dnceng-public.visualstudio.com/public/_git/public",
			want:  true,
		},
		{
			name:  "nested legacy host remains Azure-associated",
			input: "https://nested.dnceng-public.visualstudio.com/public/_git/public",
			want:  true,
		},
		{
			name:  "unsupported host",
			input: "https://gitlab.com/ossf/scorecard",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := HasAzureDevOpsHost(tt.input); got != tt.want {
				t.Errorf("HasAzureDevOpsHost() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestRepo_IsValid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		inputURL     string
		repo         Repo
		wantErr      bool
		flagRequired bool
	}{
		{
			name: "valid azuredevops project",
			repo: Repo{
				host:         "dev.azure.com",
				organization: "dnceng-public",
				project:      "public",
				name:         "public",
			},
			wantErr: false,
		},
		{
			name: "invalid azuredevops project",
			repo: Repo{
				host:         "dev.azure.com",
				organization: "dnceng-public",
				project:      "public",
				name:         "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := tt.repo.IsValid(); (err != nil) != tt.wantErr {
				t.Errorf("repoURL.IsValid() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
		})
	}
}

func TestRepo_MakeAzureDevOpsRepo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		repouri  string
		expected bool
	}{
		{
			repouri:  "github.com/ossf/scorecard",
			expected: false,
		},
		{
			repouri:  "ossf/scorecard",
			expected: false,
		},
		{
			repouri:  "https://github.com/ossf/scorecard",
			expected: false,
		},
		{
			repouri:  "https://dev.azure.com/dnceng-public/public/_git/public",
			expected: true,
		},
		{
			repouri:  "dev.azure.com/dnceng-public/public/_git/public",
			expected: true,
		},
		{
			repouri:  "https://dnceng-public.visualstudio.com/public/_git/public",
			expected: true,
		},
		{
			repouri:  "dnceng-public.visualstudio.com/DefaultCollection/public/_git/public",
			expected: true,
		},
		{
			repouri:  "https://example.com/dnceng-public/public/_git/public",
			expected: false,
		},
	}

	for _, tt := range tests {
		g, err := MakeAzureDevOpsRepo(tt.repouri)
		if (g != nil) != (err == nil) {
			t.Errorf("got azuredevopsrepo: %s with err %s", g, err)
		}
		isAzureDevOps := g != nil && err == nil
		if isAzureDevOps != tt.expected {
			t.Errorf("got %s isazuredevops: %t expected %t", tt.repouri, isAzureDevOps, tt.expected)
		}
	}
}
