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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ossf/scorecard/v5/checker"
)

func TestGetPyPIArtifacts_Success(t *testing.T) {
	t.Parallel()

	artifactContent := []byte("wheel file content")
	attestationContent := []byte(`{
		"version": 1,
		"attestation_bundles": [{
			"publisher": {"kind": "important"},
			"attestations": [{"envelope": "test"}]
		}]
	}`)

	// Mock PyPI response
	pypiResponse := PyPISimpleResponse{
		Files: []PyPIFile{
			{
				Filename:   "test_package-1.0.0-py3-none-any.whl",
				URL:        "/test_package-1.0.0-py3-none-any.whl",
				Provenance: "/test_package-1.0.0-py3-none-any.whl.attestation",
				Hashes: map[string]string{
					"sha256": "abc123",
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/simple/"):
			// Check Accept header
			if r.Header.Get("Accept") != "application/vnd.pypi.simple.v1+json" {
				t.Errorf("Wrong Accept header: %s", r.Header.Get("Accept"))
			}
			w.Header().Set("Content-Type", "application/vnd.pypi.simple.v1+json")
			json.NewEncoder(w).Encode(pypiResponse) //nolint:errcheck // test mock server
		case strings.HasSuffix(r.URL.Path, ".whl"):
			w.WriteHeader(http.StatusOK)
			w.Write(artifactContent) //nolint:errcheck // test mock server
		case strings.HasSuffix(r.URL.Path, ".attestation"):
			w.WriteHeader(http.StatusOK)
			w.Write(attestationContent) //nolint:errcheck // test mock server
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	ctx := context.Background()
	client := &mockHTTPClient{server: server}

	pairs, err := GetPyPIArtifacts(ctx, "test-package", "1.0.0", client)
	if err != nil {
		t.Fatalf("GetPyPIArtifacts() error = %v", err)
	}

	if len(pairs) != 1 {
		t.Fatalf("Expected 1 pair, got %d", len(pairs))
	}

	pair := pairs[0]
	if pair.SignatureType != checker.SignatureTypeSigstore {
		t.Errorf("SignatureType = %v, want %v", pair.SignatureType, checker.SignatureTypeSigstore)
	}

	if string(pair.ArtifactData) != string(artifactContent) {
		t.Error("Artifact data mismatch")
	}

	if string(pair.SignatureData) != string(attestationContent) {
		t.Error("Attestation data mismatch")
	}
}

func TestGetPyPIArtifacts_NoAttestations(t *testing.T) {
	t.Parallel()

	// File without provenance field
	pypiResponse := PyPISimpleResponse{
		Files: []PyPIFile{
			{
				Filename:   "test_package-1.0.0-py3-none-any.whl",
				URL:        "/test_package-1.0.0-py3-none-any.whl",
				Provenance: "", // No attestation
				Hashes: map[string]string{
					"sha256": "abc123",
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.pypi.simple.v1+json")
		json.NewEncoder(w).Encode(pypiResponse) //nolint:errcheck // test mock server
	}))
	defer server.Close()

	ctx := context.Background()
	client := &mockHTTPClient{server: server}

	pairs, err := GetPyPIArtifacts(ctx, "test-package", "1.0.0", client)
	if err != nil {
		t.Fatalf("GetPyPIArtifacts() error = %v", err)
	}

	// Should return empty list when no attestations
	if len(pairs) != 0 {
		t.Errorf("Expected 0 pairs without attestations, got %d", len(pairs))
	}
}

func TestGetPyPIArtifacts_VersionNotFound(t *testing.T) {
	t.Parallel()

	// Return files for different version
	pypiResponse := PyPISimpleResponse{
		Files: []PyPIFile{
			{
				Filename:   "test_package-2.0.0-py3-none-any.whl",
				URL:        "/test_package-2.0.0-py3-none-any.whl",
				Provenance: "/test_package-2.0.0-py3-none-any.whl.attestation",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.pypi.simple.v1+json")
		json.NewEncoder(w).Encode(pypiResponse) //nolint:errcheck // test mock server
	}))
	defer server.Close()

	ctx := context.Background()
	client := &mockHTTPClient{server: server}

	_, err := GetPyPIArtifacts(ctx, "test-package", "1.0.0", client)
	if err == nil {
		t.Error("Expected error for version not found")
	}

	if !strings.Contains(err.Error(), "version not found") {
		t.Errorf("Expected 'version not found' error, got: %v", err)
	}
}

func TestGetPyPIArtifacts_MultipleFiles(t *testing.T) {
	t.Parallel()

	pypiResponse := PyPISimpleResponse{
		Files: []PyPIFile{
			{
				Filename:   "test_package-1.0.0-py3-none-any.whl",
				URL:        "/test_package-1.0.0-py3-none-any.whl",
				Provenance: "/test_package-1.0.0-py3-none-any.whl.attestation",
			},
			{
				Filename:   "test_package-1.0.0.tar.gz",
				URL:        "/test_package-1.0.0.tar.gz",
				Provenance: "/test_package-1.0.0.tar.gz.attestation",
			},
			{
				Filename:   "test_package-2.0.0-py3-none-any.whl", // Different version
				URL:        "/test_package-2.0.0-py3-none-any.whl",
				Provenance: "/test_package-2.0.0-py3-none-any.whl.attestation",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/simple/"):
			w.Header().Set("Content-Type", "application/vnd.pypi.simple.v1+json")
			json.NewEncoder(w).Encode(pypiResponse) //nolint:errcheck // test mock server
		default:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("content")) //nolint:errcheck // test mock server
		}
	}))
	defer server.Close()

	ctx := context.Background()
	client := &mockHTTPClient{server: server}

	pairs, err := GetPyPIArtifacts(ctx, "test-package", "1.0.0", client)
	if err != nil {
		t.Fatalf("GetPyPIArtifacts() error = %v", err)
	}

	// Should get 2 pairs (wheel and tar.gz for 1.0.0, not 2.0.0)
	if len(pairs) != 2 {
		t.Errorf("Expected 2 pairs for version 1.0.0, got %d", len(pairs))
	}

	// Verify both are for version 1.0.0
	for _, pair := range pairs {
		if !strings.Contains(pair.ArtifactURL, "1.0.0") {
			t.Errorf("Unexpected artifact URL: %s", pair.ArtifactURL)
		}
	}
}

func TestGetPyPIArtifacts_APIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	ctx := context.Background()
	client := &mockHTTPClient{server: server}

	_, err := GetPyPIArtifacts(ctx, "nonexistent-package", "1.0.0", client)
	if err == nil {
		t.Error("Expected error for 404 API response")
	}
}

func TestGetPyPIArtifacts_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not valid json")) //nolint:errcheck // test mock server
	}))
	defer server.Close()

	ctx := context.Background()
	client := &mockHTTPClient{server: server}

	_, err := GetPyPIArtifacts(ctx, "test-package", "1.0.0", client)
	if err == nil {
		t.Error("Expected error for invalid JSON response")
	}
}

func TestGetPyPIArtifacts_ArtifactDownloadFails(t *testing.T) {
	t.Parallel()

	pypiResponse := PyPISimpleResponse{
		Files: []PyPIFile{
			{
				Filename:   "test_package-1.0.0-py3-none-any.whl",
				URL:        "/test_package-1.0.0-py3-none-any.whl",
				Provenance: "/test_package-1.0.0-py3-none-any.whl.attestation",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/simple/"):
			w.Header().Set("Content-Type", "application/vnd.pypi.simple.v1+json")
			json.NewEncoder(w).Encode(pypiResponse) //nolint:errcheck // test mock server
		case strings.HasSuffix(r.URL.Path, ".whl"):
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("attestation")) //nolint:errcheck // test mock server
		}
	}))
	defer server.Close()

	ctx := context.Background()
	client := &mockHTTPClient{server: server}

	pairs, err := GetPyPIArtifacts(ctx, "test-package", "1.0.0", client)
	if err != nil {
		t.Fatalf("GetPyPIArtifacts() error = %v", err)
	}

	// Should skip files where download fails
	if len(pairs) != 0 {
		t.Errorf("Expected 0 pairs when artifact download fails, got %d", len(pairs))
	}
}

func TestGetPyPIArtifacts_AttestationDownloadFails(t *testing.T) {
	t.Parallel()

	pypiResponse := PyPISimpleResponse{
		Files: []PyPIFile{
			{
				Filename:   "test_package-1.0.0-py3-none-any.whl",
				URL:        "/test_package-1.0.0-py3-none-any.whl",
				Provenance: "/test_package-1.0.0-py3-none-any.whl.attestation",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/simple/"):
			w.Header().Set("Content-Type", "application/vnd.pypi.simple.v1+json")
			json.NewEncoder(w).Encode(pypiResponse) //nolint:errcheck // test mock server
		case strings.HasSuffix(r.URL.Path, ".whl"):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("artifact")) //nolint:errcheck // test mock server
		case strings.HasSuffix(r.URL.Path, ".attestation"):
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	ctx := context.Background()
	client := &mockHTTPClient{server: server}

	pairs, err := GetPyPIArtifacts(ctx, "test-package", "1.0.0", client)
	if err != nil {
		t.Fatalf("GetPyPIArtifacts() error = %v", err)
	}

	// Should skip files where attestation download fails
	if len(pairs) != 0 {
		t.Errorf("Expected 0 pairs when attestation download fails, got %d", len(pairs))
	}
}

func TestGetPyPIArtifacts_ContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := GetPyPIArtifacts(ctx, "test-package", "1.0.0", nil)
	if err == nil {
		t.Error("Expected error for cancelled context")
	}
}

func TestGetPyPIArtifacts_NilClient(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}
	t.Parallel()

	ctx := context.Background()
	// Test with nil client - should use default HTTP client
	_, err := GetPyPIArtifacts(ctx, "this-package-definitely-does-not-exist-12345", "1.0.0", nil)

	if err == nil {
		t.Error("Expected error for nonexistent package")
	}
}

func TestDownloadFileContent_Success(t *testing.T) {
	t.Parallel()

	expected := []byte("file content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(expected) //nolint:errcheck // test mock server
	}))
	defer server.Close()

	ctx := context.Background()
	client := &mockHTTPClient{server: server}

	data, err := downloadFileContent(ctx, client, server.URL+"/test.whl")
	if err != nil {
		t.Fatalf("downloadFileContent() error = %v", err)
	}

	if string(data) != string(expected) {
		t.Errorf("Downloaded data = %v, want %v", data, expected)
	}
}

func TestDownloadFileContent_StatusNotOK(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
	}{
		{"not found", http.StatusNotFound},
		{"server error", http.StatusInternalServerError},
		{"forbidden", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			ctx := context.Background()
			client := &mockHTTPClient{server: server}

			_, err := downloadFileContent(ctx, client, server.URL+"/test")
			if err == nil {
				t.Errorf("Expected error for status %d", tt.statusCode)
			}
		})
	}
}

func TestGetPyPIArtifacts_EmptyFiles(t *testing.T) {
	t.Parallel()

	pypiResponse := PyPISimpleResponse{
		Files: []PyPIFile{},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.pypi.simple.v1+json")
		json.NewEncoder(w).Encode(pypiResponse) //nolint:errcheck // test mock server
	}))
	defer server.Close()

	ctx := context.Background()
	client := &mockHTTPClient{server: server}

	_, err := GetPyPIArtifacts(ctx, "test-package", "1.0.0", client)
	if err == nil {
		t.Error("Expected error for empty files list")
	}

	if !strings.Contains(err.Error(), "version not found") {
		t.Errorf("Expected 'version not found' error, got: %v", err)
	}
}
