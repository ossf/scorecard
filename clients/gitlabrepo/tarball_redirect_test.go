// Copyright 2021 Security Scorecard Authors
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

package gitlabrepo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

func TestApiFunctionTokenNotForwardedOnRedirect(t *testing.T) {
	const token = "super-secret-pat"
	t.Setenv("GITLAB_AUTH_TOKEN", token)

	var forwarded string
	// Stands in for the redirect target (object storage / CDN / attacker host).
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		forwarded = r.Header.Get("PRIVATE-TOKEN")
		w.WriteHeader(http.StatusOK)
	}))
	defer other.Close()

	// Stands in for the GitLab instance; redirects the archive request elsewhere.
	instance := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, other.URL, http.StatusFound)
	}))
	defer instance.Close()

	tempDir := t.TempDir()
	repoFile, err := os.CreateTemp(tempDir, repoFilename)
	if err != nil {
		t.Fatalf("os.CreateTemp: %v", err)
	}
	defer repoFile.Close()

	handler := tarballHandler{ctx: context.Background(), once: new(sync.Once)}
	if err := handler.apiFunction(instance.URL, tempDir, repoFile); err != nil {
		t.Fatalf("apiFunction: %v", err)
	}

	if forwarded != "" {
		t.Fatalf("PRIVATE-TOKEN forwarded to redirect host: got %q, want it dropped", forwarded)
	}
}
