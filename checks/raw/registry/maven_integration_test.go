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

//go:build integration
// +build integration

package registry

import (
	"context"
	"testing"
	"time"
)

func TestGetMavenArtifacts_RealProject(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test with a small, well-known Maven package
	pairs, err := GetMavenArtifacts(ctx, "junit:junit", "4.13.2", nil)
	if err != nil {
		t.Fatalf("GetMavenArtifacts() error = %v", err)
	}

	if len(pairs) == 0 {
		t.Error("Expected at least one artifact pair, got 0")
	}

	for i, pair := range pairs {
		t.Logf("Pair %d:", i)
		t.Logf("  Artifact URL: %s", pair.ArtifactURL)
		t.Logf("  Signature URL: %s", pair.SignatureURL)
		t.Logf("  Signature Type: %s", pair.SignatureType)
		t.Logf("  Artifact Size: %d bytes", len(pair.ArtifactData))
		t.Logf("  Signature Size: %d bytes", len(pair.SignatureData))
	}
}
