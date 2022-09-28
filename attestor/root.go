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

package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/golang/glog"
	"github.com/grafeas/kritis/pkg/attestlib"
	"github.com/grafeas/kritis/pkg/kritis/metadata/containeranalysis"
	"github.com/grafeas/kritis/pkg/kritis/signer"
	"github.com/grafeas/kritis/pkg/kritis/util"
	"github.com/ossf/scorecard-attestor/policy"
	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks"
	sclog "github.com/ossf/scorecard/v4/log"
	"github.com/ossf/scorecard/v4/pkg"
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
	fs.StringVar(&repoURL, "repoURL", "", "(required) repo URL from which source was built")
	fs.StringVar(&commitSHA, "commit", "", "git SHA at which image was built")
}

func addCheckFlags(fs *flag.FlagSet) {
	fs.StringVar(&policyPath, "policy", "", "(required for check) scorecard attestation policy file path, e.g., /tmp/policy-binauthz.yml")
}

func addSignFlags(fs *flag.FlagSet) {
	fs.StringVar(&noteName, "note_name", "", "(required for sign) note name that created attestations are attached to, in the form of projects/[PROVIDER_ID]/notes/[NOTE_ID]")
	fs.StringVar(&attestationProject, "attestation_project", "", "project id for GCP project that stores attestation, use image project if set to empty")
	fs.BoolVar(&overwrite, "overwrite", false, "overwrite attestation if already existed")
	fs.StringVar(&kmsKeyName, "kms_key_name", "", "kms key name, in the format of in the format projects/*/locations/*/keyRings/*/cryptoKeys/*/cryptoKeyVersions/*")
	fs.StringVar(&kmsDigestAlg, "kms_digest_alg", "", "kms digest algorithm, must be one of SHA256|SHA384|SHA512, and the same as specified by the key version's algorithm")
	fs.StringVar(&pgpPriKeyPath, "pgp_private_key", "", "pgp private signing key path, e.g., /dev/shm/key.pgp")
	fs.StringVar(&pgpPassphrase, "pgp_passphrase", "", "passphrase for pgp private key, if any")
	fs.StringVar(&pkixPriKeyPath, "pkix_private_key", "", "pkix private signing key path, e.g., /dev/shm/key.pem")
	fs.StringVar(&pkixAlg, "pkix_alg", "", "pkix signature algorithm, e.g., ecdsa-p256-sha256")
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
	-note_name=projects/$NOTE_PROJECT/NOTES/$NOTE_ID \
	-kms_key_name=projects/$KMS_PROJECT/locations/$KMS_KEYLOCATION/keyRings/$KMS_KEYRING/cryptoKeys/$KMS_KEYNAME/cryptoKeyVersions/$KMS_KEYVERSION \
	-kms_digest_alg=SHA512`
	case BypassAndSign:
		addSignFlags(modeFlags)
		doSign = true
		modeExample = `	./scorecard-attestor \
	-mode=bypass-and-sign \
	-image=gcr.io/my-image-repo/image-1@sha256:123 \
	-note_name=projects/$NOTE_PROJECT/NOTES/$NOTE_ID \
	-kms_key_name=projects/$KMS_PROJECT/locations/$KMS_KEYLOCATION/keyRings/$KMS_KEYRING/cryptoKeys/$KMS_KEYNAME/cryptoKeyVersions/$KMS_KEYVERSION \
	-kms_digest_alg=SHA512`
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

func main() {
	flag.Parse()
	glog.Infof("Signer mode: %s.", mode)
	ctx := context.Background()

	doCheck, doSign, err := parseSignerMode(SignerMode(mode))
	if err != nil {
		glog.Fatalf("Parse mode err %v.", err)
	}

	// Check image url is non-empty
	if image == "" {
		exitOnBadFlags(SignerMode(mode), "image url is empty")
	}

	// Create a client
	client, err := containeranalysis.New()
	if err != nil {
		glog.Fatalf("Could not initialize the client %v", err)
	}

	if doCheck {
		// Read the Binauthz attestation policy
		if policyPath == "" {
			exitOnBadFlags(SignerMode(mode), "policy path is empty")
		}
		attestationPolicy, err := policy.ParseAttestationPolicyFromFile(policyPath)
		if err != nil {
			glog.Fatalf("Fail to load scorecard attestation policy: %v", err)
		}

		if repoURL == "" {
			buildRepo := os.Getenv("REPO_NAME")
			if buildRepo == "" {
				exitOnBadFlags(SignerMode(mode), "repoURL not specified")
			}
			repoURL = buildRepo
			glog.Infof("Found repo URL %s Cloud Build environment", repoURL)
		} else {
			glog.Infof("Running scorecard on %s", repoURL)
		}

		if commitSHA == "" {
			buildSHA := os.Getenv("COMMIT_SHA")
			if buildSHA == "" {
				glog.Infof("commit not specified, running on HEAD")
			} else {
				commitSHA = buildSHA
				glog.Infof("Found revision %s Cloud Build environment", commitSHA)
			}
		}

		logger := sclog.NewLogger(sclog.DefaultLevel)

		repo, repoClient, ossFuzzRepoClient, ciiClient, vulnsClient, err := checker.GetClients(
			ctx, repoURL, "", logger)

		enabledChecks := map[string]checker.Check{
			"BinaryArtifacts": {
				Fn: checks.BinaryArtifacts,
				SupportedRequestTypes: []checker.RequestType{
					checker.CommitBased,
				},
			},
		}

		repoResult, err := pkg.RunScorecards(
			ctx,
			repo,
			commitSHA,
			enabledChecks,
			repoClient,
			ossFuzzRepoClient,
			ciiClient,
			vulnsClient,
		)
		if err != nil {
			glog.Fatalf("RunScorecards: %w", err)
		}

		result, err := policy.RunChecksForPolicy(attestationPolicy, &repoResult.RawResults)
		if err != nil {
			glog.Fatalf("Error when evaluating image %q against policy", image)
		}
		if result != policy.Pass {
			glog.Errorf("policy check failed on image %s:", image)
			os.Exit(1)
		}
		glog.Infof("Image %q passes policy check", image)
	}

	if doSign {
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
		err := r.SignImage(image)
		if err != nil {
			glog.Fatalf("Signing image failed: %v", err)
		}
	}
}
