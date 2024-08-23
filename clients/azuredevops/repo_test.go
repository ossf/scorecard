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

package azuredevops

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestRepo_parse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		inputURL     string
		expected     Repo
		wantErr      bool
		flagRequired bool
	}{
		{
			name: "valid azuredevops project with scheme",
			expected: Repo{
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
				host:         "dev.azure.com",
				organization: "dnceng-public",
				project:      "public",
				name:         "public",
			},
			inputURL: "dev.azure.com/dnceng-public/public/_git/public",
			wantErr:  false,
		},
		{
			name:     "invalid azuredevops project missing repo",
			expected: Repo{},
			inputURL: "https://dev.azure.com/dnceng-public/public",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := Repo{}
			if err := r.parse(tt.inputURL); (err != nil) != tt.wantErr {
				t.Errorf("repoURL.parse() error = %v", err)
			}
			if tt.wantErr {
				return
			}
			t.Log(r.URI())
			if !tt.wantErr && !cmp.Equal(tt.expected, r, cmpopts.IgnoreUnexported(Repo{})) {
				t.Logf("expected: %s GOT: %s", tt.expected.host, r.host)
				t.Logf("expected: %s GOT: %s", tt.expected.organization, r.organization)
				t.Logf("expected: %s GOT: %s", tt.expected.project, r.project)
				t.Logf("expected: %s GOT: %s", tt.expected.name, r.name)
				t.Errorf("Got diff: %s", cmp.Diff(tt.expected, r))
			}
			if !cmp.Equal(r.Host(), tt.expected.host) {
				t.Errorf("%s expected host: %s got host %s", tt.inputURL, tt.expected.host, r.Host())
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
		tt := tt
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
		repouri      string
		expected     bool
		flagRequired bool
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
