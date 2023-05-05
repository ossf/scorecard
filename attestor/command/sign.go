// Copyright 2022 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/grafeas/kritis/pkg/attestlib"
	"github.com/grafeas/kritis/pkg/kritis/metadata/containeranalysis"
	"github.com/grafeas/kritis/pkg/kritis/signer"
	"github.com/grafeas/kritis/pkg/kritis/util"

	sclog "github.com/ossf/scorecard/v4/log"
)

const scorecardNoteID = "ossf-scorecard-attestation"

type EncryptionParamError struct {
	Message string
}

func (ep EncryptionParamError) Error() string {
	return ep.Message
}

func runSign() error {
	logger := sclog.NewLogger(sclog.DefaultLevel)

	// Create a client
	client, err := containeranalysis.New()
	if err != nil {
		return fmt.Errorf("could not initialize the client %w", err)
	}

	// Read the signing credentials
	// Either kmsKeyName or pgpPriKeyPath needs to be set
	if kmsKeyName == "" && pgpPriKeyPath == "" && pkixPriKeyPath == "" {
		return EncryptionParamError{"neither kms_key_name, pgp_private_key, or pkix_private_key is specified"}
	}
	var cSigner attestlib.Signer
	switch {
	case kmsKeyName != "":
		logger.Info(fmt.Sprintf("Using kms key %s for signing.", kmsKeyName))
		if kmsDigestAlg == "" {
			//nolint:lll
			return EncryptionParamError{"kms_digest_alg is unspecified, must be one of SHA256|SHA384|SHA512, and the same as specified by the key version's algorithm"}
		}
		kmsDigestAlg = strings.ToUpper(kmsDigestAlg)
		cSigner, err = signer.NewCloudKmsSigner(kmsKeyName, signer.DigestAlgorithm(kmsDigestAlg))
		if err != nil {
			return fmt.Errorf("creating kms signer failed: %w", err)
		}
	case pgpPriKeyPath != "":
		logger.Info("Using pgp key for signing.")
		signerKey, err := os.ReadFile(pgpPriKeyPath)
		if err != nil {
			return fmt.Errorf("fail to read signer key: %w", err)
		}
		// Create a cryptolib signer
		cSigner, err = attestlib.NewPgpSigner(signerKey, pgpPassphrase)
		if err != nil {
			return fmt.Errorf("creating pgp signer failed: %w", err)
		}
	default:
		logger.Info("Using pkix key for signing.")
		signerKey, err := os.ReadFile(pkixPriKeyPath)
		if err != nil {
			return fmt.Errorf("fail to read signer key: %w", err)
		}
		sAlg := attestlib.ParseSignatureAlgorithm(pkixAlg)
		if sAlg == attestlib.UnknownSigningAlgorithm {
			return EncryptionParamError{fmt.Sprintf("empty or unknown PKIX signature algorithm: %s", pkixAlg)}
		}
		cSigner, err = attestlib.NewPkixSigner(signerKey, sAlg, "")
		if err != nil {
			return fmt.Errorf("creating pkix signer failed: %w", err)
		}
	}

	// Parse attestation project
	if attestationProject == "" {
		attestationProject = util.GetProjectFromContainerImage(image)
		logger.Info(fmt.Sprintf("Using image project as attestation project: %s", attestationProject))
	} else {
		logger.Info(fmt.Sprintf("Using specified attestation project: %s", attestationProject))
	}

	// Check note name
	scorecardNoteName := fmt.Sprintf("projects/%s/notes/%s", attestationProject, scorecardNoteID)

	err = util.CheckNoteName(scorecardNoteName)
	if err != nil {
		return fmt.Errorf("note name is invalid %w", err)
	}

	// Create signer
	r := signer.New(client, cSigner, scorecardNoteName, attestationProject, overwrite)
	// Sign image
	err = r.SignImage(image)
	if err != nil {
		return fmt.Errorf("signing image failed: %w", err)
	}
	return nil
}
