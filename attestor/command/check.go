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
	"context"
	"fmt"
	"os"

	"github.com/ossf/scorecard/v5/attestor/policy"
	"github.com/ossf/scorecard/v5/checker"
	sclog "github.com/ossf/scorecard/v5/log"
	"github.com/ossf/scorecard/v5/pkg/scorecard"
)

type EmptyParameterError struct {
	Param string
}

func (ep EmptyParameterError) Error() string {
	return fmt.Sprintf("param %s is empty", ep.Param)
}

func runCheck() (policy.PolicyResult, error) {
	return RunCheckWithParams(repoURL, commitSHA, policyPath)
}

// RunCheckWithParams: Run scorecard check on repo. Export for testability.
func RunCheckWithParams(repoURL, commitSHA, policyPath string) (policy.PolicyResult, error) {
	ctx := context.Background()
	logger := sclog.NewLogger(sclog.DefaultLevel)

	// Read the Binauthz attestation policy
	if policyPath == "" {
		return policy.Fail, EmptyParameterError{Param: "policy"}
	}

	var attestationPolicy *policy.AttestationPolicy

	attestationPolicy, err := policy.ParseAttestationPolicyFromFile(policyPath)
	if err != nil {
		return policy.Fail, fmt.Errorf("fail to load scorecard attestation policy: %w", err)
	}

	if repoURL == "" {
		buildRepo := os.Getenv("REPO_NAME")
		if buildRepo == "" {
			return policy.Fail, EmptyParameterError{Param: "repoURL"}
		}
		repoURL = buildRepo
		logger.Info(fmt.Sprintf("Found repo URL %s Cloud Build environment", repoURL))
	} else {
		logger.Info(fmt.Sprintf("Running scorecard on %s", repoURL))
	}

	if commitSHA == "" {
		buildSHA := os.Getenv("COMMIT_SHA")
		if buildSHA == "" {
			logger.Info("commit not specified, running on HEAD")
			commitSHA = "HEAD"
		} else {
			commitSHA = buildSHA
			logger.Info(fmt.Sprintf("Found revision %s from GCB build environment", commitSHA))
		}
	}

	repo, repoClient, ossFuzzRepoClient, ciiClient, vulnsClient, _, err := checker.GetClients(
		ctx, repoURL, "", logger)
	if err != nil {
		return policy.Fail, fmt.Errorf("couldn't set up clients: %w", err)
	}

	requiredChecks := attestationPolicy.GetRequiredChecksForPolicy()

	var enabledChecks []string
	for check, required := range requiredChecks {
		if required {
			enabledChecks = append(enabledChecks, check)
		}
	}

	repoResult, err := scorecard.Run(ctx, repo,
		scorecard.WithCommitSHA(commitSHA),
		scorecard.WithChecks(enabledChecks),
		scorecard.WithRepoClient(repoClient),
		scorecard.WithOSSFuzzClient(ossFuzzRepoClient),
		scorecard.WithOpenSSFBestPraticesClient(ciiClient),
		scorecard.WithVulnerabilitiesClient(vulnsClient),
	)
	if err != nil {
		return policy.Fail, fmt.Errorf("scorecard.Run: %w", err)
	}

	result, err := attestationPolicy.EvaluateResults(&repoResult.RawResults)
	if err != nil {
		return policy.Fail, fmt.Errorf("error when evaluating image %q against policy: %w", image, err)
	}
	if result != policy.Pass {
		logger.Info("image failed scorecard attestation policy check")
	} else {
		logger.Info("image passed scorecard attestation policy check")
	}
	return result, nil
}
