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
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	protobundle "github.com/sigstore/protobuf-specs/gen/pb-go/bundle/v1"
	protocommon "github.com/sigstore/protobuf-specs/gen/pb-go/common/v1"
	protodsse "github.com/sigstore/protobuf-specs/gen/pb-go/dsse"
	protorekor "github.com/sigstore/protobuf-specs/gen/pb-go/rekor/v1"
	"github.com/sigstore/sigstore-go/pkg/bundle"
	"github.com/sigstore/sigstore-go/pkg/root"
	"github.com/sigstore/sigstore-go/pkg/verify"

	"github.com/ossf/scorecard/v5/checker"
)

var (
	errUnsupportedFormat   = errors.New("unsupported format")
	errNoAttestationsFound = errors.New("no attestations found")
)

// decodeBase64 decodes a base64 string and returns descriptive error on failure.
func decodeBase64(s, fieldName string) ([]byte, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("failed to decode %s: %w", fieldName, err)
	}
	return b, nil
}

// parseInt64 parses a string to int64 and returns descriptive error on failure.
func parseInt64(s, fieldName string) (int64, error) {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s: %w", fieldName, err)
	}
	return v, nil
}

// SigstoreVerifier verifies Sigstore signatures.
type SigstoreVerifier struct {
	// testTrustedRoot is an optional pre-configured trusted root for testing.
	// If nil, createVerifier will fetch the trusted root from the network.
	testTrustedRoot *root.TrustedRoot
}

// CanVerify checks if this verifier can handle Sigstore signatures.
func (s *SigstoreVerifier) CanVerify(sigType checker.SignatureType) bool {
	return sigType == checker.SignatureTypeSigstore
}

// Verify verifies a Sigstore signature against an artifact.
func (s *SigstoreVerifier) Verify(
	ctx context.Context,
	artifact, signature []byte,
	opts *VerifyOptions,
) (*VerificationResult, error) {
	b, err := s.parseBundle(signature)
	if err != nil {
		return s.failedResult(err), nil
	}

	verifier, err := s.createVerifier()
	if err != nil {
		return s.failedResult(err), nil
	}

	// Build policy with artifact and permissive identity checking
	// For PyPI attestations, we verify the certificate chain against Fulcio CA
	// but don't require specific identity constraints
	policyOpts := []verify.PolicyOption{
		verify.WithoutIdentitiesUnsafe(),
	}

	policy := verify.NewPolicy(
		verify.WithArtifact(bytes.NewReader(artifact)),
		policyOpts...,
	)

	if err := s.verifyBundle(verifier, b, policy); err != nil {
		return s.failedResult(err), nil
	}

	return &VerificationResult{
		Verified: true,
		KeyID:    "", // Sigstore is keyless
	}, nil
}

// parseBundle parses a Sigstore bundle from JSON.
// Supports both standard Sigstore bundle format and PyPI PEP 740 attestation format.
func (s *SigstoreVerifier) parseBundle(
	signature []byte,
) (*bundle.Bundle, error) {
	// First try standard Sigstore bundle format
	var protoBundle protobundle.Bundle
	if err := json.Unmarshal(signature, &protoBundle); err == nil {
		b, err := bundle.NewBundle(&protoBundle)
		if err == nil {
			return b, nil
		}
	}

	// Try PyPI PEP 740 attestation format
	var pypiAttestation struct {
		AttestationBundles []struct {
			Attestations []pypiAttestationType `json:"attestations"`
		} `json:"attestation_bundles"`
		Version int `json:"version"`
	}

	if err := json.Unmarshal(signature, &pypiAttestation); err != nil {
		return nil, fmt.Errorf("failed to parse bundle: %w", errUnsupportedFormat)
	}

	if len(pypiAttestation.AttestationBundles) == 0 ||
		len(pypiAttestation.AttestationBundles[0].Attestations) == 0 {
		return nil, fmt.Errorf("failed to parse bundle: %w", errNoAttestationsFound)
	}

	// Convert PyPI attestation to Sigstore bundle format
	att := &pypiAttestation.AttestationBundles[0].Attestations[0]
	b, err := s.convertPyPIAttestationToBundle(att)
	if err != nil {
		return nil, fmt.Errorf("failed to convert PyPI attestation: %w", err)
	}

	return b, nil
}

// pypiEnvelope represents the DSSE envelope in PyPI attestations.
type pypiEnvelope struct {
	Signature string `json:"signature"`
	Statement string `json:"statement"`
}

// pypiCheckpoint represents a checkpoint in PyPI attestations.
type pypiCheckpoint struct {
	Envelope string `json:"envelope"`
}

// pypiInclusionProof represents an inclusion proof in PyPI attestations.
//
//nolint:govet // fieldalignment: JSON struct field order takes precedence over memory alignment
type pypiInclusionProof struct {
	Hashes     []string       `json:"hashes"`
	LogIndex   string         `json:"logIndex"`
	RootHash   string         `json:"rootHash"`
	TreeSize   string         `json:"treeSize"`
	Checkpoint pypiCheckpoint `json:"checkpoint"`
}

// pypiTransparencyEntry represents a transparency log entry in PyPI attestations.
type pypiTransparencyEntry struct {
	CanonicalizedBody string `json:"canonicalizedBody"`
	InclusionPromise  struct {
		SignedEntryTimestamp string `json:"signedEntryTimestamp"`
	} `json:"inclusionPromise"`
	InclusionProof pypiInclusionProof `json:"inclusionProof"`
	IntegratedTime string             `json:"integratedTime"`
	KindVersion    struct {
		Kind    string `json:"kind"`
		Version string `json:"version"`
	} `json:"kindVersion"`
	LogID struct {
		KeyID string `json:"keyId"`
	} `json:"logId"`
	LogIndex string `json:"logIndex"`
}

// pypiAttestationType represents a single PyPI PEP 740 attestation.
type pypiAttestationType struct {
	Envelope             pypiEnvelope `json:"envelope"`
	VerificationMaterial struct {
		Certificate         string                  `json:"certificate"`
		TransparencyEntries []pypiTransparencyEntry `json:"transparency_entries"`
	} `json:"verification_material"`
	Version int `json:"version"`
}

// buildTransparencyLogEntry creates a transparency log entry from PyPI attestation data.
func buildTransparencyLogEntry(
	entry *pypiTransparencyEntry,
) (*protorekor.TransparencyLogEntry, error) {
	integratedTime, err := parseInt64(entry.IntegratedTime, "integrated time")
	if err != nil {
		return nil, err
	}

	logIndex, err := parseInt64(entry.LogIndex, "log index")
	if err != nil {
		return nil, err
	}

	bodyBytes, err := decodeBase64(entry.CanonicalizedBody, "canonicalized body")
	if err != nil {
		return nil, err
	}

	setBytes, err := decodeBase64(entry.InclusionPromise.SignedEntryTimestamp, "SET")
	if err != nil {
		return nil, err
	}

	logIDBytes, err := decodeBase64(entry.LogID.KeyID, "log ID")
	if err != nil {
		return nil, err
	}

	tlogEntry := &protorekor.TransparencyLogEntry{
		LogIndex: logIndex,
		LogId: &protocommon.LogId{
			KeyId: logIDBytes,
		},
		KindVersion: &protorekor.KindVersion{
			Kind:    entry.KindVersion.Kind,
			Version: entry.KindVersion.Version,
		},
		IntegratedTime: integratedTime,
		InclusionPromise: &protorekor.InclusionPromise{
			SignedEntryTimestamp: setBytes,
		},
		CanonicalizedBody: bodyBytes,
	}

	if entry.InclusionProof.LogIndex != "" {
		proof, err := buildInclusionProof(&entry.InclusionProof)
		if err != nil {
			return nil, err
		}
		tlogEntry.InclusionProof = proof
	}

	return tlogEntry, nil
}

// buildInclusionProof creates an inclusion proof from PyPI attestation data.
func buildInclusionProof(
	proof *pypiInclusionProof,
) (*protorekor.InclusionProof, error) {
	proofLogIndex, err := parseInt64(proof.LogIndex, "proof log index")
	if err != nil {
		return nil, err
	}

	treeSize, err := parseInt64(proof.TreeSize, "tree size")
	if err != nil {
		return nil, err
	}

	hashes := make([][]byte, 0, len(proof.Hashes))
	for _, h := range proof.Hashes {
		hashBytes, err := decodeBase64(h, "hash")
		if err != nil {
			return nil, err
		}
		hashes = append(hashes, hashBytes)
	}

	rootHashBytes, err := decodeBase64(proof.RootHash, "root hash")
	if err != nil {
		return nil, err
	}

	return &protorekor.InclusionProof{
		LogIndex: proofLogIndex,
		TreeSize: treeSize,
		Hashes:   hashes,
		RootHash: rootHashBytes,
		Checkpoint: &protorekor.Checkpoint{
			Envelope: proof.Checkpoint.Envelope,
		},
	}, nil
}

// convertPyPIAttestationToBundle converts a PyPI PEP 740 attestation to a Sigstore bundle.
func (s *SigstoreVerifier) convertPyPIAttestationToBundle(att *pypiAttestationType) (*bundle.Bundle, error) {
	signatureBytes, err := decodeBase64(att.Envelope.Signature, "signature")
	if err != nil {
		return nil, err
	}

	statementBytes, err := decodeBase64(att.Envelope.Statement, "statement")
	if err != nil {
		return nil, err
	}

	certBytes, err := decodeBase64(att.VerificationMaterial.Certificate, "certificate")
	if err != nil {
		return nil, err
	}

	pb := &protobundle.Bundle{
		MediaType: "application/vnd.dev.sigstore.bundle.v0.3+json",
		VerificationMaterial: &protobundle.VerificationMaterial{
			Content: &protobundle.VerificationMaterial_Certificate{
				Certificate: &protocommon.X509Certificate{
					RawBytes: certBytes,
				},
			},
		},
		Content: &protobundle.Bundle_DsseEnvelope{
			DsseEnvelope: &protodsse.Envelope{
				Payload:     statementBytes,
				PayloadType: "application/vnd.in-toto+json",
				Signatures: []*protodsse.Signature{
					{
						Sig: signatureBytes,
					},
				},
			},
		},
	}

	if len(att.VerificationMaterial.TransparencyEntries) > 0 {
		tlogEntries := make([]*protorekor.TransparencyLogEntry, 0, len(att.VerificationMaterial.TransparencyEntries))
		for i := range att.VerificationMaterial.TransparencyEntries {
			entry, err := buildTransparencyLogEntry(&att.VerificationMaterial.TransparencyEntries[i])
			if err != nil {
				return nil, err
			}
			tlogEntries = append(tlogEntries, entry)
		}
		pb.VerificationMaterial.TlogEntries = tlogEntries
	}

	b, err := bundle.NewBundle(pb)
	if err != nil {
		return nil, fmt.Errorf("failed to create bundle: %w", err)
	}
	return b, nil
}

// createVerifier creates a Sigstore verifier with trusted root.
func (s *SigstoreVerifier) createVerifier() (*verify.Verifier, error) {
	var trustedRoot *root.TrustedRoot
	var err error

	// Use test trusted root if provided, otherwise fetch from network
	if s.testTrustedRoot != nil {
		trustedRoot = s.testTrustedRoot
	} else {
		trustedRoot, err = root.FetchTrustedRoot()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch trusted root: %w", err)
		}
	}

	verifierOpts := []verify.VerifierOption{
		verify.WithTransparencyLog(1),
		verify.WithObserverTimestamps(1),
	}

	verifier, err := verify.NewVerifier(trustedRoot, verifierOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create verifier: %w", err)
	}

	return verifier, nil
}

// verifyBundle verifies a bundle against a policy.
func (s *SigstoreVerifier) verifyBundle(
	verifier *verify.Verifier,
	b *bundle.Bundle,
	policy verify.PolicyBuilder,
) error {
	_, err := verifier.Verify(b, policy)
	if err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}
	return nil
}

// failedResult creates a failed verification result.
func (s *SigstoreVerifier) failedResult(err error) *VerificationResult {
	return &VerificationResult{
		Verified: false,
		Error:    err,
	}
}
