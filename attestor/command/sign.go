// Copyright 2022 Security Scorecard Authors
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
	"io/ioutil"

	"github.com/golang/glog"
	"github.com/grafeas/kritis/pkg/attestlib"
	"github.com/grafeas/kritis/pkg/kritis/metadata/containeranalysis"
	"github.com/grafeas/kritis/pkg/kritis/signer"
	"github.com/grafeas/kritis/pkg/kritis/util"
)

func signCommand() {
	// Create a client
	client, err := containeranalysis.New()
	if err != nil {
		glog.Fatalf("Could not initialize the client %v", err)
	}

	// Read the signing credentials
	// Either kmsKeyName or pgpPriKeyPath needs to be set
	if kmsKeyName == "" && pgpPriKeyPath == "" && pkixPriKeyPath == "" {
		exitOnBadFlags(SignerMode(mode), "Neither kms_key_name, pgp_private_key, or pkix_private_key is specified")
	}
	var cSigner attestlib.Signer
	if kmsKeyName != "" {
		glog.Infof("Using kms key %s for signing.", kmsKeyName)
		if kmsDigestAlg == "" {
			glog.Fatalf("kms_digest_alg is unspecified, must be one of SHA256|SHA384|SHA512, and the same as specified by the key version's algorithm")
		}
		cSigner, err = signer.NewCloudKmsSigner(kmsKeyName, signer.DigestAlgorithm(kmsDigestAlg))
		if err != nil {
			glog.Fatalf("Creating kms signer failed: %v\n", err)
		}
	} else if pgpPriKeyPath != "" {
		glog.Infof("Using pgp key for signing.")
		signerKey, err := ioutil.ReadFile(pgpPriKeyPath)
		if err != nil {
			glog.Fatalf("Fail to read signer key: %v\n", err)
		}
		// Create a cryptolib signer
		cSigner, err = attestlib.NewPgpSigner(signerKey, pgpPassphrase)
		if err != nil {
			glog.Fatalf("Creating pgp signer failed: %v\n", err)
		}
	} else {
		glog.Infof("Using pkix key for signing.")
		signerKey, err := ioutil.ReadFile(pkixPriKeyPath)
		if err != nil {
			glog.Fatalf("Fail to read signer key: %v\n", err)
		}
		sAlg := attestlib.ParseSignatureAlgorithm(pkixAlg)
		if sAlg == attestlib.UnknownSigningAlgorithm {
			glog.Fatalf("Empty or unknown PKIX signature algorithm: %s\n", pkixAlg)
		}
		cSigner, err = attestlib.NewPkixSigner(signerKey, sAlg, "")
		if err != nil {
			glog.Fatalf("Creating pkix signer failed: %v\n", err)
		}
	}

	// Check note name
	err = util.CheckNoteName(noteName)
	if err != nil {
		exitOnBadFlags(SignerMode(mode), fmt.Sprintf("note name is invalid %v", err))
	}

	// Parse attestation project
	if attestationProject == "" {
		attestationProject = util.GetProjectFromContainerImage(image)
		glog.Infof("Using image project as attestation project: %s\n", attestationProject)
	} else {
		glog.Infof("Using specified attestation project: %s\n", attestationProject)
	}

	// Create signer
	r := signer.New(client, cSigner, noteName, attestationProject, overwrite)
	// Sign image
	err = r.SignImage(image)
	if err != nil {
		glog.Fatalf("Signing image failed: %v", err)
	}
}
