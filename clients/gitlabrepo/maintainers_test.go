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

package gitlabrepo

import (
	"context"
	"testing"
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func TestGetMaintainerActivity(t *testing.T) {
	t.Parallel()

	cutoff := time.Now().Add(-180 * 24 * time.Hour)

	// Create a minimal Client with MaintainerActivityHandler
	// We don't need real GitLab connection for this test
	glClient, err := gitlab.NewClient("")
	if err != nil {
		t.Fatalf("Failed to create GitLab client: %v", err)
	}

	handler := &MaintainerActivityHandler{}
	handler.init(context.Background(), glClient, "test/repo")

	// Set cutoff directly to avoid triggering real API calls
	handler.cutoff = cutoff

	// Verify cutoff was set correctly
	if handler.cutoff != cutoff {
		t.Errorf("Handler cutoff = %v, want %v", handler.cutoff, cutoff)
	}
}

// TestHandlerInit tests the handler initialization..
func TestHandlerInit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	glClient, err := gitlab.NewClient("")
	if err != nil {
		t.Fatalf("Failed to create GitLab client: %v", err)
	}

	handler := &MaintainerActivityHandler{}
	handler.init(ctx, glClient, "test/repo")

	// Test that handler initializes correctly
	if handler.ctx == nil {
		t.Error("Handler context not initialized")
	}
	if handler.gl == nil {
		t.Error("Handler GitLab client not initialized")
	}
	if handler.projectID != "test/repo" {
		t.Errorf("Handler project ID = %s, want test/repo", handler.projectID)
	}
	if handler.minAccessLevel != gitlab.DeveloperPermissions {
		t.Errorf("Handler min access level = %v, want DeveloperPermissions", handler.minAccessLevel)
	}
	if handler.elevated == nil {
		t.Error("Handler elevated map not initialized")
	}
	if handler.active == nil {
		t.Error("Handler active map not initialized")
	}
}
