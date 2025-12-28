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

package registry

import (
	"context"
	"testing"
)

// Note: These are integration tests that hit the real PyPI API.
// For comprehensive unit tests with mock HTTP responses, see pypi_mock_test.go.

func TestGetPyPIArtifacts_ContextCancellationOld(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := GetPyPIArtifacts(ctx, "test-package", "1.0.0", nil)

	if err == nil {
		t.Error("Expected error for cancelled context")
	}
}

func TestGetPyPIArtifacts_NonexistentPackage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}
	t.Parallel()

	ctx := context.Background()
	_, err := GetPyPIArtifacts(ctx, "this-package-definitely-does-not-exist-12345", "1.0.0", nil)

	if err == nil {
		t.Error("Expected error for nonexistent package")
	}
}

func TestGetPyPIArtifacts_InvalidVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}
	t.Parallel()

	ctx := context.Background()
	// Use a real package with an invalid version
	_, err := GetPyPIArtifacts(ctx, "requests", "999.999.999", nil)

	if err == nil {
		t.Error("Expected error for invalid version")
	}
}
