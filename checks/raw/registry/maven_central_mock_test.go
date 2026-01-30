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
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"

	"github.com/ossf/scorecard/v5/checker"
)

// mockHTTPClient wraps httptest.Server to implement HTTPClient interface.
type mockHTTPClient struct {
	server *httptest.Server
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	// Rewrite URL to point to mock server
	req.URL.Scheme = "http"
	req.URL.Host = strings.TrimPrefix(m.server.URL, "http://")
	return m.server.Client().Do(req)
}

func TestGetMavenArtifacts_Success(t *testing.T) {
	t.Parallel()

	// Create test GPG key
	entity, err := openpgp.NewEntity(
		"Test Maven",
		"Maven Test",
		"test@maven.org",
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create test entity: %v", err)
	}

	artifactData := []byte("test jar content")

	// Sign the artifact
	var sigBuf bytes.Buffer
	if err := openpgp.ArmoredDetachSign(&sigBuf, entity, bytes.NewReader(artifactData), nil); err != nil {
		t.Fatalf("Failed to create signature: %v", err)
	}

	// Export public key
	var keyBuf bytes.Buffer
	armorWriter, err := armor.Encode(&keyBuf, openpgp.PublicKeyType, nil)
	if err != nil {
		t.Fatalf("Failed to create armor encoder: %v", err)
	}
	if err := entity.Serialize(armorWriter); err != nil {
		t.Fatalf("Failed to serialize entity: %v", err)
	}
	armorWriter.Close()

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case strings.HasSuffix(path, ".jar"):
			w.WriteHeader(http.StatusOK)
			w.Write(artifactData) //nolint:errcheck // test mock server
		case strings.HasSuffix(path, ".jar.asc"):
			w.WriteHeader(http.StatusOK)
			w.Write(sigBuf.Bytes()) //nolint:errcheck // test mock server
		case strings.HasSuffix(path, ".pom"):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("<project></project>")) //nolint:errcheck // test mock server
		case strings.HasSuffix(path, ".pom.asc"):
			w.WriteHeader(http.StatusOK)
			w.Write(sigBuf.Bytes()) //nolint:errcheck // test mock server
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	ctx := context.Background()
	client := &mockHTTPClient{server: server}

	pairs, err := GetMavenArtifacts(ctx, "com.example:test-artifact", "1.0.0", client)
	if err != nil {
		t.Fatalf("GetMavenArtifacts() error = %v", err)
	}

	if len(pairs) != 2 {
		t.Errorf("Expected 2 artifact-signature pairs, got %d", len(pairs))
	}

	// Check that we got signatures for both .jar and .pom
	foundJar := false
	foundPom := false
	for _, pair := range pairs {
		if strings.HasSuffix(pair.ArtifactURL, ".jar") {
			foundJar = true
			if pair.SignatureType != checker.SignatureTypeGPG {
				t.Errorf("Expected GPG signature type for jar, got %v", pair.SignatureType)
			}
		}
		if strings.HasSuffix(pair.ArtifactURL, ".pom") {
			foundPom = true
		}
	}

	if !foundJar {
		t.Error("Expected .jar artifact in results")
	}
	if !foundPom {
		t.Error("Expected .pom artifact in results")
	}
}

func TestGetMavenArtifacts_NoSignatures(t *testing.T) {
	t.Parallel()

	// Mock server that returns 404 for signatures
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, ".jar") || strings.HasSuffix(path, ".pom") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("content")) //nolint:errcheck // test mock server
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	ctx := context.Background()
	client := &mockHTTPClient{server: server}

	pairs, err := GetMavenArtifacts(ctx, "com.example:no-sigs", "1.0.0", client)
	if err != nil {
		t.Fatalf("GetMavenArtifacts() error = %v", err)
	}

	if len(pairs) != 0 {
		t.Errorf("Expected 0 pairs when signatures not found, got %d", len(pairs))
	}
}

func TestGetMavenArtifacts_InvalidCoordinates(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name        string
		packageName string
	}{
		{
			name:        "missing colon",
			packageName: "com.example.artifact",
		},
		{
			name:        "too many parts",
			packageName: "com.example:artifact:extra",
		},
		{
			name:        "empty string",
			packageName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := GetMavenArtifacts(ctx, tt.packageName, "1.0.0", nil)
			if err == nil {
				t.Error("Expected error for invalid coordinates")
			}
		})
	}
}

func TestGetMavenArtifacts_ServerError(t *testing.T) {
	t.Parallel()

	// Mock server that always returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	ctx := context.Background()
	client := &mockHTTPClient{server: server}

	pairs, err := GetMavenArtifacts(ctx, "com.example:error", "1.0.0", client)
	if err != nil {
		t.Fatalf("GetMavenArtifacts() error = %v", err)
	}

	// Should return empty list when artifacts can't be downloaded
	if len(pairs) != 0 {
		t.Errorf("Expected 0 pairs on server error, got %d", len(pairs))
	}
}

func TestGetMavenArtifacts_ContextCancellation(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		<-r.Context().Done()
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client := &mockHTTPClient{server: server}
	pairs, err := GetMavenArtifacts(ctx, "com.example:test", "1.0.0", client)
	// Should handle context cancellation gracefully
	if err != nil {
		t.Fatalf("GetMavenArtifacts() error = %v", err)
	}

	if len(pairs) != 0 {
		t.Errorf("Expected 0 pairs with cancelled context, got %d", len(pairs))
	}
}

func TestDownloadFile_Success(t *testing.T) {
	t.Parallel()

	expected := []byte("test file content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(expected) //nolint:errcheck // test mock server
	}))
	defer server.Close()

	ctx := context.Background()
	client := &mockHTTPClient{server: server}

	data, err := downloadFile(ctx, client, server.URL+"/test.jar")
	if err != nil {
		t.Fatalf("downloadFile() error = %v", err)
	}

	if !bytes.Equal(data, expected) {
		t.Errorf("Downloaded data = %v, want %v", data, expected)
	}
}

func TestDownloadFile_NotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	ctx := context.Background()
	client := &mockHTTPClient{server: server}

	_, err := downloadFile(ctx, client, server.URL+"/missing.jar")
	if err == nil {
		t.Error("Expected error for 404 response")
	}
}

func TestCheckURLExists(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{
			name:       "exists (200)",
			statusCode: http.StatusOK,
			want:       true,
		},
		{
			name:       "not found (404)",
			statusCode: http.StatusNotFound,
			want:       false,
		},
		{
			name:       "server error (500)",
			statusCode: http.StatusInternalServerError,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodHead {
					t.Errorf("Expected HEAD request, got %s", r.Method)
				}
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			ctx := context.Background()
			client := &mockHTTPClient{server: server}

			got := checkURLExists(ctx, client, server.URL+"/test")
			if got != tt.want {
				t.Errorf("checkURLExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDownloadArtifactPair_Success(t *testing.T) {
	t.Parallel()

	artifactContent := []byte("artifact")
	signatureContent := []byte("signature")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".jar") {
			w.Write(artifactContent) //nolint:errcheck // test mock server
		} else if strings.HasSuffix(r.URL.Path, ".asc") {
			w.Write(signatureContent) //nolint:errcheck // test mock server
		}
	}))
	defer server.Close()

	ctx := context.Background()
	client := &mockHTTPClient{server: server}

	pair, err := downloadArtifactPair(
		ctx,
		client,
		server.URL+"/test.jar",
		server.URL+"/test.jar.asc",
		checker.SignatureTypeGPG,
	)
	if err != nil {
		t.Fatalf("downloadArtifactPair() error = %v", err)
	}

	if !bytes.Equal(pair.ArtifactData, artifactContent) {
		t.Error("Artifact data mismatch")
	}

	if !bytes.Equal(pair.SignatureData, signatureContent) {
		t.Error("Signature data mismatch")
	}

	if pair.SignatureType != checker.SignatureTypeGPG {
		t.Errorf("SignatureType = %v, want %v", pair.SignatureType, checker.SignatureTypeGPG)
	}
}

func TestDownloadArtifactPair_ArtifactFails(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".jar") {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.Write([]byte("signature")) //nolint:errcheck // test mock server
		}
	}))
	defer server.Close()

	ctx := context.Background()
	client := &mockHTTPClient{server: server}

	_, err := downloadArtifactPair(
		ctx,
		client,
		server.URL+"/test.jar",
		server.URL+"/test.jar.asc",
		checker.SignatureTypeGPG,
	)
	if err == nil {
		t.Error("Expected error when artifact download fails")
	}
}

func TestDownloadArtifactPair_SignatureFails(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".jar") {
			w.Write([]byte("artifact")) //nolint:errcheck // test mock server
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	ctx := context.Background()
	client := &mockHTTPClient{server: server}

	_, err := downloadArtifactPair(
		ctx,
		client,
		server.URL+"/test.jar",
		server.URL+"/test.jar.asc",
		checker.SignatureTypeGPG,
	)
	if err == nil {
		t.Error("Expected error when signature download fails")
	}
}

func TestGetMavenArtifacts_NilClient(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}
	t.Parallel()

	ctx := context.Background()
	// Test with nil client - should use default HTTP client
	// Use a nonexistent package to avoid hitting real Maven Central
	pairs, err := GetMavenArtifacts(ctx, "com.nonexistent:fake-package", "999.999.999", nil)
	if err != nil {
		t.Fatalf("GetMavenArtifacts() error = %v", err)
	}

	// Should return empty list for nonexistent package
	if len(pairs) != 0 {
		t.Errorf("Expected 0 pairs for nonexistent package, got %d", len(pairs))
	}
}

func TestGetMavenArtifacts_ReadBodyError(t *testing.T) {
	t.Parallel()

	// Server that returns error body
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100")
		w.WriteHeader(http.StatusOK)
		// Don't write anything - causes io.ReadAll to get unexpected EOF
	}))
	defer server.Close()

	ctx := context.Background()
	client := &mockHTTPClient{server: server}

	pairs, err := GetMavenArtifacts(ctx, "com.example:test", "1.0.0", client)
	if err != nil {
		t.Fatalf("GetMavenArtifacts() error = %v", err)
	}

	// Should handle read errors gracefully
	if len(pairs) != 0 {
		t.Errorf("Expected 0 pairs on read error, got %d", len(pairs))
	}
}
