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

package signature

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"

	"github.com/ossf/scorecard/v5/checker"
)

// TestGetVerifier verifies the correct verifier is returned.
func TestGetVerifier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		signatureType checker.SignatureType
		wantType      string
	}{
		{
			name:          "GPG verifier",
			signatureType: checker.SignatureTypeGPG,
			wantType:      "*signature.GPGVerifier",
		},
		{
			name:          "Sigstore verifier",
			signatureType: checker.SignatureTypeSigstore,
			wantType:      "*signature.SigstoreVerifier",
		},
		{
			name:          "Minisign verifier",
			signatureType: checker.SignatureTypeMinisign,
			wantType:      "*signature.MinisignVerifier",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			verifier := GetVerifier(tt.signatureType)
			if verifier == nil {
				t.Errorf("GetVerifier(%v) returned nil", tt.signatureType)
				return
			}

			if !verifier.CanVerify(tt.signatureType) {
				t.Errorf("Verifier cannot verify its own signature type %v", tt.signatureType)
			}
		})
	}
}

func TestGPGVerifier_CanVerify(t *testing.T) {
	t.Parallel()

	v := &GPGVerifier{}

	if !v.CanVerify(checker.SignatureTypeGPG) {
		t.Error("GPGVerifier should be able to verify GPG signatures")
	}

	if v.CanVerify(checker.SignatureTypeSigstore) {
		t.Error("GPGVerifier should not be able to verify Sigstore signatures")
	}

	if v.CanVerify(checker.SignatureTypeMinisign) {
		t.Error("GPGVerifier should not be able to verify Minisign signatures")
	}
}

func TestSigstoreVerifier_CanVerify(t *testing.T) {
	t.Parallel()

	v := &SigstoreVerifier{}

	if !v.CanVerify(checker.SignatureTypeSigstore) {
		t.Error("SigstoreVerifier should be able to verify Sigstore signatures")
	}

	if v.CanVerify(checker.SignatureTypeGPG) {
		t.Error("SigstoreVerifier should not be able to verify GPG signatures")
	}

	if v.CanVerify(checker.SignatureTypeMinisign) {
		t.Error("SigstoreVerifier should not be able to verify Minisign signatures")
	}
}

func TestMinisignVerifier_CanVerify(t *testing.T) {
	t.Parallel()

	v := &MinisignVerifier{}

	if !v.CanVerify(checker.SignatureTypeMinisign) {
		t.Error("MinisignVerifier should be able to verify Minisign signatures")
	}

	if v.CanVerify(checker.SignatureTypeGPG) {
		t.Error("MinisignVerifier should not be able to verify GPG signatures")
	}

	if v.CanVerify(checker.SignatureTypeSigstore) {
		t.Error("MinisignVerifier should not be able to verify Sigstore signatures")
	}
}

func TestGPGVerifier_Verify_KeyExtraction(t *testing.T) {
	t.Parallel()

	v := &GPGVerifier{}
	ctx := context.Background()

	artifact := []byte("test artifact")
	signature := []byte("fake signature")

	result, err := v.Verify(ctx, artifact, signature, &VerifyOptions{})
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	if result.Verified {
		t.Error("Verification should fail with invalid signature")
	}

	if result.Error == nil {
		t.Error("Expected error with invalid signature")
	}
}

func TestMinisignVerifier_Verify_NotImplemented(t *testing.T) {
	t.Parallel()

	v := &MinisignVerifier{}
	ctx := context.Background()

	artifact := []byte("test artifact")
	signature := []byte("fake signature")

	result, err := v.Verify(ctx, artifact, signature, &VerifyOptions{})
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	if result.Verified {
		t.Error("Minisign verification should fail (not implemented)")
	}

	if result.Error == nil {
		t.Error("Expected error for unimplemented Minisign verification")
	}
}

func TestGPGVerifier_extractKeyIDFromSignature(t *testing.T) {
	t.Parallel()

	// Create a real test signature to extract key ID from
	entity, err := openpgp.NewEntity("Test User", "Test", "test@example.com", nil)
	if err != nil {
		t.Fatalf("Failed to create test entity: %v", err)
	}

	message := []byte("test message")
	var sigBuf bytes.Buffer
	err = openpgp.ArmoredDetachSign(&sigBuf, entity, bytes.NewReader(message), nil)
	if err != nil {
		t.Fatalf("Failed to create signature: %v", err)
	}

	tests := []struct {
		name      string
		signature []byte
		wantError bool
	}{
		{
			name:      "valid signature",
			signature: sigBuf.Bytes(),
			wantError: false,
		},
		{
			name:      "invalid armored data",
			signature: []byte("-----BEGIN PGP SIGNATURE-----\ninvalid\n-----END PGP SIGNATURE-----"),
			wantError: true,
		},
		{
			name:      "not armored",
			signature: []byte("this is not an armored signature"),
			wantError: true,
		},
		{
			name:      "empty signature",
			signature: []byte(""),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v := &GPGVerifier{}
			keyID, err := v.extractKeyIDFromSignature(tt.signature)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if keyID == "" {
				t.Error("Expected non-empty key ID")
			}

			// Key ID should be a hex string
			if len(keyID) != 16 {
				t.Errorf(
					"Expected 16-character key ID, got %d characters: %s",
					len(keyID),
					keyID,
				)
			}
		})
	}
}

func TestGPGVerifier_fetchPublicKeyFromURL(t *testing.T) {
	t.Parallel()

	// Create a test GPG key
	entity, err := openpgp.NewEntity(
		"Test User",
		"Test",
		"test@example.com",
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create test entity: %v", err)
	}

	// Serialize the public key
	var keyBuf bytes.Buffer
	armorWriter, err := armor.Encode(&keyBuf, openpgp.PublicKeyType, nil)
	if err != nil {
		t.Fatalf("Failed to create armor encoder: %v", err)
	}
	if err := entity.Serialize(armorWriter); err != nil {
		t.Fatalf("Failed to serialize entity: %v", err)
	}
	armorWriter.Close()

	tests := []struct {
		body       string
		name       string
		statusCode int
		wantError  bool
	}{
		{
			name:       "successful key fetch",
			statusCode: http.StatusOK,
			body:       keyBuf.String(),
			wantError:  false,
		},
		{
			name:       "404 not found",
			statusCode: http.StatusNotFound,
			body:       "Not Found",
			wantError:  true,
		},
		{
			name:       "500 server error",
			statusCode: http.StatusInternalServerError,
			body:       "Server Error",
			wantError:  true,
		},
		{
			name:       "invalid key data",
			statusCode: http.StatusOK,
			body:       "not a valid pgp key",
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create test server
			server := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tt.statusCode)
					if _, err := w.Write([]byte(tt.body)); err != nil {
						t.Errorf("Failed to write response: %v", err)
					}
				},
			))
			defer server.Close()

			v := &GPGVerifier{client: server.Client()}
			ctx := context.Background()

			keyring, keyID, err := v.fetchPublicKeyFromURL(ctx, server.URL)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if keyring == nil {
				t.Error("Expected keyring but got nil")
			}

			if keyID == "" {
				t.Error("Expected key ID but got empty string")
			}
		})
	}
}

func TestGPGVerifier_Verify_WithKeyExtraction(t *testing.T) {
	t.Parallel()

	// Create a real test key and signature
	entity, err := openpgp.NewEntity(
		"Test User",
		"Test",
		"test@example.com",
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create test entity: %v", err)
	}

	// Sign a test message
	message := []byte("test message to sign")
	var sigBuf bytes.Buffer
	err = openpgp.ArmoredDetachSign(&sigBuf, entity, bytes.NewReader(message), nil)
	if err != nil {
		t.Fatalf("Failed to create signature: %v", err)
	}

	// Serialize the public key
	var keyBuf bytes.Buffer
	armorWriter, err := armor.Encode(&keyBuf, openpgp.PublicKeyType, nil)
	if err != nil {
		t.Fatalf("Failed to create armor encoder: %v", err)
	}
	if err := entity.Serialize(armorWriter); err != nil {
		t.Fatalf("Failed to serialize entity: %v", err)
	}
	armorWriter.Close()

	// Create test server that returns the public key
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(keyBuf.Bytes()); err != nil {
			t.Errorf("Failed to write key: %v", err)
		}
	}))
	defer server.Close()

	v := &GPGVerifier{client: server.Client()}
	ctx := context.Background()

	// Test verification with keyserver URL
	opts := VerifyOptions{
		KeyserverURL: server.URL,
	}

	result, err := v.Verify(ctx, message, sigBuf.Bytes(), &opts)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	if !result.Verified {
		t.Errorf("Verification failed: %v", result.Error)
	}

	if result.KeyID == "" {
		t.Error("Expected key ID but got empty string")
	}
}

func TestGPGVerifier_Verify_InvalidSignature(t *testing.T) {
	t.Parallel()

	v := &GPGVerifier{}
	ctx := context.Background()

	tests := []struct {
		name      string
		opts      VerifyOptions
		artifact  []byte
		signature []byte
	}{
		{
			name:      "empty signature",
			artifact:  []byte("test"),
			signature: []byte(""),
			opts:      VerifyOptions{},
		},
		{
			name:      "invalid signature format",
			artifact:  []byte("test"),
			signature: []byte("not a valid signature"),
			opts:      VerifyOptions{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := v.Verify(ctx, tt.artifact, tt.signature, &tt.opts)
			if err != nil {
				t.Fatalf("Verify() error = %v", err)
			}

			if result.Verified {
				t.Error("Expected verification to fail")
			}

			if result.Error == nil {
				t.Error("Expected error in result")
			}
		})
	}
}

func TestGPGVerifier_findKeyIDInPackets(t *testing.T) {
	t.Parallel()

	v := &GPGVerifier{}

	tests := []struct {
		name        string
		signature   string
		expectKeyID bool
		expectError bool
	}{
		{
			name:        "empty data",
			signature:   "",
			expectKeyID: false,
			expectError: true,
		},
		{
			name:        "invalid armored data",
			signature:   "not a valid signature",
			expectKeyID: false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			keyID, err := v.extractKeyIDFromSignature([]byte(tt.signature))

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.expectKeyID && keyID == "" {
				t.Error("Expected key ID but got empty string")
			}
		})
	}
}

func TestGPGVerifier_extractKeyIDFromPacket(t *testing.T) {
	t.Parallel()

	v := &GPGVerifier{}

	// Test with nil packet
	keyID := v.extractKeyIDFromPacket(nil)
	if keyID != "" {
		t.Errorf("Expected empty key ID for nil packet, got %s", keyID)
	}
}

func TestGPGVerifier_checkSignature_Coverage(t *testing.T) {
	t.Parallel()

	v := &GPGVerifier{}
	ctx := context.Background()

	tests := []struct {
		name      string
		artifact  []byte
		keyURL    string
		signature []byte
	}{
		{
			name:      "malformed signature",
			artifact:  []byte("test artifact"),
			signature: []byte("not a valid signature"),
		},
		{
			name:      "empty signature",
			artifact:  []byte("test artifact"),
			signature: []byte(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := v.Verify(ctx, tt.artifact, tt.signature, &VerifyOptions{})
			if err != nil {
				t.Fatalf("Verify() error = %v", err)
			}

			if result.Verified {
				t.Error("Expected verification to fail")
			}
		})
	}
}
