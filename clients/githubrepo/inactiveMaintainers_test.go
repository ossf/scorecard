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
)

// TestInactiveMaintainers_Basic tests basic functionality.
func TestInactiveMaintainers_Basic(t *testing.T) {
	t.Parallel()

	cutoff := time.Now().UTC().AddDate(0, -6, 0)
	repo := &Repo{owner: "test", repo: "repo"}

	handler := &maintainerHandler{}
	handler.init(context.Background(), nil, repo)
	handler.setCutoff(cutoff)

	// Test that elevated and active maps are initialized
	if handler.elevated == nil {
		t.Error("elevated map should be initialized")
	}
	if handler.active == nil {
		t.Error("active map should be initialized")
	}

	// Test cutoff is set correctly
	if handler.cutoff.IsZero() {
		t.Error("cutoff should be set")
	}
}

// TestMarkActive tests markActive.
func TestMarkActive(t *testing.T) {
	t.Parallel()

	handler := &maintainerHandler{
		elevated: map[string]struct{}{
			"user1": {},
			"user2": {},
		},
		active: make(map[string]bool),
	}

	// Test marking elevated user as active returns true
	if !handler.markActive("USER1") {
		t.Error("markActive should return true when marking new user")
	}
	if !handler.active["user1"] {
		t.Error("user1 should be marked as active")
	}

	// Test marking already active user returns false
	if handler.markActive("user1") {
		t.Error("markActive should return false when user is already active")
	}

	// Test non-elevated user returns false
	if handler.markActive("user3") {
		t.Error("markActive should return false for non-elevated user")
	}
	if handler.active["user3"] {
		t.Error("user3 should not be marked as active (not elevated)")
	}

	// Test empty login returns false
	if handler.markActive("") {
		t.Error("markActive should return false for empty login")
	}
	if len(handler.active) != 1 {
		t.Error("empty login should be ignored")
	}
}

// TestAllActive tests allActive.
func TestAllActive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		elevated map[string]struct{}
		active   map[string]bool
		name     string
		want     bool
	}{
		{
			name:     "no maintainers",
			elevated: map[string]struct{}{},
			active:   map[string]bool{},
			want:     false,
		},
		{
			name: "all active",
			elevated: map[string]struct{}{
				"user1": {},
				"user2": {},
			},
			active: map[string]bool{
				"user1": true,
				"user2": true,
			},
			want: true,
		},
		{
			name: "some active",
			elevated: map[string]struct{}{
				"user1": {},
				"user2": {},
			},
			active: map[string]bool{
				"user1": true,
			},
			want: false,
		},
		{
			name: "none active",
			elevated: map[string]struct{}{
				"user1": {},
				"user2": {},
			},
			active: map[string]bool{},
			want:   false,
		},
	}

	for _, tt := range tests {
		// capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			handler := &maintainerHandler{
				elevated: tt.elevated,
				active:   tt.active,
			}
			if got := handler.allActive(); got != tt.want {
				t.Errorf("allActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestMarkActiveEarlyTermination tests early termination support.
func TestMarkActiveEarlyTermination(t *testing.T) {
	t.Parallel()

	handler := &maintainerHandler{
		elevated: map[string]struct{}{
			"user1": {},
			"user2": {},
		},
		active: make(map[string]bool),
	}

	// Mark first user - should return true and allActive should be false
	if !handler.markActive("user1") {
		t.Error("first markActive should return true")
	}
	if handler.allActive() {
		t.Error("allActive should be false with only one of two users active")
	}

	// Mark second user - should return true and allActive should be true
	if !handler.markActive("user2") {
		t.Error("second markActive should return true")
	}
	if !handler.allActive() {
		t.Error("allActive should be true when all users are active")
	}

	// Mark already active user - should return false
	if handler.markActive("user1") {
		t.Error("markActive on already active user should return false")
	}
}

// TestEarlyTerminationBetweenSignals tests stopping between checks.
func TestEarlyTerminationBetweenSignals(t *testing.T) {
	t.Parallel()

	// Simulate the collectActivity pattern
	handler := &maintainerHandler{
		elevated: map[string]struct{}{
			"user1": {},
			"user2": {},
			"user3": {},
		},
		active: make(map[string]bool),
	}

	signalsChecked := 0
	maxSignals := 5

	// Simulate checking multiple signals
	for i := 0; i < maxSignals; i++ {
		signalsChecked++

		// Simulate finding activity
		switch i {
		case 0:
			handler.markActive("user1")
		case 1:
			handler.markActive("user2")
		case 2:
			handler.markActive("user3")
		}

		// Check for early termination
		if handler.allActive() {
			break
		}
	}

	// Should have stopped after 3 signals (when all users became active)
	if signalsChecked != 3 {
		t.Errorf("should have checked 3 signals, but checked %d", signalsChecked)
	}

	// Verify all users are active
	if !handler.allActive() {
		t.Error("all users should be marked as active")
	}
}

// TestEarlyTerminationWithinNestedCalls tests stopping within nested calls.
func TestEarlyTerminationWithinNestedCalls(t *testing.T) {
	t.Parallel()

	handler := &maintainerHandler{
		elevated: map[string]struct{}{
			"user1": {},
			"user2": {},
		},
		active: make(map[string]bool),
	}

	// Simulate processing a list of items (e.g., PRs, issues) that would require nested API calls
	items := []string{"item1", "item2", "item3", "item4", "item5"}
	itemsProcessed := 0

	for _, item := range items {
		// Check for early termination before making expensive nested API call
		if handler.allActive() {
			break
		}

		itemsProcessed++

		// Simulate finding activity in nested calls
		switch item {
		case "item1":
			handler.markActive("user1")
		case "item2":
			handler.markActive("user2")
			// Both users now active, should stop on next iteration
		}
	}

	// Should have processed only 2 items before early termination
	if itemsProcessed != 2 {
		t.Errorf("should have processed 2 items, but processed %d", itemsProcessed)
	}

	// Verify all users are active
	if !handler.allActive() {
		t.Error("all users should be marked as active")
	}
}

// TestMarkActiveReturnValueOptimization tests return value optimization.
func TestMarkActiveReturnValueOptimization(t *testing.T) {
	t.Parallel()

	handler := &maintainerHandler{
		elevated: map[string]struct{}{
			"user1": {},
			"user2": {},
		},
		active: make(map[string]bool),
	}

	// Simulate processing events where the same user appears multiple times
	events := []string{"user1", "user2", "user1", "user1", "user2"}
	newActivations := 0

	for _, user := range events {
		// markActive returns true only for NEW activations
		if handler.markActive(user) {
			newActivations++
		}
	}

	// Should have had only 2 new activations despite 5 events
	if newActivations != 2 {
		t.Errorf("should have had 2 new activations, but had %d", newActivations)
	}

	// Verify both users are active
	if !handler.active["user1"] || !handler.active["user2"] {
		t.Error("both users should be marked as active")
	}
}

// TestEarlyTerminationWithMarkActiveReturn tests return value checking pattern.
func TestEarlyTerminationWithMarkActiveReturn(t *testing.T) {
	t.Parallel()

	handler := &maintainerHandler{
		elevated: map[string]struct{}{
			"user1": {},
			"user2": {},
		},
		active: make(map[string]bool),
	}

	// Simulate the pattern: if markActive returns true AND all are now active, terminate
	events := []struct {
		user            string
		shouldTerminate bool
	}{
		{"user1", false}, // First user active, but not all
		{"user2", true},  // Second user active, now all active
		{"user1", false}, // Already active, no termination needed
	}

	for i, event := range events {
		wasNew := handler.markActive(event.user)
		allActive := handler.allActive()

		if wasNew && allActive {
			if !event.shouldTerminate {
				t.Errorf("event %d: unexpected early termination", i)
			}
			// In real code, we would return/break here
			break
		} else if event.shouldTerminate {
			t.Errorf("event %d: expected early termination but didn't get it", i)
		}
	}

	// Verify both users are active
	if !handler.allActive() {
		t.Error("all users should be marked as active")
	}
}
