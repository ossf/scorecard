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
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/sigstore/sigstore-go/pkg/bundle"
)

// Helper to create mock Rekor DSSE body.
func createMockRekorBody() string {
	mockSigB64 := base64.StdEncoding.EncodeToString([]byte("mock_sig"))
	mockVerifierB64 := base64.StdEncoding.EncodeToString([]byte("mock_verifier"))
	dsseBody := map[string]interface{}{
		"apiVersion": "0.0.1",
		"kind":       "dsse",
		"spec": map[string]interface{}{
			"payloadHash": map[string]interface{}{
				"algorithm": "sha256",
				"value":     "abc123def456",
			},
			"envelopeHash": map[string]interface{}{
				"algorithm": "sha256",
				"value":     "xyz789uvw012",
			},
			"signatures": []interface{}{
				map[string]interface{}{
					"signature": mockSigB64,
					"verifier":  mockVerifierB64,
				},
			},
		},
	}
	dsseBodyJSON, err := json.Marshal(dsseBody)
	if err != nil {
		panic("failed to marshal mock DSSE body: " + err.Error())
	}
	return base64.StdEncoding.EncodeToString(dsseBodyJSON)
}

// Helper to create a minimal pypiAttestationType for testing.
func createMinimalAttestation(signature, statement, certificate, body, set, logID string) *pypiAttestationType {
	return &pypiAttestationType{
		Envelope: pypiEnvelope{
			Signature: signature,
			Statement: statement,
		},
		VerificationMaterial: struct {
			Certificate         string                  `json:"certificate"`
			TransparencyEntries []pypiTransparencyEntry `json:"transparency_entries"`
		}{
			Certificate: certificate,
			TransparencyEntries: []pypiTransparencyEntry{
				{
					CanonicalizedBody: body,
					InclusionPromise: struct {
						SignedEntryTimestamp string `json:"signedEntryTimestamp"`
					}{
						SignedEntryTimestamp: set,
					},
					IntegratedTime: "1729868390",
					KindVersion: struct {
						Kind    string `json:"kind"`
						Version string `json:"version"`
					}{
						Kind:    "dsse",
						Version: "0.0.1",
					},
					LogID: struct {
						KeyID string `json:"keyId"`
					}{
						KeyID: logID,
					},
					LogIndex: "14365388",
				},
			},
		},
		Version: 1,
	}
}

// Helper to add inclusion proof to an attestation.
func addInclusionProof(att *pypiAttestationType, checkpoint string, hashes []string, rootHash string) {
	if len(att.VerificationMaterial.TransparencyEntries) > 0 {
		att.VerificationMaterial.TransparencyEntries[0].InclusionProof = pypiInclusionProof{
			Hashes: hashes,
			Checkpoint: pypiCheckpoint{
				Envelope: checkpoint,
			},
			LogIndex: "14365388",
			RootHash: rootHash,
			TreeSize: "21749622",
		}
	}
}

// TestConvertPyPIAttestationToBundle tests the PEP 740 to Sigstore bundle conversion.
// verifyBundleStructure checks basic bundle structure.
func verifyBundleStructure(t *testing.T, b *bundle.Bundle) {
	t.Helper()
	if b == nil {
		t.Error("bundle is nil")
		return
	}
	if b.GetMediaType() != "application/vnd.dev.sigstore.bundle.v0.3+json" {
		t.Errorf("bundle media type = %v, want %v", b.GetMediaType(), "application/vnd.dev.sigstore.bundle.v0.3+json")
	}
	if b.GetVerificationMaterial() == nil {
		t.Error("bundle verification material is nil")
	}
	if b.GetDsseEnvelope() == nil {
		t.Error("bundle DSSE envelope is nil")
	}
}

func TestConvertPyPIAttestationToBundle(t *testing.T) {
	t.Parallel()

	sampleSignature := base64.StdEncoding.EncodeToString([]byte("mock_signature_data"))
	sampleStatement := base64.StdEncoding.EncodeToString([]byte("mock_statement_data"))
	sampleCertificate := base64.StdEncoding.EncodeToString([]byte("mock_certificate_data"))
	sampleBody := createMockRekorBody()
	sampleSET := base64.StdEncoding.EncodeToString([]byte("mock_set_data"))
	sampleLogID := base64.StdEncoding.EncodeToString([]byte("mock_log_id"))
	sampleHash := base64.StdEncoding.EncodeToString([]byte("mock_hash"))
	sampleRootHash := base64.StdEncoding.EncodeToString([]byte("mock_root_hash"))

	tests := []struct {
		name                 string
		attestation          *pypiAttestationType
		errContains          string
		skipBundleValidation bool
		wantErr              bool
	}{
		{
			name:                 "valid attestation with minimal fields",
			attestation:          createMinimalAttestation(sampleSignature, sampleStatement, sampleCertificate, sampleBody, sampleSET, sampleLogID),
			wantErr:              false,
			skipBundleValidation: true,
		},
		{
			name: "valid attestation with inclusion proof",
			attestation: func() *pypiAttestationType {
				att := createMinimalAttestation(sampleSignature, sampleStatement, sampleCertificate, sampleBody, sampleSET, sampleLogID)
				addInclusionProof(att, "mock_checkpoint", []string{sampleHash}, sampleRootHash)
				return att
			}(),
			wantErr:              false,
			skipBundleValidation: true,
		},
		{
			name:        "invalid base64 signature",
			attestation: createMinimalAttestation("invalid!!!base64", sampleStatement, sampleCertificate, sampleBody, sampleSET, sampleLogID),
			wantErr:     true,
			errContains: "failed to decode signature",
		},
		{
			name:        "invalid base64 statement",
			attestation: createMinimalAttestation(sampleSignature, "invalid!!!base64", sampleCertificate, sampleBody, sampleSET, sampleLogID),
			wantErr:     true,
			errContains: "failed to decode statement",
		},
		{
			name:        "invalid base64 certificate",
			attestation: createMinimalAttestation(sampleSignature, sampleStatement, "invalid!!!base64", sampleBody, sampleSET, sampleLogID),
			wantErr:     true,
			errContains: "failed to decode certificate",
		},
		{
			name: "invalid integrated time",
			attestation: func() *pypiAttestationType {
				att := createMinimalAttestation(sampleSignature, sampleStatement, sampleCertificate, sampleBody, sampleSET, sampleLogID)
				att.VerificationMaterial.TransparencyEntries[0].IntegratedTime = "not_a_number"
				return att
			}(),
			wantErr:     true,
			errContains: "failed to parse integrated time",
		},
		{
			name: "invalid log index",
			attestation: func() *pypiAttestationType {
				att := createMinimalAttestation(sampleSignature, sampleStatement, sampleCertificate, sampleBody, sampleSET, sampleLogID)
				att.VerificationMaterial.TransparencyEntries[0].LogIndex = "not_a_number"
				return att
			}(),
			wantErr:     true,
			errContains: "failed to parse log index",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			verifier := &SigstoreVerifier{}
			b, err := verifier.convertPyPIAttestationToBundle(tt.attestation)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				if tt.skipBundleValidation && (strings.Contains(err.Error(), "validation error") ||
					strings.Contains(err.Error(), "inclusion proof missing")) {
					return
				}
				t.Errorf("unexpected error = %v", err)
				return
			}

			verifyBundleStructure(t, b)
		})
	}
}

// TestParseBundlePyPIFormat tests parsing of PyPI PEP 740 format.
func TestParseBundlePyPIFormat(t *testing.T) {
	t.Parallel()

	// Sample base64-encoded data
	sampleSignature := base64.StdEncoding.EncodeToString([]byte("mock_signature"))
	sampleStatement := base64.StdEncoding.EncodeToString([]byte("mock_statement"))
	sampleCertificate := base64.StdEncoding.EncodeToString([]byte("mock_certificate"))
	sampleBody := createMockRekorBody()
	sampleSET := base64.StdEncoding.EncodeToString([]byte("mock_set"))
	sampleLogID := base64.StdEncoding.EncodeToString([]byte("mock_logid"))

	pypiJSON := map[string]interface{}{
		"version": 1,
		"attestation_bundles": []map[string]interface{}{
			{
				"attestations": []map[string]interface{}{
					{
						"version": 1,
						"envelope": map[string]interface{}{
							"signature": sampleSignature,
							"statement": sampleStatement,
						},
						"verification_material": map[string]interface{}{
							"certificate": sampleCertificate,
							"transparency_entries": []map[string]interface{}{
								{
									"canonicalizedBody": sampleBody,
									"inclusionPromise": map[string]interface{}{
										"signedEntryTimestamp": sampleSET,
									},
									"integratedTime": "1729868390",
									"kindVersion": map[string]interface{}{
										"kind":    "dsse",
										"version": "0.0.1",
									},
									"logId": map[string]interface{}{
										"keyId": sampleLogID,
									},
									"logIndex": "14365388",
								},
							},
						},
					},
				},
			},
		},
	}

	jsonBytes, err := json.Marshal(pypiJSON)
	if err != nil {
		t.Fatalf("failed to marshal test JSON: %v", err)
	}

	verifier := &SigstoreVerifier{}
	b, err := verifier.parseBundle(jsonBytes)
	if err != nil {
		// For mock data, we expect validation errors but want to test parsing succeeds
		if strings.Contains(err.Error(), "validation error") || strings.Contains(err.Error(), "inclusion proof missing") {
			// Conversion succeeded, validation failed (expected for mock data)
			return
		}
		t.Errorf("parseBundle() error = %v, want nil or validation error", err)
		return
	}

	if b == nil {
		t.Error("parseBundle() returned nil bundle")
		return
	}

	// Verify the bundle was created correctly
	if b.GetMediaType() != "application/vnd.dev.sigstore.bundle.v0.3+json" {
		t.Errorf("bundle media type = %v, want %v", b.GetMediaType(), "application/vnd.dev.sigstore.bundle.v0.3+json")
	}
}

// TestParseBundlePyPIFormatErrors tests error handling for malformed PyPI attestations.
func TestParseBundlePyPIFormatErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		jsonData    string
		errContains string
	}{
		{
			name:        "empty attestation bundles",
			jsonData:    `{"version": 1, "attestation_bundles": []}`,
			errContains: "no attestations found",
		},
		{
			name:        "empty attestations array",
			jsonData:    `{"version": 1, "attestation_bundles": [{"attestations": []}]}`,
			errContains: "no attestations found",
		},
		{
			name:        "completely invalid JSON",
			jsonData:    `{invalid json`,
			errContains: "unsupported format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			verifier := &SigstoreVerifier{}
			_, err := verifier.parseBundle([]byte(tt.jsonData))

			if err == nil {
				t.Error("parseBundle() expected error but got none")
				return
			}

			if !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("parseBundle() error = %v, want error containing %q", err, tt.errContains)
			}
		})
	}
}
