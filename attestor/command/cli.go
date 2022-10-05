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
	"fmt"
	"os"

	"github.com/spf13/cobra"
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
)

func addCheckFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&policyPath, "policy", "", "(required for check) scorecard attestation policy file path, e.g., /tmp/policy-binauthz.yml")
	cmd.MarkPersistentFlagRequired("policy")
	cmd.PersistentFlags().StringVar(&repoURL, "repo-url", "", "Repo URL from which source was built")
	cmd.MarkPersistentFlagRequired("repo-url")
	cmd.PersistentFlags().StringVar(&commitSHA, "commit", "", "Git SHA at which image was built")
}

func addSignFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&image, "image", "", "Image url, e.g., gcr.io/foo/bar@sha256:abcd")
	cmd.MarkPersistentFlagRequired("image")
	cmd.PersistentFlags().StringVar(&noteName, "note-name", "", "note name that created attestations are attached to, in the form of projects/[PROVIDER_ID]/notes/[NOTE_ID]")
	cmd.MarkPersistentFlagRequired("note-name")
	cmd.PersistentFlags().StringVar(&attestationProject, "attestation-project", "", "project id for GCP project that stores attestation, use image project if set to empty")
	cmd.PersistentFlags().BoolVar(&overwrite, "overwrite", false, "overwrite attestation if already existed")
	cmd.PersistentFlags().StringVar(&kmsKeyName, "kms-key-name", "", "kms key name, in the format of in the format projects/*/locations/*/keyRings/*/cryptoKeys/*/cryptoKeyVersions/*")
	cmd.PersistentFlags().StringVar(&kmsDigestAlg, "kms-digest-alg", "", "kms digest algorithm, must be one of SHA256|SHA384|SHA512, and the same as specified by the key version's algorithm")
	cmd.PersistentFlags().StringVar(&pgpPriKeyPath, "pgp-private-key", "", "pgp private signing key path, e.g., /dev/shm/key.pgp")
	cmd.PersistentFlags().StringVar(&pgpPassphrase, "pgp-passphrase", "", "passphrase for pgp private key, if any")
	cmd.PersistentFlags().StringVar(&pkixPriKeyPath, "pkix-private-key", "", "pkix private signing key path, e.g., /dev/shm/key.pem")
	cmd.PersistentFlags().StringVar(&pkixAlg, "pkix-alg", "", "pkix signature algorithm, e.g., ecdsa-p256-sha256")

}

// Export for testability
var RootCmd = &cobra.Command{
	Use:   "scorecard-attestor",
	Short: "scorecard-attestor generates attestations based on scorecard results",
}

var checkAndSignCmd = &cobra.Command{
	Use:   "attest",
	Short: "Run scorecard and sign a container image according to policy",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := runCheck(); err != nil {
			return err
		}
		return runSign()
	},
}

var checkCmd = &cobra.Command{
	Use:   "verify",
	Short: "Run scorecard and check an image against a policy",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCheck()
	},
}

func init() {
	RootCmd.AddCommand(checkCmd, checkAndSignCmd)

	addCheckFlags(checkAndSignCmd)
	addSignFlags(checkAndSignCmd)

	addCheckFlags(checkCmd)
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
