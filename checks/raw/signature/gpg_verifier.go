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
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/packet"

	"github.com/ossf/scorecard/v5/checker"
)

var (
	errKeyServerStatusNotOK = errors.New("keyserver returned non-OK status")
	errNoKeysFound          = errors.New("no keys found")
	errNoKeyIDFound         = errors.New("no key ID found in signature")
)

const (
	defaultKeyserver = "https://keys.openpgp.org"
	keyserverPath    = "/vks/v1/by-keyid/"
)

// GPGVerifier verifies GPG/PGP signatures.
type GPGVerifier struct {
	client *http.Client
}

// CanVerify checks if this verifier can handle GPG signatures.
func (g *GPGVerifier) CanVerify(sigType checker.SignatureType) bool {
	return sigType == checker.SignatureTypeGPG
}

// Verify verifies a GPG signature against an artifact.
func (g *GPGVerifier) Verify(
	ctx context.Context,
	artifact, signature []byte,
	opts *VerifyOptions,
) (*VerificationResult, error) {
	g.ensureClient()
	keyserver := g.getKeyserver(opts)

	// Extract key ID and fetch from keyserver
	return g.verifyWithKeyExtraction(ctx, artifact, signature, keyserver)
}

// ensureClient initializes HTTP client if needed.
func (g *GPGVerifier) ensureClient() {
	if g.client == nil {
		g.client = &http.Client{}
	}
}

// getKeyserver returns configured keyserver or default.
func (g *GPGVerifier) getKeyserver(opts *VerifyOptions) string {
	if opts.KeyserverURL != "" {
		return opts.KeyserverURL
	}
	return defaultKeyserver
}

// verifyWithKeyExtraction extracts key ID from signature and fetches key.
func (g *GPGVerifier) verifyWithKeyExtraction(
	ctx context.Context,
	artifact, signature []byte,
	keyserver string,
) (*VerificationResult, error) {
	keyID, err := g.extractKeyIDFromSignature(signature)
	if err != nil {
		return g.failedResult(
			"",
			fmt.Errorf("failed to extract key ID: %w", err),
		), nil
	}

	keyURL := fmt.Sprintf("%s%s%s", keyserver, keyserverPath, keyID)
	keyring, _, err := g.fetchPublicKeyFromURL(ctx, keyURL)
	if err != nil {
		return g.failedResult(
			keyID,
			fmt.Errorf("failed to fetch key %s: %w", keyID, err),
		), nil
	}

	return g.checkSignature(artifact, signature, keyring, keyID)
}

// checkSignature performs the actual signature verification.
func (g *GPGVerifier) checkSignature(
	artifact, signature []byte,
	keyring openpgp.EntityList,
	keyID string,
) (*VerificationResult, error) {
	signer, err := openpgp.CheckArmoredDetachedSignature(
		keyring,
		bytes.NewReader(artifact),
		bytes.NewReader(signature),
		nil, // config
	)
	if err != nil {
		return g.failedResult(
			keyID,
			fmt.Errorf("signature verification failed: %w", err),
		), nil
	}

	verifiedKeyID := fmt.Sprintf("%X", signer.PrimaryKey.KeyId)
	return &VerificationResult{
		Verified: true,
		KeyID:    verifiedKeyID,
	}, nil
}

// failedResult creates a failed verification result.
func (g *GPGVerifier) failedResult(
	keyID string,
	err error,
) *VerificationResult {
	return &VerificationResult{
		Verified: false,
		KeyID:    keyID,
		Error:    err,
	}
}

// fetchPublicKeyFromURL retrieves a public key from a URL.
func (g *GPGVerifier) fetchPublicKeyFromURL(
	ctx context.Context,
	url string,
) (openpgp.EntityList, string, error) {
	body, err := g.downloadKey(ctx, url)
	if err != nil {
		return nil, "", err
	}

	return g.parseKey(body)
}

// downloadKey fetches key data from URL.
func (g *GPGVerifier) downloadKey(
	ctx context.Context,
	url string,
) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch key: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", errKeyServerStatusNotOK, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read key: %w", err)
	}

	return body, nil
}

// parseKey parses armored PGP key data.
func (g *GPGVerifier) parseKey(
	data []byte,
) (openpgp.EntityList, string, error) {
	keyring, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(data))
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse key: %w", err)
	}

	if len(keyring) == 0 {
		return nil, "", errNoKeysFound
	}

	keyID := fmt.Sprintf("%X", keyring[0].PrimaryKey.KeyId)
	return keyring, keyID, nil
}

// extractKeyIDFromSignature extracts the signer's key ID from signature.
func (g *GPGVerifier) extractKeyIDFromSignature(
	signature []byte,
) (string, error) {
	block, err := armor.Decode(bytes.NewReader(signature))
	if err != nil {
		return "", fmt.Errorf("failed to decode signature: %w", err)
	}

	return g.findKeyIDInPackets(block.Body)
}

// findKeyIDInPackets searches for key ID in PGP packets.
func (g *GPGVerifier) findKeyIDInPackets(body io.Reader) (string, error) {
	packets := packet.NewReader(body)

	for {
		p, err := packets.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read packet: %w", err)
		}

		if keyID := g.extractKeyIDFromPacket(p); keyID != "" {
			return keyID, nil
		}
	}

	return "", errNoKeyIDFound
}

// extractKeyIDFromPacket extracts key ID from a single packet.
func (g *GPGVerifier) extractKeyIDFromPacket(p packet.Packet) string {
	if sig, ok := p.(*packet.Signature); ok {
		// V4+ signature
		if sig.IssuerKeyId != nil {
			return fmt.Sprintf("%X", *sig.IssuerKeyId)
		}
	}
	return ""
}
