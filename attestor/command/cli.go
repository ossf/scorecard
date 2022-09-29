// Copyright 2022 Security Scorecard Authors
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

// Package pkg defines fns for running Scorecard checks on a Repo.

package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/golang/glog"
)

type SignerMode string

const (
	CheckAndSign  SignerMode = "check-and-sign"
	CheckOnly     SignerMode = "check-only"
	BypassAndSign SignerMode = "bypass-and-sign"
)

var (
	// input flags
	repoURL            string
	commitSHA          string
	mode               string
	image              string
	policyPath         string
	attestationProject string
	overwrite          bool
	noteName           string
	// input flags: pgp key flags
	pgpPriKeyPath string
	pgpPassphrase string
	// pkix key flags
	pkixPriKeyPath string
	pkixAlg        string

	// input flags: kms flags
	kmsKeyName   string
	kmsDigestAlg string

	// helper global variables
	modeFlags   *flag.FlagSet
	modeExample string
)

func init() {
	// need to add all flags to avoid "flag not provided error"
	addBasicFlags(flag.CommandLine)
	addCheckFlags(flag.CommandLine)
	addSignFlags(flag.CommandLine)
}

func addBasicFlags(fs *flag.FlagSet) {
	fs.StringVar(&mode, "mode", "check-and-sign", "(required) mode of operation, check-and-sign|check-only|bypass-and-sign")
	fs.StringVar(&image, "image", "", "(required) image url, e.g., gcr.io/foo/bar@sha256:abcd")
	fs.StringVar(&repoURL, "repo-url", "", "(required) repo URL from which source was built")
	fs.StringVar(&commitSHA, "commit", "", "git SHA at which image was built")
}

func addCheckFlags(fs *flag.FlagSet) {
	fs.StringVar(&policyPath, "policy", "", "(required for check) scorecard attestation policy file path, e.g., /tmp/policy-binauthz.yml")
}

func addSignFlags(fs *flag.FlagSet) {
	fs.StringVar(&noteName, "note-name", "", "(required for sign) note name that created attestations are attached to, in the form of projects/[PROVIDER_ID]/notes/[NOTE_ID]")
	fs.StringVar(&attestationProject, "attestation-project", "", "project id for GCP project that stores attestation, use image project if set to empty")
	fs.BoolVar(&overwrite, "overwrite", false, "overwrite attestation if already existed")
	fs.StringVar(&kmsKeyName, "kms-key-name", "", "kms key name, in the format of in the format projects/*/locations/*/keyRings/*/cryptoKeys/*/cryptoKeyVersions/*")
	fs.StringVar(&kmsDigestAlg, "kms-digest-alg", "", "kms digest algorithm, must be one of SHA256|SHA384|SHA512, and the same as specified by the key version's algorithm")
	fs.StringVar(&pgpPriKeyPath, "pgp-private-key", "", "pgp private signing key path, e.g., /dev/shm/key.pgp")
	fs.StringVar(&pgpPassphrase, "pgp-passphrase", "", "passphrase for pgp private key, if any")
	fs.StringVar(&pkixPriKeyPath, "pkix-private_key", "", "pkix private signing key path, e.g., /dev/shm/key.pem")
	fs.StringVar(&pkixAlg, "pkix-alg", "", "pkix signature algorithm, e.g., ecdsa-p256-sha256")
}

// parseSignerMode creates mode-specific flagset and analyze actions (check, sign) for given mode
func parseSignerMode(mode SignerMode) (doCheck bool, doSign bool, err error) {
	modeFlags, doCheck, doSign, err = flag.NewFlagSet("", flag.ExitOnError), false, false, nil
	addBasicFlags(modeFlags)
	switch mode {
	case CheckAndSign:
		addCheckFlags(modeFlags)
		addSignFlags(modeFlags)
		doCheck, doSign = true, true
		modeExample = `	./scorecard-attestor \
	-mode=check-and-sign \
	-image=gcr.io/my-image-repo/image-1@sha256:123 \
	-policy=policy.yaml \
	-note-name=projects/$NOTE_PROJECT/NOTES/$NOTE_ID \
	-kms-key-name=projects/$KMS_PROJECT/locations/$KMS_KEYLOCATION/keyRings/$KMS_KEYRING/cryptoKeys/$KMS_KEYNAME/cryptoKeyVersions/$KMS_KEYVERSION \
	-kms-digest-alg=SHA512`
	case BypassAndSign:
		addSignFlags(modeFlags)
		doSign = true
		modeExample = `	./scorecard-attestor \
	-mode=bypass-and-sign \
	-image=gcr.io/my-image-repo/image-1@sha256:123 \
	-note-name=projects/$NOTE_PROJECT/NOTES/$NOTE_ID \
	-kms-key-name=projects/$KMS_PROJECT/locations/$KMS_KEYLOCATION/keyRings/$KMS_KEYRING/cryptoKeys/$KMS_KEYNAME/cryptoKeyVersions/$KMS_KEYVERSION \
	-kms-digest-alg=SHA512`
	case CheckOnly:
		addCheckFlags(modeFlags)
		doCheck = true
		modeExample = `	./scorecard-attestor \
	-mode=check-only \
	-image=gcr.io/my-image-repo/image-1@sha256:123 \
	-policy=policy.yaml`
	default:
		return false, false, fmt.Errorf("unrecognized mode %s, must be one of check-and-sign|check-only|bypass-and-sign", mode)
	}
	flag.Parse()
	return doCheck, doSign, err
}

func exitOnBadFlags(mode SignerMode, err string) {
	fmt.Fprintf(modeFlags.Output(), "Usage of signer's %s mode:\n", mode)
	modeFlags.PrintDefaults()
	fmt.Fprintf(modeFlags.Output(), "Example (%s mode):\n %s\n", mode, modeExample)
	fmt.Fprintf(modeFlags.Output(), "Bad flags for mode %s: %v. \n", mode, err)
	os.Exit(1)
}

func Command() {
	flag.Parse()
	glog.Infof("Signer mode: %s.", mode)

	doCheck, doSign, err := parseSignerMode(SignerMode(mode))
	if err != nil {
		glog.Fatalf("Parse mode err %v.", err)
	}

	// Check image url is non-empty
	if image == "" {
		exitOnBadFlags(SignerMode(mode), "image url is empty")
	}

	if doCheck {
		checkCommand()
	}

	if doSign {
		signCommand()
	}
}
