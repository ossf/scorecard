// Copyright 2026 OpenSSF Scorecard Authors
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

package cmd

import (
	"testing"

	"github.com/ossf/scorecard/v5/clients"
)

func TestMakeRepoLegacyAzureDevOpsURL(t *testing.T) {
	t.Setenv("SCORECARD_EXPERIMENTAL", "1")

	repo, err := makeRepo(
		"https://dnceng-public.visualstudio.com/DefaultCollection/public/_git/public",
	)
	if err != nil {
		t.Fatalf("makeRepo() error = %v", err)
	}
	if got, want := repo.Type(), clients.RepoTypeAzureDevOps; got != want {
		t.Errorf("Type() = %q, want %q", got, want)
	}
	if got, want := repo.URI(), "dev.azure.com/dnceng-public/public/_git/public"; got != want {
		t.Errorf("URI() = %q, want %q", got, want)
	}
}

func TestMakeRepoRejectsMalformedAzureDevOpsURL(t *testing.T) {
	t.Setenv("SCORECARD_EXPERIMENTAL", "1")

	tests := []string{
		"http://teamgitlab.visualstudio.com/project/_git/repo",
		"https://teamgitlab.visualstudio.com:8443/project/_git/repo",
	}
	for _, repoURI := range tests {
		if repo, err := makeRepo(repoURI); err == nil || repo != nil {
			t.Errorf("makeRepo(%q) = %v, %v; want nil repo and error", repoURI, repo, err)
		}
	}
}
