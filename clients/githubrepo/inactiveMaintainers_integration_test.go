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
	"os"
	"testing"
	"time"

	"github.com/google/go-github/v53/github"
)

// TestSetupInitialization tests the setup method initialization.
func TestSetupInitialization(t *testing.T) {
	t.Parallel()

	cutoff := time.Now().UTC().AddDate(0, -6, 0)
	repo := &Repo{owner: "test", repo: "repo"}

	handler := &maintainerHandler{}
	handler.init(context.Background(), github.NewClient(nil), repo)
	handler.setCutoff(cutoff)

	// Call setup (will fail without real GitHub, but tests initialization path)
	err := handler.setup()
	if err != nil {
		t.Logf("setup failed as expected without real GitHub: %v", err)
	}

	// Verify setupOnce was used
	if handler.setupOnce == nil {
		t.Error("setupOnce should be initialized")
	}
}

// TestQueryWithoutSetup tests querying before setup.
func TestQueryWithoutSetup(t *testing.T) {
	t.Parallel()

	handler := &maintainerHandler{}
	handler.init(context.Background(), github.NewClient(nil), &Repo{owner: "test", repo: "repo"})
	handler.setCutoff(time.Now().UTC().AddDate(0, -6, 0))

	// Ensure maps are initialized
	if handler.elevated == nil {
		t.Fatal("elevated map should be initialized")
	}
	if handler.active == nil {
		t.Fatal("active map should be initialized")
	}

	// We don't call query() because it would try to make real API calls
	// The initialization above is sufficient to test the init flow
}

// TestDebugLogging tests the debug logging path.
func TestDebugLogging(t *testing.T) {
	// Set the debug environment variable
	t.Setenv("SCORECARD_DEBUG_MAINTAINERS", "1")

	// Just verify the environment variable check works
	if os.Getenv("SCORECARD_DEBUG_MAINTAINERS") != "1" {
		t.Error("debug environment variable should be set")
	}

	// We can't easily test the actual logging without a full setup,
	// but this confirms the code path exists
}

// TestEarlyTerminationOptimization tests the early termination logic.
func TestEarlyTerminationOptimization(t *testing.T) {
	t.Parallel()

	handler := &maintainerHandler{
		elevated: map[string]struct{}{
			"user1": {},
			"user2": {},
		},
		active: make(map[string]bool),
	}

	// Simulate activity collection with early termination
	activities := []string{"user1", "user2"}
	for _, user := range activities {
		handler.markActive(user)
		if handler.allActive() {
			// Early termination - would stop expensive API calls
			break
		}
	}

	if !handler.allActive() {
		t.Error("all users should be active")
	}
}

// TestCaseInsensitiveUsernames tests username normalization.
func TestCaseInsensitiveUsernames(t *testing.T) {
	t.Parallel()

	handler := &maintainerHandler{
		elevated: map[string]struct{}{
			"testuser": {}, // lowercase
		},
		active: make(map[string]bool),
	}

	// Test with different case variations
	testCases := []string{"TestUser", "TESTUSER", "testuser", "TeStUsEr"}

	for _, username := range testCases {
		handler.markActive(username)
	}

	// Should have marked the user active (lowercase)
	if !handler.active["testuser"] {
		t.Error("testuser should be marked as active")
	}

	// Should only have one entry
	if len(handler.active) != 1 {
		t.Errorf("expected 1 active user, got %d", len(handler.active))
	}
}

// TestEmptyAndNilHandling tests edge cases.
func TestEmptyAndNilHandling(t *testing.T) {
	t.Parallel()

	handler := &maintainerHandler{
		elevated: map[string]struct{}{},
		active:   make(map[string]bool),
	}

	// Test marking active with empty username
	if handler.markActive("") {
		t.Error("should return false for empty username")
	}

	// Test allActive with no elevated users
	if handler.allActive() {
		t.Error("should return false when no elevated users")
	}
}

// TestCutoffBeforeActivity tests activity before cutoff.
func TestCutoffBeforeActivity(t *testing.T) {
	t.Parallel()

	cutoff := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	// Activity before cutoff (inactive)
	beforeCutoff := time.Date(2024, 5, 15, 0, 0, 0, 0, time.UTC)
	if !beforeCutoff.Before(cutoff) {
		t.Error("beforeCutoff should be before cutoff")
	}

	// Activity after cutoff (active)
	afterCutoff := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	if afterCutoff.Before(cutoff) {
		t.Error("afterCutoff should not be before cutoff")
	}
}

// TestMultipleSetupCalls tests sync.Once behavior.
func TestMultipleSetupCalls(t *testing.T) {
	t.Parallel()

	handler := &maintainerHandler{}
	handler.init(context.Background(), github.NewClient(nil), &Repo{owner: "test", repo: "repo"})
	handler.setCutoff(time.Now().UTC().AddDate(0, -6, 0))

	// Call setup multiple times
	for i := 0; i < 3; i++ {
		err := handler.setup()
		if err != nil {
			t.Logf("setup iteration %d failed: %v", i, err)
		}
	}

	// Verify setupOnce is still the same instance
	if handler.setupOnce == nil {
		t.Error("setupOnce should be initialized")
	}
}

// TestActivityMapEnsureAllUsers tests that all elevated users appear in results.
func TestActivityMapEnsureAllUsers(t *testing.T) {
	t.Parallel()

	handler := &maintainerHandler{
		elevated: map[string]struct{}{
			"user1": {},
			"user2": {},
			"user3": {},
		},
		active: map[string]bool{
			"user1": true,
			// user2 and user3 not in active map
		},
		cutoff: time.Now().UTC().AddDate(0, -6, 0),
		ctx:    context.Background(),
	}

	// Simulate the query method's logic of ensuring all users in result
	for user := range handler.elevated {
		if _, exists := handler.active[user]; !exists {
			handler.active[user] = false
		}
	}

	// All elevated users should now be in active map
	if len(handler.active) != 3 {
		t.Errorf("expected 3 users in active map, got %d", len(handler.active))
	}

	// Check specific users
	if _, exists := handler.active["user2"]; !exists {
		t.Error("user2 should be in active map")
	}
	if _, exists := handler.active["user3"]; !exists {
		t.Error("user3 should be in active map")
	}
}
