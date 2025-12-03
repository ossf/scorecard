// Copyright 2025 OpenSSF Scorecard Authors
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

package clients

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNewDirectDepsClient(t *testing.T) {
	t.Parallel()
	client := NewDirectDepsClient()
	if client == nil {
		t.Fatal("NewDirectDepsClient() returned nil")
	}
}

func TestScalibrDepsClient_GetDeps_EmptyDir(t *testing.T) {
	t.Parallel()
	client := NewDirectDepsClient()

	// Test with empty string
	_, err := client.GetDeps(context.Background(), "")
	if err == nil {
		t.Error("GetDeps with empty dir should return error")
	}
}

func TestScalibrDepsClient_GetDeps_NonexistentDir(t *testing.T) {
	t.Parallel()
	client := NewDirectDepsClient()

	// Test with nonexistent directory - scalibr succeeds but finds no dependencies
	resp, err := client.GetDeps(context.Background(), "/nonexistent/path/to/nowhere")
	if err != nil {
		t.Errorf("GetDeps with nonexistent dir should not error, got: %v", err)
	}
	if len(resp.Deps) != 0 {
		t.Errorf("Expected 0 dependencies from nonexistent dir, got %d", len(resp.Deps))
	}
}

func TestScalibrDepsClient_GetDeps_ValidGoProject(t *testing.T) {
	t.Parallel()
	// Create a temporary directory with a simple Go project
	tmpDir := t.TempDir()

	// Create a go.mod file
	goModContent := `module example.com/test

go 1.21

require (
	github.com/google/uuid v1.3.0
)
`
	err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create a main.go file
	mainGoContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`
	err = os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainGoContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create main.go: %v", err)
	}

	client := NewDirectDepsClient()
	resp, err := client.GetDeps(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("GetDeps failed: %v", err)
	}

	// Check that we got the uuid dependency
	found := false
	for _, dep := range resp.Deps {
		if dep.Name == "github.com/google/uuid" {
			found = true
			// Note: scalibr's Go module extractor returns versions without "v" prefix
			if dep.Version != "1.3.0" {
				t.Errorf("Expected version 1.3.0, got %s", dep.Version)
			}
			if dep.Ecosystem != "golang" {
				t.Errorf("Expected ecosystem golang, got %s", dep.Ecosystem)
			}
		}
	}

	if !found {
		t.Error("Expected to find github.com/google/uuid dependency")
	}
}

func TestScalibrDepsClient_GetDeps_ValidNPMProject(t *testing.T) {
	t.Parallel()
	// Create a temporary directory with a simple NPM project
	tmpDir := t.TempDir()

	// Create a package.json file
	packageJSONContent := `{
  "name": "test-project",
  "version": "1.0.0",
  "dependencies": {
    "express": "4.18.2"
  }
}
`
	err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJSONContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	// Create a package-lock.json as scalibr needs it to find dependencies
	packageLockJSONContent := `{
  "name": "test-project",
  "version": "1.0.0",
  "lockfileVersion": 2,
  "requires": true,
  "packages": {
    "": {
      "name": "test-project",
      "version": "1.0.0",
      "dependencies": {
        "express": "4.18.2"
      }
    },
    "node_modules/express": {
      "version": "4.18.2",
      "resolved": "https://registry.npmjs.org/express/-/express-4.18.2.tgz"
    }
  },
  "dependencies": {
    "express": {
      "version": "4.18.2",
      "resolved": "https://registry.npmjs.org/express/-/express-4.18.2.tgz"
    }
  }
}
`
	err = os.WriteFile(filepath.Join(tmpDir, "package-lock.json"), []byte(packageLockJSONContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create package-lock.json: %v", err)
	}

	client := NewDirectDepsClient()
	resp, err := client.GetDeps(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("GetDeps failed: %v", err)
	}

	// Check that we got the express dependency
	found := false
	for _, dep := range resp.Deps {
		if dep.Name == "express" {
			found = true
			if dep.Version != "4.18.2" {
				t.Errorf("Expected version 4.18.2, got %s", dep.Version)
			}
			if dep.Ecosystem != "npm" {
				t.Errorf("Expected ecosystem npm, got %s", dep.Ecosystem)
			}
		}
	}

	if !found {
		t.Error("Expected to find express dependency")
	}
}

func TestNormalizePURLType(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected string
	}{
		{"golang", "golang"},
		{"go", "golang"},
		{"gomod", "golang"},
		{"npm", "npm"},
		{"node", "npm"},
		{"pypi", "pypi"},
		{"python", "pypi"},
		{"maven", "maven"},
		{"cargo", "cargo"},
		{"rust", "cargo"},
		{"nuget", "nuget"},
		{"gem", "gem"},
		{"ruby", "gem"},
		{"unknown", "unknown"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			result := normalizePURLType(tt.input)
			if result != tt.expected {
				t.Errorf("normalizePURLType(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestScalibrDepsClient_GetDeps_DeterministicOrder(t *testing.T) {
	t.Parallel()
	// Create a temporary directory with multiple dependencies
	tmpDir := t.TempDir()

	goModContent := `module example.com/test

go 1.21

require (
	github.com/google/uuid v1.3.0
	github.com/stretchr/testify v1.8.0
	golang.org/x/sync v0.1.0
)
`
	err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	client := NewDirectDepsClient()

	// Run GetDeps multiple times and ensure order is consistent
	resp1, err := client.GetDeps(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("GetDeps failed: %v", err)
	}

	resp2, err := client.GetDeps(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("GetDeps failed: %v", err)
	}

	if len(resp1.Deps) != len(resp2.Deps) {
		t.Fatalf("Inconsistent number of deps: %d vs %d", len(resp1.Deps), len(resp2.Deps))
	}

	for i := range resp1.Deps {
		if resp1.Deps[i].Name != resp2.Deps[i].Name {
			t.Errorf("Deps order mismatch at index %d: %s vs %s",
				i, resp1.Deps[i].Name, resp2.Deps[i].Name)
		}
	}
}
