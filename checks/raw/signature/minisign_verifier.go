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
	"errors"

	"github.com/ossf/scorecard/v5/checker"
)

var errMinisignNotImplemented = errors.New("minisign verification not yet implemented")

// MinisignVerifier verifies Minisign signatures.
type MinisignVerifier struct{}

// CanVerify checks if this verifier can handle Minisign signatures.
func (m *MinisignVerifier) CanVerify(signatureType checker.SignatureType) bool {
	return signatureType == checker.SignatureTypeMinisign
}

// Verify verifies a Minisign signature against an artifact.
//
// NOTE: This is a placeholder implementation. Minisign verification requires
// public keys to verify signatures, but there is currently no standard way
// for Scorecard to discover where project maintainers publish their Minisign
// public keys.
//
// This will be much easier to implement once Scorecard supports maintainer
// annotations, which will allow projects to specify the location of their
// public keys (e.g., in a .scorecard/config.yml file or as repository metadata).
// With maintainer annotations, Scorecard can automatically discover and fetch
// the appropriate public keys for verification.
func (m *MinisignVerifier) Verify(
	ctx context.Context,
	artifact, signature []byte,
	opts *VerifyOptions,
) (*VerificationResult, error) {
	// TODO: Implement Minisign verification once maintainer annotations are available
	// This requires:
	// 1. Maintainer annotations to specify .minisign.pub public key location
	// 2. Using a Minisign library (e.g., github.com/jedisct1/go-minisign)
	// 3. Verifying the signature

	return &VerificationResult{
		Verified: false,
		Error:    errMinisignNotImplemented,
	}, nil
}
