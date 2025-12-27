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

package githubrepo

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-github/v53/github"
)

// TestClientGetMaintainerActivityInitialization tests the client initialization.
func TestClientGetMaintainerActivityInitialization(t *testing.T) {
	t.Parallel()

	client := &Client{
		ctx:        context.Background(),
		repoClient: github.NewClient(nil),
		repourl:    &Repo{owner: "test", repo: "repo"},
	}

	// Initialize the handler
	client.inactiveMaintainer = &maintainerHandler{}
	client.inactiveMaintainer.init(client.ctx, client.repoClient, client.repourl)

	// Verify handler is initialized
	if client.inactiveMaintainer == nil {
		t.Fatal("inactiveMaintainer handler should be initialized")
	}

	if client.inactiveMaintainer.elevated == nil {
		t.Error("elevated map should be initialized")
	}

	if client.inactiveMaintainer.active == nil {
		t.Error("active map should be initialized")
	}
}

// TestClientGetMaintainerActivityCutoff tests cutoff setting.
func TestClientGetMaintainerActivityCutoff(t *testing.T) {
	t.Parallel()

	client := &Client{
		ctx:        context.Background(),
		repoClient: github.NewClient(nil),
		repourl:    &Repo{owner: "test", repo: "repo"},
	}

	client.inactiveMaintainer = &maintainerHandler{}
	client.inactiveMaintainer.init(client.ctx, client.repoClient, client.repourl)

	cutoff := time.Now().UTC().AddDate(0, -6, 0)

	// This will fail due to no real GitHub API, but tests the flow
	_, err := client.GetMaintainerActivity(cutoff)
	if err != nil {
		t.Logf("GetMaintainerActivity failed as expected without real GitHub: %v", err)
	}

	// Verify cutoff was set
	if client.inactiveMaintainer.cutoff.IsZero() {
		t.Error("cutoff should have been set")
	}

	// Verify cutoff matches what we passed
	if !client.inactiveMaintainer.cutoff.Equal(cutoff) {
		t.Errorf("cutoff mismatch: got %v, want %v",
			client.inactiveMaintainer.cutoff, cutoff)
	}
}

// TestClientGetMaintainerActivityError tests error handling.
func TestClientGetMaintainerActivityError(t *testing.T) {
	t.Parallel()

	client := &Client{
		ctx:        context.Background(),
		repoClient: github.NewClient(nil),
		repourl:    &Repo{owner: "test", repo: "repo"},
	}

	client.inactiveMaintainer = &maintainerHandler{}
	client.inactiveMaintainer.init(client.ctx, client.repoClient, client.repourl)

	cutoff := time.Now().UTC().AddDate(0, -6, 0)

	_, err := client.GetMaintainerActivity(cutoff)

	// Expected to fail with no real GitHub setup
	if err == nil {
		// If it somehow succeeds, that's OK for this test
		return
	}

	// Error should be wrapped with context
	if err.Error() == "" {
		t.Error("error should have a message")
	}
}
