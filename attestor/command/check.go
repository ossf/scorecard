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
	"context"
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/ossf/scorecard-attestor/policy"
	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks"
	sclog "github.com/ossf/scorecard/v4/log"
	"github.com/ossf/scorecard/v4/pkg"
)

func runCheck() error {
	ctx := context.Background()

	// Read the Binauthz attestation policy
	if policyPath == "" {
		return fmt.Errorf("policy path is empty")
	}
	attestationPolicy, err := policy.ParseAttestationPolicyFromFile(policyPath)
	if err != nil {
		return fmt.Errorf("fail to load scorecard attestation policy: %v", err)
	}

	if repoURL == "" {
		buildRepo := os.Getenv("REPO_NAME")
		if buildRepo == "" {
			return fmt.Errorf("repoURL not specified")
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

	requiredChecks := policy.GetRequiredChecksForPolicy(attestationPolicy)

	enabledChecks := map[string]checker.Check{
		"BinaryArtifacts": {
			Fn: checks.BinaryArtifacts,
			SupportedRequestTypes: []checker.RequestType{
				checker.CommitBased,
			},
		},
	}

	// Filter out checks that won't be needed for policy-evaluation time
	for name := range enabledChecks {
		if _, isRequired := requiredChecks[name]; !isRequired {
			delete(enabledChecks, name)
		}
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
		return fmt.Errorf("RunScorecards: %w", err)
	}

	result, err := policy.EvaluateResults(attestationPolicy, &repoResult.RawResults)
	if err != nil {
		return fmt.Errorf("error when evaluating image %q against policy", image)
	}
	if result != policy.Pass {
		return fmt.Errorf("image failed policy check %s:", image)
	}
	glog.Infof("Policy check passed")
	return nil
}
