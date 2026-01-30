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

	"github.com/ossf/scorecard/v5/checker"
)

// SignatureVerifier defines the interface for cryptographic signature verification.
type SignatureVerifier interface {
	// CanVerify checks if this verifier can handle the given signature type.
	CanVerify(signatureType checker.SignatureType) bool

	// Verify performs cryptographic verification of a signature.
	Verify(
		ctx context.Context,
		artifact, signature []byte,
		opts *VerifyOptions,
	) (*VerificationResult, error)
}

// VerifyOptions contains options for signature verification.
type VerifyOptions struct {
	// KeyserverURL is the keyserver to query for GPG keys.
	// Default: keys.openpgp.org
	KeyserverURL string
}

// VerificationResult contains the result of signature verification.
type VerificationResult struct {
	// Error contains the verification error, if any.
	Error error

	// KeyID is the identifier of the key used (for GPG).
	KeyID string

	// Verified indicates whether the signature is valid.
	Verified bool
}

// GetVerifier returns the appropriate verifier for a signature type.
func GetVerifier(signatureType checker.SignatureType) SignatureVerifier {
	switch signatureType {
	case checker.SignatureTypeGPG:
		return &GPGVerifier{}
	case checker.SignatureTypeSigstore:
		return &SigstoreVerifier{}
	case checker.SignatureTypeMinisign:
		return &MinisignVerifier{}
	default:
		return nil
	}
}
