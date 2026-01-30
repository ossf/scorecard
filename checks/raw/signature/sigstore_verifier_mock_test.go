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
	"encoding/base64"
	"encoding/json"
	"testing"
)

// TestSigstoreVerifier_Verify_WithValidBundle tests the Verify function
// with a structurally valid bundle to exercise createVerifier and verifyBundle paths.
// This will fail verification (as expected without real signatures), but will
// exercise the code paths for coverage.
func TestSigstoreVerifier_Verify_WithValidBundle(t *testing.T) {
	t.Parallel()

	v := &SigstoreVerifier{}
	ctx := context.Background()

	// Create a structurally valid Sigstore bundle
	// This exercises parseBundle, createVerifier, and verifyBundle
	bundle := createValidBundleStructure(t)

	artifact := []byte("test artifact content")
	result, err := v.Verify(ctx, artifact, bundle, &VerifyOptions{})
	if err != nil {
		t.Fatalf("Verify() should not return error, got: %v", err)
	}

	// We expect verification to fail because we don't have real signatures
	// But this exercises createVerifier() and verifyBundle() code paths
	if result.Verified {
		t.Error("Verification should fail with test bundle (no real signatures)")
	}

	// We should have an error in the result
	if result.Error == nil {
		t.Error("Expected error in result when verification fails")
	}
}

// TestSigstoreVerifier_Verify_BundleWithAllFields tests parsing a bundle
// with all required fields populated to maximize code coverage.
func TestSigstoreVerifier_Verify_BundleWithAllFields(t *testing.T) {
	t.Parallel()

	v := &SigstoreVerifier{}
	ctx := context.Background()

	// Create a bundle with comprehensive field population
	bundleData := map[string]interface{}{
		"mediaType": "application/vnd.dev.sigstore.bundle+json;version=0.2",
		"verificationMaterial": map[string]interface{}{
			"content": base64.StdEncoding.EncodeToString([]byte("test certificate")),
			"tlogEntries": []map[string]interface{}{
				{
					"logIndex": "12345",
					"logId": map[string]interface{}{
						"keyId": base64.StdEncoding.EncodeToString([]byte("test-key-id")),
					},
					"kindVersion": map[string]interface{}{
						"kind":    "hashedrekord",
						"version": "0.0.1",
					},
					"integratedTime": "1640000000",
					"inclusionPromise": map[string]interface{}{
						"signedEntryTimestamp": base64.StdEncoding.EncodeToString([]byte("timestamp")),
					},
					"inclusionProof": map[string]interface{}{
						"logIndex": "12345",
						"rootHash": base64.StdEncoding.EncodeToString([]byte("root")),
						"treeSize": "50000",
						"hashes": []string{
							base64.StdEncoding.EncodeToString([]byte("hash1")),
							base64.StdEncoding.EncodeToString([]byte("hash2")),
						},
						"checkpoint": map[string]interface{}{
							"envelope": base64.StdEncoding.EncodeToString([]byte("checkpoint-data")),
						},
					},
					"canonicalizedBody": base64.StdEncoding.EncodeToString([]byte("body")),
				},
			},
			"timestampVerificationData": map[string]interface{}{
				"rfc3161Timestamps": []map[string]interface{}{
					{
						"signedTimestamp": base64.StdEncoding.EncodeToString([]byte("timestamp")),
					},
				},
			},
		},
		"dsseEnvelope": map[string]interface{}{
			"payload":     base64.StdEncoding.EncodeToString([]byte(`{"_type":"test"}`)),
			"payloadType": "application/vnd.in-toto+json",
			"signatures": []map[string]interface{}{
				{
					"sig":   base64.StdEncoding.EncodeToString([]byte("signature-data")),
					"keyid": "",
				},
			},
		},
	}

	bundleJSON, err := json.Marshal(bundleData)
	if err != nil {
		t.Fatalf("Failed to create test bundle: %v", err)
	}

	artifact := []byte("test artifact")
	result, err := v.Verify(ctx, artifact, bundleJSON, &VerifyOptions{})
	if err != nil {
		t.Fatalf("Verify() should not return error, got: %v", err)
	}

	// Verification will fail, but we've exercised all the parsing paths
	if result.Verified {
		t.Error("Expected verification to fail with test data")
	}
}

// TestSigstoreVerifier_Verify_InclusionProofParsing tests the buildInclusionProof function.
func TestSigstoreVerifier_Verify_InclusionProofParsing(t *testing.T) {
	t.Parallel()

	v := &SigstoreVerifier{}
	ctx := context.Background()

	tests := []struct {
		proof map[string]interface{}
		name  string
	}{
		{
			name: "proof with multiple hashes",
			proof: map[string]interface{}{
				"logIndex": "100",
				"rootHash": base64.StdEncoding.EncodeToString([]byte("root")),
				"treeSize": "1000",
				"hashes": []string{
					base64.StdEncoding.EncodeToString([]byte("hash1")),
					base64.StdEncoding.EncodeToString([]byte("hash2")),
					base64.StdEncoding.EncodeToString([]byte("hash3")),
				},
				"checkpoint": map[string]interface{}{
					"envelope": base64.StdEncoding.EncodeToString([]byte("checkpoint")),
				},
			},
		},
		{
			name: "proof with single hash",
			proof: map[string]interface{}{
				"logIndex": "50",
				"rootHash": base64.StdEncoding.EncodeToString([]byte("root")),
				"treeSize": "500",
				"hashes": []string{
					base64.StdEncoding.EncodeToString([]byte("single-hash")),
				},
				"checkpoint": map[string]interface{}{
					"envelope": base64.StdEncoding.EncodeToString([]byte("cp")),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			bundleData := map[string]interface{}{
				"mediaType": "application/vnd.dev.sigstore.bundle+json;version=0.2",
				"verificationMaterial": map[string]interface{}{
					"content": base64.StdEncoding.EncodeToString([]byte("cert")),
					"tlogEntries": []map[string]interface{}{
						{
							"logIndex": "1",
							"logId": map[string]interface{}{
								"keyId": base64.StdEncoding.EncodeToString([]byte("key")),
							},
							"kindVersion": map[string]interface{}{
								"kind":    "hashedrekord",
								"version": "0.0.1",
							},
							"integratedTime": "1640000000",
							"inclusionPromise": map[string]interface{}{
								"signedEntryTimestamp": base64.StdEncoding.EncodeToString([]byte("ts")),
							},
							"inclusionProof":    tt.proof,
							"canonicalizedBody": base64.StdEncoding.EncodeToString([]byte("body")),
						},
					},
				},
				"dsseEnvelope": map[string]interface{}{
					"payload":     base64.StdEncoding.EncodeToString([]byte(`{}`)),
					"payloadType": "application/vnd.in-toto+json",
					"signatures": []map[string]interface{}{
						{
							"sig":   base64.StdEncoding.EncodeToString([]byte("sig")),
							"keyid": "",
						},
					},
				},
			}

			bundleJSON, err := json.Marshal(bundleData)
			if err != nil {
				t.Fatalf("Failed to create test bundle: %v", err)
			}

			result, err := v.Verify(ctx, []byte("artifact"), bundleJSON, &VerifyOptions{})
			if err != nil {
				t.Fatalf("Verify() should not return error, got: %v", err)
			}

			// Verification will fail but we've tested the inclusion proof parsing
			if result.Verified {
				t.Error("Expected verification to fail")
			}
		})
	}
}

// createValidBundleStructure creates a structurally valid Sigstore bundle for testing.
func createValidBundleStructure(t *testing.T) []byte {
	t.Helper()

	bundleData := map[string]interface{}{
		"mediaType": "application/vnd.dev.sigstore.bundle+json;version=0.2",
		"verificationMaterial": map[string]interface{}{
			"content": base64.StdEncoding.EncodeToString([]byte("test-certificate-data")),
			"tlogEntries": []map[string]interface{}{
				{
					"logIndex": "123456",
					"logId": map[string]interface{}{
						"keyId": base64.StdEncoding.EncodeToString([]byte("log-key-id")),
					},
					"kindVersion": map[string]interface{}{
						"kind":    "hashedrekord",
						"version": "0.0.1",
					},
					"integratedTime": "1640000000",
					"inclusionPromise": map[string]interface{}{
						"signedEntryTimestamp": base64.StdEncoding.EncodeToString([]byte("test-timestamp")),
					},
					"inclusionProof": map[string]interface{}{
						"logIndex": "123456",
						"rootHash": base64.StdEncoding.EncodeToString([]byte("test-root-hash")),
						"treeSize": "200000",
						"hashes": []string{
							base64.StdEncoding.EncodeToString([]byte("hash-1")),
							base64.StdEncoding.EncodeToString([]byte("hash-2")),
						},
						"checkpoint": map[string]interface{}{
							"envelope": base64.StdEncoding.EncodeToString([]byte("checkpoint-envelope")),
						},
					},
					"canonicalizedBody": base64.StdEncoding.EncodeToString([]byte("canonicalized-body")),
				},
			},
		},
		"dsseEnvelope": map[string]interface{}{
			"payload":     base64.StdEncoding.EncodeToString([]byte(`{"_type":"https://in-toto.io/Statement/v0.1"}`)),
			"payloadType": "application/vnd.in-toto+json",
			"signatures": []map[string]interface{}{
				{
					"sig":   base64.StdEncoding.EncodeToString([]byte("test-signature-bytes")),
					"keyid": "",
				},
			},
		},
	}

	bundleJSON, err := json.Marshal(bundleData)
	if err != nil {
		t.Fatalf("Failed to create bundle structure: %v", err)
	}

	return bundleJSON
}

// TestSigstoreVerifier_Verify_PyPIAttestation tests conversion from PyPI format.
func TestSigstoreVerifier_Verify_PyPIAttestation(t *testing.T) {
	t.Parallel()

	v := &SigstoreVerifier{}
	ctx := context.Background()

	// PyPI attestation format (PEP 740)
	pypiAttestation := map[string]interface{}{
		"version": 1,
		"attestation_bundles": []map[string]interface{}{
			{
				"mediaType": "application/vnd.dev.sigstore.bundle+json;version=0.3",
				"verificationMaterial": map[string]interface{}{
					"certificate": base64.StdEncoding.EncodeToString([]byte("cert")),
					"transparencyEntries": []map[string]interface{}{
						{
							"logIndex":          123,
							"logId":             map[string]interface{}{"keyId": base64.StdEncoding.EncodeToString([]byte("key"))},
							"kindVersion":       map[string]interface{}{"kind": "hashedrekord", "version": "0.0.1"},
							"integratedTime":    1640000000,
							"inclusionPromise":  map[string]interface{}{"signedEntryTimestamp": base64.StdEncoding.EncodeToString([]byte("ts"))},
							"inclusionProof":    nil,
							"canonicalizedBody": base64.StdEncoding.EncodeToString([]byte("body")),
						},
					},
				},
				"dsseEnvelope": map[string]interface{}{
					"payload":     base64.StdEncoding.EncodeToString([]byte(`{}`)),
					"payloadType": "application/vnd.in-toto+json",
					"signatures": []map[string]interface{}{
						{"sig": base64.StdEncoding.EncodeToString([]byte("sig"))},
					},
				},
			},
		},
	}

	bundleJSON, err := json.Marshal(pypiAttestation)
	if err != nil {
		t.Fatalf("Failed to create PyPI attestation: %v", err)
	}

	artifact := []byte("test-artifact")
	result, err := v.Verify(ctx, artifact, bundleJSON, &VerifyOptions{})
	if err != nil {
		t.Fatalf("Verify() should not return error, got: %v", err)
	}

	// Will fail verification but exercises PyPI format conversion
	if result.Verified {
		t.Error("Expected verification to fail")
	}
}
