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
	"context"
	"testing"

	"github.com/ossf/scorecard/v5/checker"
)

func TestSigstoreVerifier_Verify_EmptyBundle(t *testing.T) {
	t.Parallel()

	v := &SigstoreVerifier{}
	ctx := context.Background()

	artifact := []byte("test artifact")
	bundle := []byte("")

	result, err := v.Verify(ctx, artifact, bundle, &VerifyOptions{})
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	if result.Verified {
		t.Error("Verification should fail with empty bundle")
	}

	if result.Error == nil {
		t.Error("Expected error with empty bundle")
	}
}

func TestSigstoreVerifier_Verify_InvalidJSON(t *testing.T) {
	t.Parallel()

	v := &SigstoreVerifier{}
	ctx := context.Background()

	artifact := []byte("test artifact")
	bundle := []byte("not valid json")

	result, err := v.Verify(ctx, artifact, bundle, &VerifyOptions{})
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	if result.Verified {
		t.Error("Verification should fail with invalid JSON")
	}

	if result.Error == nil {
		t.Error("Expected error with invalid JSON")
	}
}

func TestSigstoreVerifier_Verify_MalformedBundle(t *testing.T) {
	t.Parallel()

	v := &SigstoreVerifier{}
	ctx := context.Background()

	tests := []struct {
		name   string
		bundle string
	}{
		{
			name:   "empty JSON object",
			bundle: "{}",
		},
		{
			name:   "missing required fields",
			bundle: `{"mediaType": "application/vnd.dev.sigstore.bundle+json;version=0.2"}`,
		},
		{
			name:   "invalid structure",
			bundle: `{"foo": "bar"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			artifact := []byte("test artifact")
			result, err := v.Verify(ctx, artifact, []byte(tt.bundle), &VerifyOptions{})
			if err != nil {
				t.Fatalf("Verify() error = %v", err)
			}

			if result.Verified {
				t.Errorf("Verification should fail with %s", tt.name)
			}

			if result.Error == nil {
				t.Errorf("Expected error with %s", tt.name)
			}
		})
	}
}

func TestSigstoreVerifier_Verify_InvalidBundleStructure(t *testing.T) {
	t.Parallel()

	v := &SigstoreVerifier{}
	ctx := context.Background()

	tests := []struct {
		name   string
		bundle string
	}{
		{
			name: "bundle with invalid base64",
			bundle: `{
				"mediaType": "application/vnd.dev.sigstore.bundle+json;version=0.2",
				"verificationMaterial": {
					"content": "invalid-base64!@#$"
				}
			}`,
		},
		{
			name: "bundle with missing verification material",
			bundle: `{
				"mediaType": "application/vnd.dev.sigstore.bundle+json;version=0.2",
				"dsseEnvelope": {}
			}`,
		},
		{
			name: "bundle with invalid tlog entry",
			bundle: `{
				"mediaType": "application/vnd.dev.sigstore.bundle+json;version=0.2",
				"verificationMaterial": {
					"tlogEntries": [{"logIndex": "not-a-number"}]
				}
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			artifact := []byte("test artifact")
			result, err := v.Verify(ctx, artifact, []byte(tt.bundle), &VerifyOptions{})
			if err != nil {
				t.Fatalf("Verify() error = %v", err)
			}

			if result.Verified {
				t.Errorf("Verification should fail with %s", tt.name)
			}

			if result.Error == nil {
				t.Errorf("Expected error with %s", tt.name)
			}
		})
	}
}

func TestSigstoreVerifier_Verify_CreateVerifierError(t *testing.T) {
	t.Parallel()

	// This test covers the createVerifier() function by forcing Verify to call it
	// The function will fail when it tries to fetch the trusted root
	v := &SigstoreVerifier{}
	ctx := context.Background()

	// Create a minimal valid bundle structure that will pass parseBundle
	// but fail at createVerifier
	validBundle := `{
		"mediaType": "application/vnd.dev.sigstore.bundle+json;version=0.2",
		"verificationMaterial": {
			"content": "` + "dGVzdA==" + `",
			"tlogEntries": [
				{
					"logIndex": "1",
					"logId": {
						"keyId": "dGVzdA=="
					},
					"kindVersion": {
						"kind": "hashedrekord",
						"version": "0.0.1"
					},
					"integratedTime": "1234567890",
					"inclusionPromise": {
						"signedEntryTimestamp": "dGVzdA=="
					},
					"inclusionProof": {
						"logIndex": "1",
						"rootHash": "dGVzdA==",
						"treeSize": "100",
						"hashes": ["dGVzdA=="],
						"checkpoint": {
							"envelope": "dGVzdA=="
						}
					},
					"canonicalizedBody": "dGVzdA=="
				}
			]
		},
		"dsseEnvelope": {
			"payload": "dGVzdA==",
			"payloadType": "application/vnd.in-toto+json",
			"signatures": [
				{
					"sig": "dGVzdA==",
					"keyid": ""
				}
			]
		}
	}`

	artifact := []byte("test artifact")
	result, err := v.Verify(ctx, artifact, []byte(validBundle), &VerifyOptions{})
	// We expect this to fail because createVerifier will fail
	// (either network error or verification failure)
	if err != nil {
		t.Fatalf("Verify() should not return error, got: %v", err)
	}

	if result.Verified {
		t.Error("Verification should fail when verifier creation fails")
	}

	if result.Error == nil {
		t.Error("Expected error in result when verification fails")
	}
}

func TestSigstoreVerifier_CanVerify_Extended(t *testing.T) {
	t.Parallel()

	v := &SigstoreVerifier{}

	tests := []struct {
		name     string
		sigType  checker.SignatureType
		expected bool
	}{
		{
			name:     "sigstore type",
			sigType:  checker.SignatureTypeSigstore,
			expected: true,
		},
		{
			name:     "gpg type",
			sigType:  checker.SignatureTypeGPG,
			expected: false,
		},
		{
			name:     "minisign type",
			sigType:  checker.SignatureTypeMinisign,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := v.CanVerify(tt.sigType)
			if result != tt.expected {
				t.Errorf("CanVerify() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
